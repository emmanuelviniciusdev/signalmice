package shutdown

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/signalmice/signalmice/internal/config"
	"github.com/signalmice/signalmice/internal/logger"
)

// Manager handles host machine shutdown
type Manager struct {
	hostProcPath string
	logger       *logger.Logger
}

// NewManager creates a new shutdown manager
func NewManager(cfg *config.Config, log *logger.Logger) *Manager {
	return &Manager{
		hostProcPath: cfg.HostProcPath,
		logger:       log,
	}
}

// NeutralizeStuartLittle attempts to shutdown the host machine using multiple methods.
// This function catches the shutdown signal and neutralizes the target machine.
// https://www.reddit.com/r/stuartlittlefacts/
func (m *Manager) NeutralizeStuartLittle(ctx context.Context) error {
	m.logger.Info(ctx, "Initiating host machine shutdown...")

	// Try multiple methods in order of preference
	methods := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"nsenter", m.shutdownViaNsenter},
		{"sysrq-trigger", m.shutdownViaSysrq},
		{"direct-command", m.shutdownViaDirect},
	}

	var lastErr error
	for _, method := range methods {
		m.logger.InfoWithExtra(ctx, fmt.Sprintf("Attempting shutdown via %s", method.name), nil)
		if err := method.fn(ctx); err != nil {
			m.logger.WarnWithExtra(ctx, fmt.Sprintf("Shutdown via %s failed", method.name), map[string]string{"error": err.Error()})
			lastErr = err
			continue
		}
		m.logger.Info(ctx, fmt.Sprintf("Shutdown initiated successfully via %s", method.name))
		return nil
	}

	return fmt.Errorf("all shutdown methods failed, last error: %w", lastErr)
}

// shutdownViaNsenter uses nsenter to enter the host namespace and run shutdown
func (m *Manager) shutdownViaNsenter(ctx context.Context) error {
	// Use nsenter to enter the host's namespace and run poweroff
	// This requires --privileged and --pid=host on the container
	cmd := exec.CommandContext(ctx,
		"nsenter",
		"--target", "1",
		"--mount",
		"--uts",
		"--ipc",
		"--net",
		"--pid",
		"--",
		"poweroff",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nsenter poweroff failed: %w, output: %s", err, string(output))
	}

	return nil
}

// shutdownViaSysrq uses the sysrq-trigger to power off the machine
func (m *Manager) shutdownViaSysrq(ctx context.Context) error {
	// First, sync all filesystems
	syncPath := filepath.Join(m.hostProcPath, "sysrq-trigger")

	// Check if we have access to host's proc
	if _, err := os.Stat(m.hostProcPath); os.IsNotExist(err) {
		return fmt.Errorf("host proc path not mounted: %s", m.hostProcPath)
	}

	// Sync filesystems first (sysrq 's')
	if err := os.WriteFile(syncPath, []byte("s"), 0644); err != nil {
		m.logger.Warn(ctx, "Failed to sync filesystems via sysrq")
	}

	// Remount filesystems read-only (sysrq 'u')
	if err := os.WriteFile(syncPath, []byte("u"), 0644); err != nil {
		m.logger.Warn(ctx, "Failed to remount filesystems read-only via sysrq")
	}

	// Power off (sysrq 'o')
	if err := os.WriteFile(syncPath, []byte("o"), 0644); err != nil {
		return fmt.Errorf("failed to write to sysrq-trigger: %w", err)
	}

	return nil
}

// shutdownViaDirect uses the shutdown command directly
// This only works if the container has access to host's init system
func (m *Manager) shutdownViaDirect(ctx context.Context) error {
	// Try poweroff command
	cmd := exec.CommandContext(ctx, "poweroff")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try shutdown -h now as fallback
		cmd = exec.CommandContext(ctx, "shutdown", "-h", "now")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("shutdown commands failed: %w, output: %s", err, string(output))
		}
	}

	return nil
}
