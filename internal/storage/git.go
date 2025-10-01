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
)

type GitStorage interface {
	GetRepoDir(name string) string
	IsRepoExist(name string) bool
	EnsureBareRepo(ctx context.Context, name string) error
	InitBareRepo(ctx context.Context, name string) error
	RemoveRepo(ctx context.Context, name string) error
}

type gitStorageImpl struct {
	rootDir string
	log     zerolog.Logger
}

func (g *gitStorageImpl) GetRepoDir(reponame string) string {
	return lo.Must(filepath.Abs(filepath.Join(g.rootDir, reponame+".git")))
}

func (g *gitStorageImpl) IsRepoExist(reponame string) bool {
	repodir := g.GetRepoDir(reponame)
	_, err := os.Stat(repodir)
	return !os.IsNotExist(err)
}

// EnsureBareRepo implements GitStorage.
func (g *gitStorageImpl) EnsureBareRepo(ctx context.Context, reponame string) error {
	repodir := g.GetRepoDir(reponame)
	if !g.IsRepoExist(reponame) {
		g.log.Debug().Str("dir", repodir).Msg("repo does not exist, initializing")
		if err := g.InitBareRepo(ctx, reponame); err != nil {
			return err
		}
	}
	return nil
}

// InitBareRepo implements GitStorage.
func (g *gitStorageImpl) InitBareRepo(ctx context.Context, reponame string) error {
	repodir := g.GetRepoDir(reponame)
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

// RemoveRepo implements GitStorage.
func (g *gitStorageImpl) RemoveRepo(ctx context.Context, reponame string) error {
	repodir := g.GetRepoDir(reponame)
	if !g.IsRepoExist(reponame) {
		return nil
	}
	if err := os.RemoveAll(repodir); err != nil {
		return fmt.Errorf("remove repo dir: %w", err)
	}
	return nil
}

func shellScript(lines ...string) string {
	return "#!/bin/sh\n" + strings.Join(lines, "\n") + "\n"
}

func NewGitStorage(root string, log zerolog.Logger) GitStorage {
	return &gitStorageImpl{
		rootDir: root,
		log:     log,
	}
}
