package api

import "github.com/dbrun3/nexus-vector/model"

type NexusRequest struct {
	UserId  string        `json:"userId"`
	Trigger model.Trigger `json:"trigger"`
}

type NexusResponse struct {
	// Pages is an array of pages that will be shown to a user after receipt details for CONTENT or ACTION
	Pages []model.Page `json:"pages"`
}
