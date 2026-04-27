package handler

import (
	"context"
	"encoding/json"
	"net/http"

	splixv1 "github.com/halooid/gateway/gen/go/splix/v1"
	"google.golang.org/grpc/metadata"
)

type SplixHandler struct {
	userClient    splixv1.UserServiceClient
	groupClient   splixv1.GroupServiceClient
	expenseClient splixv1.ExpenseServiceClient
}

func NewSplixHandler(userClient splixv1.UserServiceClient, groupClient splixv1.GroupServiceClient, expenseClient splixv1.ExpenseServiceClient) *SplixHandler {
	return &SplixHandler{
		userClient:    userClient,
		groupClient:   groupClient,
		expenseClient: expenseClient,
	}
}

func (h *SplixHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req splixv1.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.userClient.CreateUser(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *SplixHandler) AddConnection(w http.ResponseWriter, r *http.Request) {
	var req splixv1.AddConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := passAuth(r)
	resp, err := h.userClient.AddConnection(ctx, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *SplixHandler) ListConnections(w http.ResponseWriter, r *http.Request) {
	ctx := passAuth(r)
	resp, err := h.userClient.ListConnections(ctx, &splixv1.ListConnectionsRequest{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *SplixHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var req splixv1.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := passAuth(r)
	resp, err := h.groupClient.CreateGroup(ctx, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *SplixHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	ctx := passAuth(r)
	resp, err := h.groupClient.ListGroups(ctx, &splixv1.ListGroupsRequest{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *SplixHandler) AddExpense(w http.ResponseWriter, r *http.Request) {
	var req splixv1.AddExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := passAuth(r)
	resp, err := h.expenseClient.AddExpense(ctx, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *SplixHandler) GetBalances(w http.ResponseWriter, r *http.Request) {
	ctx := passAuth(r)
	resp, err := h.expenseClient.GetUserBalances(ctx, &splixv1.GetUserBalancesRequest{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *SplixHandler) GetExpenses(w http.ResponseWriter, r *http.Request) {
	groupID := r.URL.Query().Get("group_id")
	ctx := passAuth(r)
	resp, err := h.expenseClient.GetUserExpenses(ctx, &splixv1.GetUserExpensesRequest{
		GroupId: groupID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// passAuth extracts the Authorization header and passes it as gRPC metadata
func passAuth(r *http.Request) context.Context {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return r.Context()
	}
	md := metadata.Pairs("authorization", authHeader)
	return metadata.NewOutgoingContext(r.Context(), md)
}
