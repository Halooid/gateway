package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/halooid/backend/go-shared/auth"
)

func main() {
	// Configuration
	issuerURL := os.Getenv("OIDC_ISSUER_URL")
	jwksURL := os.Getenv("OIDC_JWKS_URL")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	if issuerURL == "" || jwksURL == "" {
		log.Fatal("OIDC_ISSUER_URL and OIDC_JWKS_URL are required")
	}

	// Initialize Auth Validator (OIDC/JWKS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	validator, err := auth.NewValidator(ctx, jwksURL, 1*time.Hour)
	if err != nil {
		log.Fatalf("failed to initialize auth validator: %v", err)
	}

	// Middleware & Routing
	authMiddleware := auth.Middleware(validator)
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Placeholder for API routing
	// apiMux := http.NewServeMux()
	// Register gRPC client handlers here...
	// mux.Handle("/api/v1/", authMiddleware(apiMux))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: loggerMiddleware(mux),
	}

	// Graceful Shutdown
	go func() {
		log.Printf("Gateway starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gateway...")
	ctxShut, cancelShut := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShut()

	if err := server.Shutdown(ctxShut); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Gateway exiting")
}

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}
