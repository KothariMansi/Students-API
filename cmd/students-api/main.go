package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KothariMansi/students-api/internal/config"
	"github.com/KothariMansi/students-api/internal/http/handlers/student"
	"github.com/KothariMansi/students-api/internal/storage/sqlite"
)

func main() {
	// load config
	fmt.Println("Main function started") 
	cfg := config.MustLoad()

	// database setup
	storage, err := sqlite.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	slog.Info("Storage Initialized", slog.String("env", cfg.Env), slog.String("version", "1.0.0"))

	// setup router
	router := http.NewServeMux()
	router.HandleFunc("POST /api/students", student.New(storage))
	fmt.Println("Server setup done")

	// setup server
	server := http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}
	slog.Info("server started", slog.String("address", cfg.Addr))

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	// Add gracefull shutdown
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			fmt.Println(err)
			log.Fatal("Failed to start server")
		}
	}()

	<-done

	slog.Info("Shutting down the server....")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		slog.Error("Failed to shutdown server", slog.String("error", err.Error()))
	}
	slog.Info("Server shutdown successfully.")
}
