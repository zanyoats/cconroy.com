package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/zanyoats/cconroy.com/internal/config"
	"github.com/zanyoats/cconroy.com/internal/ops"
	"github.com/zanyoats/cconroy.com/internal/page"
	"github.com/zanyoats/cconroy.com/internal/server"
)

const (
	port            = 8000
	startupTimeout  = 10 * time.Second
	shutdownTimeout = 10 * time.Second
)

type cliOptions struct {
	notesPath string
}

func main() {
	opts, err := parseOptions(os.Args[1:], os.Stderr)
	switch {
	case errors.Is(err, flag.ErrHelp):
		return
	case err != nil:
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if err := run(opts); err != nil {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}

func parseOptions(args []string, output io.Writer) (cliOptions, error) {
	flags := flag.NewFlagSet("web", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.Usage = func() {
		fmt.Fprintln(output, "Usage: web --notes-path PATH")
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Options:")
		fmt.Fprintln(output, "  --notes-path PATH")
		fmt.Fprintln(output, "        directory containing Markdown note files (required)")
		fmt.Fprintln(output, "  -h, --help")
		fmt.Fprintln(output, "        show this help")
	}

	notesPath := flags.String("notes-path", "", "directory containing Markdown note files")
	if err := flags.Parse(args); err != nil {
		return cliOptions{}, err
	}

	if flags.NArg() != 0 {
		return cliOptions{}, fmt.Errorf("unexpected positional arguments: %q", flags.Args())
	}
	if strings.TrimSpace(*notesPath) == "" {
		return cliOptions{}, errors.New("--notes-path is required; use --help for usage")
	}

	return cliOptions{notesPath: *notesPath}, nil
}

func run(opts cliOptions) (runErr error) {
	appCtx, stopSignals := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stopSignals()

	conf, err := config.InitializeConfig()
	if err != nil {
		return fmt.Errorf("initialize config: %w", err)
	}

	startupCtx, cancelStartup := context.WithTimeout(appCtx, startupTimeout)
	defer cancelStartup()

	writeDB, err := sql.Open("sqlite3", conf.WriteDSN)
	if err != nil {
		return fmt.Errorf("open SQLite write pool: %w", err)
	}
	defer func() {
		if err := writeDB.Close(); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("close SQLite write pool: %w", err))
		}
	}()

	writeDB.SetMaxOpenConns(1)
	writeDB.SetMaxIdleConns(1)

	if err := writeDB.PingContext(startupCtx); err != nil {
		return fmt.Errorf("ping SQLite write pool: %w", err)
	}

	var journalMode string
	if err = writeDB.QueryRowContext(
		startupCtx,
		`PRAGMA journal_mode = WAL`,
	).Scan(&journalMode); err != nil {
		return fmt.Errorf("enable SQLite WAL mode: %w", err)
	}

	if journalMode != "wal" {
		return fmt.Errorf("enable SQLite WAL mode: got journal mode %q", journalMode)
	}

	readDB, err := sql.Open("sqlite3", conf.ReadDSN)
	if err != nil {
		return fmt.Errorf("open SQLite read pool: %w", err)
	}
	defer func() {
		if err := readDB.Close(); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("close SQLite read pool: %w", err))
		}
	}()

	readDB.SetMaxOpenConns(4)
	readDB.SetMaxIdleConns(4)

	if err := readDB.PingContext(startupCtx); err != nil {
		return fmt.Errorf("ping SQLite read pool: %w", err)
	}

	noteOps := ops.NewNotesImpl(readDB, writeDB, opts.notesPath)

	if err := noteOps.PublishStaticNotes(startupCtx); err != nil {
		return fmt.Errorf("publish static notes: %w", err)
	}
	cancelStartup()

	renderer, err := page.NewRenderer()
	if err != nil {
		return fmt.Errorf("initialize page renderer: %w", err)
	}

	webRouter := server.NewWebRouter(noteOps, renderer)
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: webRouter,
	}
	serveErr := make(chan error, 1)

	log.Printf("listening on http://localhost:%d", port)
	go func() {
		serveErr <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)

	case <-appCtx.Done():
		// Restore the default signal behavior so a second signal can force exit.
		stopSignals()
		log.Print("shutdown signal received; draining HTTP requests")
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(
		context.Background(),
		shutdownTimeout,
	)
	defer cancelShutdown()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		shutdownErr := fmt.Errorf("graceful HTTP shutdown: %w", err)
		if closeErr := httpServer.Close(); closeErr != nil {
			return errors.Join(
				shutdownErr,
				fmt.Errorf("force HTTP server closed: %w", closeErr),
			)
		}
		return shutdownErr
	}

	if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP during shutdown: %w", err)
	}

	log.Print("HTTP server stopped; closing database pools")
	return nil
}
