package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cnmt/internal/app"
	"cnmt/internal/common/env"
	"cnmt/internal/infra/migrate"

	"github.com/jackc/pgx/v5/pgxpool"
	"riverqueue.com/riverui"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			if err := runMigrate(os.Args[2:]); err != nil {
				log.Fatalf("migrate: %v", err)
			}
			return
		case "help", "-h", "--help":
			printUsage()
			return
		}
	}

	runServer()
}

func runMigrate(args []string) error {
	databaseURL := env.GetString("DATABASE_URL", "")
	cmd := "up"
	if len(args) > 0 {
		cmd = args[0]
	}

	switch cmd {
	case "up":
		log.Println("running migrations...")
		if err := migrate.Up(databaseURL); err != nil {
			return err
		}
		log.Println("migrations applied")
		return nil
	case "status":
		return migrate.Status(databaseURL)
	default:
		return fmt.Errorf("unknown migrate command %q (supported: up, status)", cmd)
	}
}

func runServer() {
	ctx := context.Background()
	dbConn, err := pgxpool.New(ctx, env.GetString("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/cnmt"))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	application := app.NewApp(dbConn)

	app, err := application.Run()
	if err != nil {
		log.Fatalf("failed to run app: %v", err)
	}

	
	endpoints := riverui.NewEndpoints(app.WorkerClient,nil)
	opts := &riverui.HandlerOpts{
		Endpoints: endpoints,
		Prefix: "/riverui",
		Logger: app.Logger,
	}

	handler, err := riverui.NewHandler(opts)

	if err != nil {
		log.Fatalf("failed to create riverui handler: %v", err)
	}

	if err := handler.Start(ctx); err != nil {
		log.Fatalf("failed to start riverui: %v", err)
	}

	// Mount outside chi so StripSlashes/CleanPath do not rewrite /riverui/.
	mux := http.NewServeMux()
	mux.Handle("/riverui/", handler)
	mux.Handle("/", app.Router)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", env.GetString("HOST", ""), env.GetString("PORT", "8080")),
		Handler: mux,
	}

	if err := app.WorkerClient.Start(ctx); err != nil {
		log.Fatalf("failed to start workers: %v", err)
	}


	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	// Stop worker client after HTTP server so in-flight requests can still enqueue jobs.
	if err := app.WorkerClient.Stop(shutdownCtx); err != nil {
		log.Printf("Worker shutdown error: %v", err)
	}

	log.Println("Server shut down gracefully")
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `cnmt api

Usage:
  api                 Start the HTTP server
  api migrate         Apply pending migrations (goose up)
  api migrate up      Same as migrate
  api migrate status  Show migration status

Environment:
  DATABASE_URL        Postgres connection string (required for migrate)
  HOST                Listen host (default empty = all interfaces)
  PORT                Listen port (default 8080)
`)
}
