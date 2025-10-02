package hook

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/go-archive"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const BRANCH_FOR_DEPLOY = "refs/heads/main"

var postReceiveCmd = &cobra.Command{
	Use:           "post-receive",
	Short:         "Handle post-receive git hook. Not intended to be run manually.",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		gitDir, err := os.Getwd()
		if err != nil {
			log.Fatal().Err(err).Msg("getwd")
		}
		reponame := strings.TrimSuffix(filepath.Base(gitDir), ".git")

		shouldDeploy := false
		s := bufio.NewScanner(os.Stdin)
		var oldsha, newsha, refName string
		for s.Scan() {
			line := s.Text()
			parts := strings.Fields(line)
			if len(parts) != 3 {
				log.Error().Str("line", line).Msg("invalid input line")
				continue
			}
			oldsha, newsha, refName = parts[0], parts[1], parts[2]

			if refName == BRANCH_FOR_DEPLOY {
				shouldDeploy = true
				break
			}
		}

		if err := s.Err(); err != nil {
			log.Fatal().Err(err).Msg("read stdin")
		}

		if !shouldDeploy {
			log.Info().Msg("no deployment needed")
			return nil
		}

		log.Info().Str("old_sha", oldsha).Str("new_sha", newsha).Str("ref", refName).Msg("starting deployment...")

		tmpDir, err := os.MkdirTemp("", fmt.Sprintf("deployment-%s-*", newsha))
		if err != nil {
			log.Error().Err(err).Msg("create temp dir")
			return err
		}
		defer os.RemoveAll(tmpDir)

		if err := exec.Command("git", "clone", "--depth=1", "--branch=main", gitDir, tmpDir).Run(); err != nil {
			log.Error().Err(err).Msg("git clone")
			return err
		}

		if _, err := os.Stat(filepath.Join(tmpDir, "Dockerfile")); os.IsNotExist(err) {
			log.Warn().Msg("no Dockerfile found, skipping deployment")
			return nil
		}

		log.Info().Msg("starting deployment with docker")

		if err := deployWithDocker(tmpDir, reponame, newsha); err != nil {
			log.Error().Err(err).Msg("failed to deploy with docker")
			return err
		}

		return nil
	},
}

func deployWithDocker(repodir, reponame, commitSHA string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer cli.Close()

	imageID, err := buildDockerImage(cli, repodir, reponame, commitSHA)
	if err != nil {
		return fmt.Errorf("failed to build docker image: %w", err)
	}

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", "githost.enabled=true"),
			filters.Arg("label", fmt.Sprintf("githost.repo=%s", reponame)),
		),
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to list containers")
		return err
	}
	if len(containers) > 0 {
		for _, c := range containers {
			log.Info().Str("container", c.ID).Msg("removing existing container")
			if err := cli.ContainerStop(context.Background(), c.ID, container.StopOptions{}); err != nil {
				log.Error().Err(err).Msg("failed to stop existing container")
				return err
			}
			if err := cli.ContainerRemove(context.Background(), c.ID, container.RemoveOptions{}); err != nil {
				log.Error().Err(err).Msg("failed to remove existing container")
				return err
			}
			log.Info().Str("container", c.ID).Msg("removed existing container")
		}
	}

	log.Info().Str("image", imageID).Msg("starting new container")

	containerName := fmt.Sprintf("%s-%s", reponame, commitSHA[:7])
	resp, err := cli.ContainerCreate(context.Background(),
		&container.Config{
			Image: fmt.Sprintf("%s:%s", reponame, commitSHA),
			Labels: map[string]string{
				"githost.enabled": "true",
				"githost.repo":    reponame,
				"githost.commit":  commitSHA,
			},
		},
		&container.HostConfig{
			RestartPolicy: container.RestartPolicy{
				Name: container.RestartPolicyUnlessStopped,
			},
		}, nil, nil, containerName)

	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	if err := cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	log.Info().Str("container", resp.ID).Msg("started new container")

	return nil
}

func buildDockerImage(cli *client.Client, repodir, reponame, commitSHA string) (string, error) {
	buildContext, err := archive.TarWithOptions(repodir, &archive.TarOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create tar archive: %w", err)
	}
	buildOptions := build.ImageBuildOptions{
		Tags: []string{fmt.Sprintf("%s:%s", reponame, commitSHA), fmt.Sprintf("%s:latest", reponame)},
		Labels: map[string]string{
			"githost.enabled": "true",
			"githost.repo":    reponame,
			"githost.commit":  commitSHA,
		},
		Dockerfile: "Dockerfile",
		Remove:     true,
		NoCache:    true,
	}
	resp, err := cli.ImageBuild(context.Background(), buildContext, buildOptions)
	if err != nil {
		return "", fmt.Errorf("failed to build image: %w", err)
	}
	defer resp.Body.Close()

	imageID := ""
	dec := json.NewDecoder(resp.Body)
	for {
		var jm jsonmessage.JSONMessage
		if err := dec.Decode(&jm); err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("failed to decode json message: %w", err)
		}
		if stream := strings.TrimSpace(jm.Stream); stream != "" {
			log.Info().Msg(stream)
		}
		if jm.Aux != nil {
			var result build.Result
			if err := json.Unmarshal(*jm.Aux, &result); err != nil {
				return "", fmt.Errorf("failed to unmarshal json message: %w", err)
			}
			imageID = result.ID
		}
	}
	if imageID == "" {
		return "", fmt.Errorf("failed to get image ID")
	}

	log.Info().Str("image", imageID).Msg("built image successfully")

	return imageID, nil
}
