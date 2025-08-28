package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// DebugBootstrap generates random users for testing
func (h *handler) DebugBootstrap(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	countStr := r.URL.Query().Get("count")
	seedStr := r.URL.Query().Get("seed")

	// Default values
	count := 10
	seed := uint64(1000)

	// Parse count if provided
	if countStr != "" {
		if c, err := strconv.Atoi(countStr); err == nil && c > 0 {
			count = c
		}
	}

	// Parse seed if provided
	if seedStr != "" {
		if s, err := strconv.ParseUint(seedStr, 10, 64); err == nil {
			seed = s
		}
	}

	// Generate users
	userIds, err := h.Nexus.DebugBootstrap(r.Context(), count, seed)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to bootstrap users: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := map[string]any{
		"message":    fmt.Sprintf("Generated %d test users and populated pages via GetNexus calls", len(userIds)),
		"user_ids":   userIds,
		"count":      len(userIds),
		"seed_start": seed,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
	}
}
