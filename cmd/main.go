package main

import (
	"log/slog"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"fixit/web/app"
)

var rootCmd = &cobra.Command{
	Use:   "fixit",
	Short: "FixIt - A web application server",
	Long:  "FixIt is a web application server with various features and capabilities.",
}

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start the web server",
	Long:  "Start the web server and listen for HTTP connections",
	RunE:  runWebServer,
}

var (
	port string
)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	webCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to listen on")
	rootCmd.AddCommand(webCmd)
}

func runWebServer(cmd *cobra.Command, args []string) error {
	slog.Info("starting web server")

	cfg := app.Config{}
	if err := env.Parse(&cfg); err != nil {
		return errors.Wrap(err, "failed to parse env")
	}
	webapp, err := app.New(cfg)
	if err != nil {
		return err
	}

	slog.Info("server listening", "port", cfg.Port)
	return webapp.Start()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("failed to execute command", "error", err)
		os.Exit(1)
	}
}
