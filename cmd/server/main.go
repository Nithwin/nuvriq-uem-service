package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"nuvriq-uem-service/internal/router"
	"nuvriq-uem-service/pkg/database"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		slog.Warn("No .env file found, using system environment variables")
	}

	log.Println("Starting Nuvriq UEM Service...")

	var db *gorm.DB
	var err error

	db, err = database.ConnectDB()
	if err != nil {
		log.Printf("[WARNING] Error connecting to database: %v. Running in degraded mode.\n", err)
	}

	mux := http.NewServeMux()

	router.RegisterRoutes(mux, db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}

	fmt.Printf("Server running in localhost:%s\n", port)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()

	shutDownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = server.Shutdown(shutDownCtx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
}
