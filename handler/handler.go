package handler

import (
	"net/http"

	"github.com/dbrun3/nexus-vector/nexus"
)

type handler struct {
	Nexus nexus.Nexus
}

func NewHandler(n nexus.Nexus) *handler {
	return &handler{Nexus: n}
}

func (h *handler) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /get-nexus", h.GetNexus)
	mux.HandleFunc("PUT /injest-user", h.InjestUser)
	return mux
}
