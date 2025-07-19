package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/TFK70/kube-apiserver-audit-exporter/internal/logging"
	"github.com/TFK70/kube-apiserver-audit-exporter/pkg/collector"
	"github.com/TFK70/kube-apiserver-audit-exporter/pkg/server"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

const (
	BIN_NAME = "kube-apiserver-audit-exporter"
	VERSION  = "dev"
)

func main() {
	cmd := &cli.Command{
		Name:                  BIN_NAME,
		Usage:                 "Export kube-apiserver audit logs data as prometheus metrics",
		Version:               VERSION,
		Copyright:             fmt.Sprintf("(c) %d TFK70", time.Now().Year()),
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Value:   false,
				Usage:   "Enable debug logging",
				Sources: cli.EnvVars("KUBE_APISERVER_AUDIT_EXPORTER_DEBUG"),
			},
			&cli.IntFlag{
				Name:    "bind-port",
				Aliases: []string{"p"},
				Value:   8080,
				Usage:   "Bind port",
				Sources: cli.EnvVars("KUBE_APISERVER_AUDIT_EXPORTER_BIND_PORT"),
			},
			&cli.StringFlag{
				Name:    "audit-log-path",
				Aliases: []string{"a"},
				Value:   "/var/log/kubernetes/audit.log",
				Usage:   "Audit log path",
				Sources: cli.EnvVars("KUBE_APISERVER_AUDIT_EXPORTER_AUDIT_LOG_PATH"),
			},
		},
		Action: run,
	}

	logger := logging.SetupLogger()

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		if logging.IsNullified {
			fmt.Println(fmt.Errorf("Error during execution: %v", err))
		} else {
			logger.Errorf("Error during execution: %v", err)
		}

		os.Exit(1)
	}
}

func run(context context.Context, cmd *cli.Command) error {
	rootLogger, err := logging.GetLogger()
	if err != nil {
		return err
	}

	if cmd.Bool("debug") {
		rootLogger.SetLevel(logrus.DebugLevel)
	}

	c, err := collector.NewAPIServerRequestsCollector(
		collector.WithAuditLogPath(cmd.String("audit-log-path")),
	)
	if err != nil {
		return fmt.Errorf("failed to create collector: %v", err)
	}

	c.Start()
	defer c.Stop()

	s, err := server.NewServer(
		server.WithBindPort(cmd.Int("bind-port")),
		server.WithHander("/metrics", collector.HandleMetrics),
	)
	if err != nil {
		return fmt.Errorf("failed to create server: %v", err)
	}

	errch := make(chan error)
	defer close(errch)

	go func() {
		err := s.Start()
		if err != nil {
			errch <- fmt.Errorf("Failed to start server: %v", err)
		}
	}()

	return  <-errch
}
