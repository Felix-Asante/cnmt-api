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

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	dbConn, err := pgxpool.New(ctx, env.GetString("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/cnmt"))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	defer dbConn.Close()

	app := app.NewApp(dbConn)

	r, err := app.Run()
	if err != nil {
		log.Fatalf("failed to run app: %v", err)
	}

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", env.GetString("HOST", ""), env.GetString("PORT", "8080")),
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
	log.Println("Server shut down gracefully")
}
