package nexus

import (
	"context"
	"fmt"
	"time"

	"github.com/dbrun3/nexus-vector/dao"
	"github.com/dbrun3/nexus-vector/model"
	"github.com/dbrun3/nexus-vector/util"
	"github.com/qdrant/go-client/qdrant"
)

// queryQdrant queries the db for embeddings for pages where the current day exists within their eligible time range
func (n *Nexus) queryQdrant(ctx context.Context, embedding []float32) ([]*qdrant.ScoredPoint, error) {
	now := float64(time.Now().Unix())
	return n.qdClient.Query(ctx, &qdrant.QueryPoints{
		CollectionName: PageCollection,
		Query:          qdrant.NewQuery(embedding...),
		WithPayload:    qdrant.NewWithPayload(true),
		Filter: &qdrant.Filter{
			Must: []*qdrant.Condition{
				qdrant.NewRange("from", &qdrant.Range{
					Lte: &now,
				}),
				qdrant.NewRange("until", &qdrant.Range{
					Gte: &now,
				}),
			},
		},
		Limit: qdrant.PtrOf(uint64(4)),
	})
}

func (n *Nexus) getAsyncResults(ctx context.Context, userId string) ([]*qdrant.ScoredPoint, []float32, error) {
	redisEmbedding, err := n.rdClient.Get(ctx, userId).Result()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get embedding: %w", err)
	}
	userEmbedding, err := dao.EmbeddingFromRedis(redisEmbedding)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert embedding: %w", err)
	}

	userResults, err := n.queryQdrant(ctx, userEmbedding)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get pages: %w", err)
	}

	return userResults, userEmbedding, nil
}

func (n *Nexus) getSyncResults(ctx context.Context, trigger model.Trigger) ([]*qdrant.ScoredPoint, []float32, error) {
	// Clean trigger for better embedding generation
	cleanText, err := util.CleanTriggerForEmbedding(trigger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to clean trigger: %w", err)
	}
	triggerEmbeddings, err := n.tsClient.TextToEmbeddings(ctx, cleanText)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create trigger embedding: %w", err)
	}
	if len(triggerEmbeddings) != 1 {
		return nil, nil, fmt.Errorf("invalid number of embeddings returned")
	}

	triggerEmbedding := triggerEmbeddings[0]
	triggerResults, err := n.queryQdrant(ctx, triggerEmbedding)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get pages: %w", err)
	}
	return triggerResults, triggerEmbedding, nil
}

func convertResultsToRelevantPages(searchResults []*qdrant.ScoredPoint, minScore float32) []model.Page {
	pages := make([]model.Page, 0)
	for _, result := range searchResults {
		if result.Score < minScore {
			break // All remaining results will be below threshold (sorted)
		}

		// Extract page data from Qdrant Value
		if dataValue, ok := result.Payload["page"]; ok {
			if structValue := dataValue.GetStructValue(); structValue != nil {
				// Extract actual values from protobuf structure
				fields := structValue.GetFields()

				page := model.Page{}

				if idVal, exists := fields["id"]; exists {
					if stringVal := idVal.GetStringValue(); stringVal != "" {
						page.Id = stringVal
					}
				}

				if layoutVal, exists := fields["layout"]; exists {
					if stringVal := layoutVal.GetStringValue(); stringVal != "" {
						page.Layout = stringVal
					}
				}

				if typeVal, exists := fields["type"]; exists {
					if stringVal := typeVal.GetStringValue(); stringVal != "" {
						page.Type = stringVal
					}
				}

				if categoryVal, exists := fields["category"]; exists {
					if stringVal := categoryVal.GetStringValue(); stringVal != "" {
						page.Category = stringVal
					}
				}

				if titleVal, exists := fields["title"]; exists {
					if listVal := titleVal.GetListValue(); listVal != nil {
						for _, val := range listVal.GetValues() {
							if stringVal := val.GetStringValue(); stringVal != "" {
								page.Title = append(page.Title, stringVal)
							}
						}
					}
				}

				if subTitleVal, exists := fields["subTitle"]; exists {
					if listVal := subTitleVal.GetListValue(); listVal != nil {
						for _, val := range listVal.GetValues() {
							if stringVal := val.GetStringValue(); stringVal != "" {
								page.SubTitle = append(page.SubTitle, stringVal)
							}
						}
					}
				}

				pages = append(pages, page)
			}
		}
	}

	return pages
}
