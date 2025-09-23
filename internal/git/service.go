package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
)

// Service represents a supported git smart protocol service.
type Service string

const (
	ServiceUploadPack  Service = "git-upload-pack"
	ServiceReceivePack Service = "git-receive-pack"
)

var validService = map[Service]bool{
	ServiceUploadPack:  true,
	ServiceReceivePack: true,
}

var nameRe = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

// SanitizeName validates a single path segment (username / reponame).
func SanitizeName(s string) (string, error) {
	if s == "" {
		return "", errors.New("empty name")
	}
	if strings.Contains(s, "..") || strings.ContainsRune(s, '/') || !nameRe.MatchString(s) {
		return "", fmt.Errorf("invalid name: %s", s)
	}
	return s, nil
}

// EnsureBareRepo ensures the bare repository exists (initializing if needed) and returns its absolute path.
func EnsureBareRepo(ctx context.Context, root, user, repo string) (string, error) {
	log := zerolog.Ctx(ctx)
	user, err := SanitizeName(user)
	if err != nil {
		return "", err
	}
	repo, err = SanitizeName(repo)
	if err != nil {
		return "", err
	}
	repodir, err := filepath.Abs(filepath.Join(root, user, repo))
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

func createGitHooks(ctx context.Context, repodir string) error {
	log := zerolog.Ctx(ctx)
	hooksDir := filepath.Join(repodir, "hooks")
	if err := os.MkdirAll(hooksDir, os.ModePerm); err != nil {
		return fmt.Errorf("create hooks dir: %w", err)
	}

	scriptPath := filepath.Join(hooksDir, "post-receive")
	scriptContent := fmt.Sprintf(`#!/bin/sh
echo "${cat}" | %s hook post-receive
`, os.Args[0])
	if err := os.WriteFile(scriptPath, []byte(scriptContent), os.ModePerm); err != nil {
		return fmt.Errorf("write post-receive hook: %w", err)
	}

	log.Info().Str("dir", hooksDir).Msg("created git hooks")

	return nil
}

// BuildServiceAnnouncement builds the pkt-lines that announce the service per git smart protocol.
// Returned bytes include the length-prefixed header line and the terminating flush (0000).
func BuildServiceAnnouncement(service Service) ([]byte, error) {
	if !validService[service] {
		return nil, fmt.Errorf("unsupported service: %s", service)
	}
	headerLine := fmt.Sprintf("# service=%s\n", service)
	// length is 4 hex digits including those 4 bytes plus payload
	totalLen := len(headerLine) + 4
	sizeHex := strconv.FormatInt(int64(totalLen), 16)
	if len(sizeHex) < 4 {
		sizeHex = strings.Repeat("0", 4-len(sizeHex)) + sizeHex
	}
	var buf bytes.Buffer
	buf.WriteString(sizeHex)
	buf.WriteString(headerLine)
	buf.WriteString("0000") // flush packet
	return buf.Bytes(), nil
}

// AdvertiseRefs writes the advertisement header and runs the git service in advertise-refs mode.
func AdvertiseRefs(ctx context.Context, service Service, repoPath string, w io.Writer) error {
	log := zerolog.Ctx(ctx)
	ann, err := BuildServiceAnnouncement(service)
	if err != nil {
		return err
	}
	if _, err := w.Write(ann); err != nil {
		return fmt.Errorf("write announcement: %w", err)
	}
	cmd := exec.Command("git", strings.TrimPrefix(string(service), "git-"), "--stateless-rpc", "--advertise-refs", repoPath)
	var stderr bytes.Buffer
	cmd.Stdout = w
	cmd.Stderr = &stderr
	log.Debug().Strs("command", cmd.Args).Msg("executing git command")
	if err := cmd.Run(); err != nil {
		log.Error().Err(err).Str("stderr", stderr.String()).Msg("git advertise command failed")
		return err
	}
	return nil
}

// ExecStatelessRPC executes the stateless-rpc for the given service.
func ExecStatelessRPC(ctx context.Context, service Service, repoPath string, in io.Reader, w io.Writer) error {
	log := zerolog.Ctx(ctx)
	if !validService[service] {
		return fmt.Errorf("unsupported service: %s", service)
	}
	cmd := exec.Command("git", strings.TrimPrefix(string(service), "git-"), "--stateless-rpc", repoPath)
	var stderr bytes.Buffer
	cmd.Stdin = in
	cmd.Stdout = w
	cmd.Stderr = &stderr
	log.Debug().Strs("command", cmd.Args).Msg("executing git command")
	if err := cmd.Run(); err != nil {
		log.Error().Err(err).Str("stderr", stderr.String()).Msg("git rpc command failed")
		return err
	}
	return nil
}
