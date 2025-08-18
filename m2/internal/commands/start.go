package commands

import (
	"context"
	"os"
	"os/exec"

	"github.com/urfave/cli/v3"

	"manage2/internal/log"
)

func Start(ctx context.Context, c *cli.Command) error {
	logLevelStr := c.String("log-level")
	log.InitTerminal(logLevelStr)

	exe, err := os.Executable()
	if err != nil {
		log.Error("Failed to get executable", "error", err)
		return err
	}

	cmd := exec.Command(exe,
		"--log-level", logLevelStr,
		"--game", c.String("game"),
		"daemon")

	log.Debug("Starting daemon", "command", cmd.String())

	if err := cmd.Start(); err != nil {
		log.Error("Failed to start daemon", "error", err)
		return err
	}

	pid := cmd.Process.Pid

	if err := cmd.Process.Release(); err != nil {
		log.Error("Failed to release daemon", "error", err)
		return err
	}

	log.Debug("Released daemon", "pid", pid)

	return nil
}
