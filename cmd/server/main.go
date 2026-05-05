// Program entry point; initializes the application and starts the HTTP server and background scheduler.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"oj-lite/internal/app"
	"oj-lite/internal/scheduler"
)

type serverOptions struct {
	skipSeed bool
}

func main() {
	os.Exit(run())
}

func run() int {
	options, err := parseServerOptions(os.Args[1:], os.Stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	application, err := app.BootstrapWithOptions(app.BootstrapOptions{
		SkipSeed: options.skipSeed,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "bootstrap failed: %v\n", err)
		return 1
	}

	if err := runApplication(application); err != nil {
		application.Logger().Errorf("server exited with error: %v", err)
		return 1
	}

	application.Logger().Infof("server stopped")
	return 0
}

func parseServerOptions(args []string, output io.Writer) (serverOptions, error) {
	var options serverOptions

	flags := flag.NewFlagSet("server", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.BoolVar(&options.skipSeed, "skip-seed", false, "skip demo data seeding on startup")

	if err := flags.Parse(args); err != nil {
		return options, err
	}

	return options, nil
}

func runApplication(application *app.App) error {
	var runErr error

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	schedulerModule := scheduler.New(application.DB(), application.Logger(), application.Config().Scheduler)
	schedulerDone := make(chan struct{})
	go func() {
		defer close(schedulerDone)
		schedulerModule.Run(runCtx)
	}()

	errCh := make(chan error, 1)
	go func() {
		errCh <- application.Run()
	}()

	select {
	case err := <-errCh:
		cancel()
		if err != nil {
			runErr = fmt.Errorf("run http server: %w", err)
		}
	case <-ctx.Done():
		cancel()
		application.Logger().Infof("shutdown signal received")
	}

	<-schedulerDone

	shutdownCtx, cancel := context.WithTimeout(context.Background(), application.Config().HTTP.ShutdownTimeout)
	defer cancel()

	if err := application.Shutdown(shutdownCtx); err != nil {
		runErr = errors.Join(runErr, fmt.Errorf("graceful shutdown: %w", err))
	}

	return runErr
}
