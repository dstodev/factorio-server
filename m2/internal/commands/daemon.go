package commands

import (
	"context"
	"os"
	"time"

	"github.com/urfave/cli/v3"

	"manage2/internal/log"
)

func Daemon(ctx context.Context, c *cli.Command) error {
	closeLogFile := log.InitFile("m2d.log", c.String("log-level"))
	defer closeLogFile()

	log.Info("Daemon started")
	log.Debug("Started with args", "args", os.Args)
	time.Sleep(1 * time.Second)
	log.Info("Daemon finished")

	return nil
}
