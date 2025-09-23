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
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/go-archive"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const BRANCH_FOR_DEPLOY = "refs/heads/main"

var postReceiveCmd = &cobra.Command{
	Use: "post-receive",
	Run: func(cmd *cobra.Command, args []string) {
		gitDir, err := os.Getwd()
		if err != nil {
			log.Fatal().Err(err).Msg("getwd")
		}
		reponame := filepath.Base(gitDir)

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
			return
		}

		log.Info().Str("old_sha", oldsha).Str("new_sha", newsha).Str("ref", refName).Msg("starting deployment...")

		tmpDir, err := os.MkdirTemp("", fmt.Sprintf("deployment-%s-*", newsha))
		if err != nil {
			log.Error().Err(err).Msg("create temp dir")
			return
		}
		defer os.RemoveAll(tmpDir)

		if err := exec.Command("git", "clone", "--depth=1", "--branch=main", gitDir, tmpDir).Run(); err != nil {
			log.Error().Err(err).Msg("git clone")
			return
		}

		if _, err := os.Stat(filepath.Join(tmpDir, "Dockerfile")); os.IsNotExist(err) {
			log.Warn().Msg("no Dockerfile found, skipping deployment")
			return
		}

		log.Info().Msg("starting deployment with docker")

		if err := deployWithDocker(tmpDir, reponame, newsha); err != nil {
			log.Error().Err(err).Msg("failed to deploy with docker")
		}
	},
}

func deployWithDocker(dir, name, sha string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	buildContext, err := archive.TarWithOptions(dir, &archive.TarOptions{})
	if err != nil {
		return err
	}
	buildOptions := build.ImageBuildOptions{
		Tags:       []string{fmt.Sprintf("%s:%s", name, sha), fmt.Sprintf("%s:latest", name)},
		Dockerfile: "Dockerfile",
		Remove:     true,
		NoCache:    true,
	}
	resp, err := cli.ImageBuild(context.Background(), buildContext, buildOptions)
	if err != nil {
		fmt.Println("ImageBuild error:", err)
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
			return err
		}
		if stream := strings.TrimSpace(jm.Stream); stream != "" {
			log.Info().Msg(stream)
		}
		if jm.Aux != nil {
			var result build.Result
			if err := json.Unmarshal(*jm.Aux, &result); err != nil {
				return err
			}
			imageID = result.ID
		}
	}
	if imageID == "" {
		return fmt.Errorf("failed to get image ID")
	}

	log.Info().Str("image", imageID).Msg("built image successfully")

	return nil
}
