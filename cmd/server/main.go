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
	"google.golang.org/grpc/credentials/insecure"

	authv1 "github.com/halooid/gateway/gen/go/auth/v1"
	lookupv1 "github.com/halooid/gateway/gen/go/lookup/v1"
	splixv1 "github.com/halooid/gateway/gen/go/splix/v1"
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

	lookupSvcAddr := os.Getenv("LOOKUP_SERVICE_ADDR")
	if lookupSvcAddr == "" {
		lookupSvcAddr = "localhost:50052"
	}

	splixSvcAddr := os.Getenv("SPLIX_SERVICE_ADDR")
	if splixSvcAddr == "" {
		splixSvcAddr = "localhost:50053"
	}

	authConn, err := grpc.NewClient(authSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to auth service: %v", err)
	}
	defer authConn.Close()

	lookupConn, err := grpc.NewClient(lookupSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to lookup service: %v", err)
	}
	defer lookupConn.Close()

	splixConn, err := grpc.NewClient(splixSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to splix service: %v", err)
	}
	defer splixConn.Close()

	authClient := authv1.NewAuthServiceClient(authConn)
	authHandler := handler.NewAuthHandler(authClient)

	lookupClient := lookupv1.NewLookupServiceClient(lookupConn)
	lookupHandler := handler.NewLookupHandler(lookupClient)

	splixUserClient := splixv1.NewUserServiceClient(splixConn)
	splixGroupClient := splixv1.NewGroupServiceClient(splixConn)
	splixExpenseClient := splixv1.NewExpenseServiceClient(splixConn)
	splixHandler := handler.NewSplixHandler(splixUserClient, splixGroupClient, splixExpenseClient)

	// Middleware & Routing
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Auth Routes
	mux.HandleFunc("/api/v1/auth/me", authHandler.Me)
	mux.HandleFunc("/api/v1/auth/validate", authHandler.ValidateToken)

	// Lookup Routes
	mux.HandleFunc("/api/v1/lookup", lookupHandler.GetLookupValues)

	// Splix Routes
	mux.HandleFunc("/api/v1/splix/users", splixHandler.CreateUser)
	mux.HandleFunc("/api/v1/splix/connections", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			splixHandler.AddConnection(w, r)
		} else {
			splixHandler.ListConnections(w, r)
		}
	})
	mux.HandleFunc("/api/v1/splix/groups", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			splixHandler.CreateGroup(w, r)
		} else {
			splixHandler.ListGroups(w, r)
		}
	})
	mux.HandleFunc("/api/v1/splix/expenses", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			splixHandler.AddExpense(w, r)
		} else {
			splixHandler.GetExpenses(w, r)
		}
	})
	mux.HandleFunc("/api/v1/splix/balances", splixHandler.GetBalances)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: corsMiddleware(loggerMiddleware(mux)),
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

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}
