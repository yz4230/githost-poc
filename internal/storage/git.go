package storage

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/yz4230/githost-poc/internal/utils"
)

type GitStorage interface {
	InitBare(ctx context.Context, name string) error
}

type GitStorageImpl struct {
	rootDir string
	log     zerolog.Logger
}

// InitBare implements GitStorage.
func (g *GitStorageImpl) InitBare(ctx context.Context, reponame string) error {
	repodir := g.getRepoDir(reponame)
	if err := os.MkdirAll(repodir, os.ModePerm); err != nil {
		return fmt.Errorf("create repo dir: %w", err)
	}
	if err := exec.Command("git", "init", "--bare", repodir).Run(); err != nil {
		return fmt.Errorf("init bare repo: %w", err)
	}

	hooksDir := filepath.Join(repodir, "hooks")
	if err := os.MkdirAll(hooksDir, os.ModePerm); err != nil {
		return fmt.Errorf("create hooks dir: %w", err)
	}

	scriptPath := filepath.Join(hooksDir, "post-receive")
	scriptContent := shellScript(fmt.Sprintf("echo $(cat) | %s hook post-receive", os.Args[0]))
	if err := os.WriteFile(scriptPath, []byte(scriptContent), os.ModePerm); err != nil {
		return fmt.Errorf("write post-receive hook: %w", err)
	}

	return nil
}

func (g *GitStorageImpl) getRepoDir(reponame string) string {
	return lo.Must(filepath.Abs(filepath.Join(g.rootDir, utils.EnsureSuffix(reponame, ".git"))))
}

func shellScript(lines ...string) string {
	return "#!/bin/sh\n" + strings.Join(lines, "\n") + "\n"
}

func NewGitStorage(root string, log zerolog.Logger) GitStorage {
	return &GitStorageImpl{
		rootDir: root,
		log:     log,
	}
}
