package nexus

import (
	"encoding/json"
	"testing"

	"github.com/dbrun3/nexus-vector/model"
	"github.com/qdrant/go-client/qdrant"
)

func TestConvertResultsToRelevantPages(t *testing.T) {
	// Create test page data
	testPage := model.Page{
		Id:       "test-page-1",
		Layout:   "card",
		Type:     "offer",
		Category: "electronics",
		Title:    []string{"Great Deal"},
		SubTitle: []string{"Limited Time"},
	}

	// Create Qdrant payload structure
	pageFields := map[string]*qdrant.Value{
		"id": {
			Kind: &qdrant.Value_StringValue{StringValue: testPage.Id},
		},
		"layout": {
			Kind: &qdrant.Value_StringValue{StringValue: testPage.Layout},
		},
		"type": {
			Kind: &qdrant.Value_StringValue{StringValue: testPage.Type},
		},
		"category": {
			Kind: &qdrant.Value_StringValue{StringValue: testPage.Category},
		},
		"title": {
			Kind: &qdrant.Value_ListValue{
				ListValue: &qdrant.ListValue{
					Values: []*qdrant.Value{
						{Kind: &qdrant.Value_StringValue{StringValue: "Great Deal"}},
					},
				},
			},
		},
		"subTitle": {
			Kind: &qdrant.Value_ListValue{
				ListValue: &qdrant.ListValue{
					Values: []*qdrant.Value{
						{Kind: &qdrant.Value_StringValue{StringValue: "Limited Time"}},
					},
				},
			},
		},
	}

	payload := map[string]*qdrant.Value{
		"page": {
			Kind: &qdrant.Value_StructValue{StructValue: &qdrant.Struct{Fields: pageFields}},
		},
	}

	tests := []struct {
		name     string
		results  []*qdrant.ScoredPoint
		minScore float32
		expected int
	}{
		{
			name: "score above threshold",
			results: []*qdrant.ScoredPoint{
				{
					Score:   1.0,
					Payload: payload,
				},
			},
			minScore: 0.9,
			expected: 1,
		},
		{
			name: "score equal to threshold",
			results: []*qdrant.ScoredPoint{
				{
					Score:   0.9,
					Payload: payload,
				},
			},
			minScore: 0.9,
			expected: 1, // Should include equal scores
		},
		{
			name: "score below threshold",
			results: []*qdrant.ScoredPoint{
				{
					Score:   0.8,
					Payload: payload,
				},
			},
			minScore: 0.9,
			expected: 0,
		},
		{
			name: "multiple results mixed scores",
			results: []*qdrant.ScoredPoint{
				{Score: 1.0, Payload: payload},
				{Score: 0.95, Payload: payload},
				{Score: 0.85, Payload: payload}, // Should be excluded
			},
			minScore: 0.9,
			expected: 2,
		},
		{
			name:     "empty results",
			results:  []*qdrant.ScoredPoint{},
			minScore: 0.9,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pages := convertResultsToRelevantPages(tt.results, tt.minScore)

			if len(pages) != tt.expected {
				t.Errorf("convertResultsToRelevantPages() returned %d pages, expected %d", len(pages), tt.expected)

				// Debug info
				for i, result := range tt.results {
					t.Logf("Result %d: score=%.4f, minScore=%.4f, passes=%t",
						i, result.Score, tt.minScore, result.Score > tt.minScore)

					// Debug the payload extraction
					if dataValue, ok := result.Payload["page"]; ok {
						t.Logf("  - Found 'page' key in payload")
						if structValue := dataValue.GetStructValue(); structValue != nil {
							t.Logf("  - StructValue is not nil")
							jsonBytes, err := json.Marshal(structValue.GetFields())
							if err != nil {
								t.Logf("  - JSON Marshal error: %v", err)
							} else {
								t.Logf("  - JSON Marshal success: %s", string(jsonBytes))
								var page model.Page
								if err := json.Unmarshal(jsonBytes, &page); err != nil {
									t.Logf("  - JSON Unmarshal error: %v", err)
								} else {
									t.Logf("  - JSON Unmarshal success: %+v", page)
								}
							}
						} else {
							t.Logf("  - StructValue is nil")
						}
					} else {
						t.Logf("  - No 'page' key found in payload")
					}
				}
			}

			// Verify the page data is correctly extracted
			if len(pages) > 0 {
				page := pages[0]
				if page.Id != testPage.Id {
					t.Errorf("Expected page ID %s, got %s", testPage.Id, page.Id)
				}
				if page.Layout != testPage.Layout {
					t.Errorf("Expected layout %s, got %s", testPage.Layout, page.Layout)
				}
			}
		})
	}
}

func TestConvertResultsToRelevantPages_InvalidPayload(t *testing.T) {
	// Test with invalid payload structure
	invalidPayload := map[string]*qdrant.Value{
		"invalid": {
			Kind: &qdrant.Value_StringValue{StringValue: "not a page"},
		},
	}

	results := []*qdrant.ScoredPoint{
		{
			Score:   1.0,
			Payload: invalidPayload,
		},
	}

	pages := convertResultsToRelevantPages(results, 0.5)

	if len(pages) != 0 {
		t.Errorf("Expected 0 pages for invalid payload, got %d", len(pages))
	}
}
