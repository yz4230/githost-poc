package gitopt

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
)

func TestBuildServiceAnnouncement(t *testing.T) {
	b, err := BuildServiceAnnouncement(ServiceUploadPack)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := string(b)
	expectedPrefix := "001e# service=git-upload-pack\n0000"
	if got != expectedPrefix {
		t.Fatalf("unexpected announcement.\n got: %q\nwant: %q", got, expectedPrefix)
	}
}

func TestEnsureBareRepoCtx(t *testing.T) {
	dir := t.TempDir()
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	ctx := logger.WithContext(context.Background())
	path, err := EnsureBareRepo(ctx, dir, "user", "repo")
	if err != nil {
		t.Fatalf("EnsureBareRepo error: %v", err)
	}
	// bare repo should have HEAD file
	if _, err := os.Stat(filepath.Join(path, "HEAD")); err != nil {
		t.Fatalf("expected HEAD in repo: %v", err)
	}
	if _, err := EnsureBareRepo(ctx, dir, "user", "repo"); err != nil {
		t.Fatalf("second EnsureBareRepo error: %v", err)
	}
}
