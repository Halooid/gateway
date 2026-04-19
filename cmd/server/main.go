package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	authv1 "github.com/halooid/gateway/gen/go/auth/v1"
	"github.com/halooid/gateway/internal/handler"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	// gRPC Connections
	authSvcAddr := os.Getenv("AUTH_SERVICE_ADDR")
	if authSvcAddr == "" {
		authSvcAddr = "localhost:50051"
	}

	authConn, err := grpc.Dial(authSvcAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to auth service: %v", err)
	}
	defer authConn.Close()

	authClient := authv1.NewAuthServiceClient(authConn)
	authHandler := handler.NewAuthHandler(authClient)

	// Middleware & Routing
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Auth Routes
	mux.HandleFunc("/api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("/api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("/api/v1/auth/me", authHandler.Me)
	mux.HandleFunc("/api/v1/auth/refresh", authHandler.RefreshToken)
	mux.HandleFunc("/api/v1/auth/validate", authHandler.ValidateToken)

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
