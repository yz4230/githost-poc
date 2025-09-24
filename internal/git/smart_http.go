package git

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
)

const (
	ServiceUploadPack  = "git-upload-pack"
	ServiceReceivePack = "git-receive-pack"
	Repositories       = "repositories"
)

// buildServiceAnnouncement builds the pkt-lines that announce the service per git smart protocol.
// Returned bytes include the length-prefixed header line and the terminating flush (0000).
func buildServiceAnnouncement(service string) ([]byte, error) {
	if service != ServiceUploadPack && service != ServiceReceivePack {
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
func AdvertiseRefs(ctx context.Context, service string, repoPath string, w io.Writer) error {
	log := zerolog.Ctx(ctx)
	ann, err := buildServiceAnnouncement(service)
	if err != nil {
		return err
	}
	if _, err := w.Write(ann); err != nil {
		return fmt.Errorf("write announcement: %w", err)
	}
	cmd := exec.Command("git", strings.TrimPrefix(service, "git-"), "--stateless-rpc", "--advertise-refs", repoPath)
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
func ExecStatelessRPC(ctx context.Context, service string, repoPath string, in io.Reader, w io.Writer) error {
	log := zerolog.Ctx(ctx)
	if service != ServiceUploadPack && service != ServiceReceivePack {
		return fmt.Errorf("unsupported service: %s", service)
	}
	cmd := exec.Command("git", strings.TrimPrefix(service, "git-"), "--stateless-rpc", repoPath)
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
