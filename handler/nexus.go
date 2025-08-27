package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dbrun3/nexus-vector/api"
	"github.com/dbrun3/nexus-vector/model"
)

func (h *handler) GetNexus(w http.ResponseWriter, r *http.Request) {
	var request api.NexusRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	pages, err := h.Nexus.GetNexus(r.Context(), request)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get pages: %v", err), http.StatusInternalServerError)
		return
	}

	// 204 when no pages
	if len(pages) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// 200 with NexusResponse containing pages
	response := api.NexusResponse{Pages: pages}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
	}
}

func (h *handler) InjectUser(w http.ResponseWriter, r *http.Request) {
	var userSnapshot model.UserSnapshot

	if err := json.NewDecoder(r.Body).Decode(&userSnapshot); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	embedding, err := h.Nexus.InjestUser(r.Context(), userSnapshot)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to inject user: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := map[string]any{
		"user_id":   userSnapshot.ID,
		"embedding": embedding,
		"status":    "injected",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
	}
}
