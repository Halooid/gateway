package handler

import (
	"encoding/json"
	"net/http"

	authv1 "github.com/halooid/gateway/gen/go/auth/v1"
	"google.golang.org/grpc/metadata"
)

type AuthHandler struct {
	client authv1.AuthServiceClient
}

func NewAuthHandler(client authv1.AuthServiceClient) *AuthHandler {
	return &AuthHandler{client: client}
}

func (h *AuthHandler) respondWithError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *AuthHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.ValidateToken(r.Context(), &authv1.ValidateTokenRequest{
		AccessToken: req.AccessToken,
	})
	if err != nil {
		h.respondWithError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	md := metadata.Pairs("authorization", r.Header.Get("Authorization"))
	ctx := metadata.NewOutgoingContext(r.Context(), md)

	resp, err := h.client.GetCurrentUser(ctx, &authv1.GetCurrentUserRequest{})
	if err != nil {
		h.respondWithError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.User)
}
