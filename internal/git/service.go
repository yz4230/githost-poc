package git

import (
	"context"
	"fmt"

	"os"
	"os/exec"
	"path/filepath"

	"strings"

	"github.com/rs/zerolog"
)

// EnsureBareRepo ensures the bare repository exists (initializing if needed) and returns its absolute path.
func EnsureBareRepo(ctx context.Context, root, reponame string) (string, error) {
	log := zerolog.Ctx(ctx)
	repodir, err := filepath.Abs(filepath.Join(root, Repositories, ensureSuffix(reponame, ".git")))
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(repodir); os.IsNotExist(err) {
		log.Debug().Str("dir", repodir).Msg("repo does not exist, initializing")
		if err := InitBareRepo(ctx, repodir); err != nil {
			return "", err
		}
	}
	return repodir, nil
}

func InitBareRepo(ctx context.Context, repodir string) error {
	log := zerolog.Ctx(ctx)
	if err := os.MkdirAll(repodir, os.ModePerm); err != nil {
		return fmt.Errorf("create repo dir: %w", err)
	}
	if err := exec.Command("git", "init", "--bare", repodir).Run(); err != nil {
		return fmt.Errorf("init bare repo: %w", err)
	}
	if err := createGitHooks(ctx, repodir); err != nil {
		return fmt.Errorf("create git hooks: %w", err)
	}

	log.Info().Str("dir", repodir).Msg("initialized bare git repository")

	return nil
}

func ensureSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		return s
	}
	return s + suffix
}

func createGitHooks(ctx context.Context, repodir string) error {
	log := zerolog.Ctx(ctx)
	hooksDir := filepath.Join(repodir, "hooks")
	if err := os.MkdirAll(hooksDir, os.ModePerm); err != nil {
		return fmt.Errorf("create hooks dir: %w", err)
	}

	scriptPath := filepath.Join(hooksDir, "post-receive")
	scriptContent := fmt.Sprintf(`#!/bin/sh
echo $(cat) | %s hook post-receive
`, os.Args[0])
	if err := os.WriteFile(scriptPath, []byte(scriptContent), os.ModePerm); err != nil {
		return fmt.Errorf("write post-receive hook: %w", err)
	}

	log.Info().Str("dir", hooksDir).Msg("created git hooks")

	return nil
}
