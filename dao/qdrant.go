package dao

import (
	"time"

	"github.com/dbrun3/nexus-vector/model"
)

// QdrantPagePayload represents the structure stored in Qdrant for page documents
type QdrantPagePayload struct {
	Page      model.Page `json:"page"`
	CreatedAt int64      `json:"created_at"`
	From      int64      `json:"from"`
	Until     int64      `json:"until"`
}

// NewQdrantPagePayload creates a new payload with the current timestamp
func NewQdrantPagePayload(page model.Page, from, until int64) QdrantPagePayload {
	return QdrantPagePayload{
		Page:      page,
		CreatedAt: time.Now().Unix(),
		From:      from,
		Until:     until,
	}
}

// ToMap converts the payload to map[string]any for Qdrant storage
func (q QdrantPagePayload) ToMap() map[string]any {
	// Convert string slices to []any for Qdrant compatibility
	titleSlice := make([]any, len(q.Page.Title))
	for i, title := range q.Page.Title {
		titleSlice[i] = title
	}

	subTitleSlice := make([]any, len(q.Page.SubTitle))
	for i, subTitle := range q.Page.SubTitle {
		subTitleSlice[i] = subTitle
	}

	return map[string]any{
		"page": map[string]any{
			"id":       q.Page.Id,
			"layout":   q.Page.Layout,
			"type":     q.Page.Type,
			"category": q.Page.Category,
			"title":    titleSlice,
			"subTitle": subTitleSlice,
		},
		"created_at": q.CreatedAt,
		"from":       q.From,
		"until":      q.Until,
	}
}
