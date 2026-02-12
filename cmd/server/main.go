package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend-test-golang/internal/config"
	"backend-test-golang/internal/handlers"
	"backend-test-golang/internal/repository"
	"backend-test-golang/internal/services"
	"backend-test-golang/pkg/cache"
	"backend-test-golang/pkg/database"
	"backend-test-golang/pkg/middlewares"
	"backend-test-golang/pkg/skinport"
)

func main() {
	conf := config.Load()

	db, err := database.Connect(conf.DBUrl)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	skinportClient, err := skinport.NewClient(conf.SkinportClientID, conf.SkinportClientSecret, conf.SkinportAddr)
	if err != nil {
		log.Fatalf("Failed to create skinport client: %v", err)
	}

	mcache := cache.New(conf.CacheCleanUpIntervalSeconds)
	defer mcache.Close()

	repo := repository.New(db)
	svc := services.New(conf.CacheTTLSeconds, mcache, skinportClient, repo)
	handler := handlers.New(svc)

	mux := http.NewServeMux()

	mux.Handle("/api/v1/items", middlewares.GzipEncode(http.HandlerFunc(handler.GetItems)))

	mux.HandleFunc("/api/v1/withdraw", handler.Withdraw)
	mux.HandleFunc("/api/v1/user/balance", handler.GetBalance)
	mux.HandleFunc("/api/v1/user/transactions", handler.GetTransactions)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	srv := &http.Server{
		Addr:         conf.Addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on %s", conf.Addr)
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Printf("Received signal %v, shutting down server...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err = srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}
