package handler

import (
	"encoding/json"
	"net/http"

	lookupv1 "github.com/halooid/gateway/gen/go/lookup/v1"
)

type LookupHandler struct {
	client lookupv1.LookupServiceClient
}

func NewLookupHandler(client lookupv1.LookupServiceClient) *LookupHandler {
	return &LookupHandler{client: client}
}

func (h *LookupHandler) GetLookupValues(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "missing key query param", http.StatusBadRequest)
		return
	}

	resp, err := h.client.GetLookupValues(r.Context(), &lookupv1.GetLookupValuesRequest{
		LookupKey: key,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
