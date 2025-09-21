package gitproto

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
	repodir := filepath.Join(root, user, repo)
	if _, err := os.Stat(repodir); os.IsNotExist(err) {
		if err := os.MkdirAll(repodir, 0o744); err != nil {
			return "", fmt.Errorf("create repo dir: %w", err)
		}
		if initErr := exec.Command("git", "init", "--bare", repodir).Run(); initErr != nil {
			return "", fmt.Errorf("init bare repo: %w", initErr)
		}
		log.Info().Str("dir", repodir).Msg("initialized bare git repository")
	}
	abs, err := filepath.Abs(repodir)
	if err != nil {
		return "", err
	}
	return abs, nil
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
