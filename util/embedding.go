package util

import (
	"encoding/json"
	"strings"

	"github.com/dbrun3/nexus-vector/model"
)

// cleanJSONText efficiently cleans JSON text in a single pass
func cleanJSONText(jsonBytes []byte) string {
	var result strings.Builder
	result.Grow(len(jsonBytes)) // Pre-allocate capacity

	for _, b := range jsonBytes {
		switch b {
		case '{', '}', '[', ']', '"':
			// Skip these characters
			continue
		case ':', ',':
			// Replace with space
			result.WriteByte(' ')
		default:
			// Keep the character
			result.WriteByte(b)
		}
	}

	// Clean up multiple spaces
	return strings.Join(strings.Fields(result.String()), " ")
}

// CleanUserSnapshotForEmbedding converts a UserSnapshot to a clean string format for embedding generation
func CleanUserSnapshotForEmbedding(snapshot model.UserSnapshot) (string, error) {
	jsonBytes, err := json.Marshal(snapshot)
	if err != nil {
		return "", err
	}

	return cleanJSONText(jsonBytes), nil
}

// CleanTriggerForEmbedding converts a Trigger to a clean string format for embedding generation
func CleanTriggerForEmbedding(trigger model.Trigger) (string, error) {
	simplifiedTrigger := SimplifyTrigger(trigger)

	jsonBytes, err := json.Marshal(simplifiedTrigger)
	if err != nil {
		return "", err
	}

	return cleanJSONText(jsonBytes), nil
}

// SimplifiedPurchaseItem represents a simplified purchase item for embedding
type SimplifiedPurchaseItem struct {
	Item string `json:"item"`
}

// SimplifiedTrigger represents a simplified trigger for embedding generation
type SimplifiedTrigger struct {
	TriggerType model.TriggerType        `json:"trigger_type"`
	Amount      float64                  `json:"amount,omitempty"`
	Category    string                   `json:"category,omitempty"`
	Items       []SimplifiedPurchaseItem `json:"items,omitempty"`
	Retailer    string                   `json:"retailer,omitempty"`
	Location    string                   `json:"location,omitempty"`

	// Redemption specific fields
	GiftCardBrand   string  `json:"gift_card_brand,omitempty"`
	GiftCardType    string  `json:"gift_card_type,omitempty"`
	RedemptionValue float64 `json:"redemption_value,omitempty"`
}

// SimplifyTrigger converts a full Trigger to a SimplifiedTrigger for embedding
func SimplifyTrigger(trigger model.Trigger) SimplifiedTrigger {
	simplified := SimplifiedTrigger{
		TriggerType:     trigger.TriggerType,
		Amount:          trigger.Amount,
		Category:        trigger.Category,
		Retailer:        trigger.Retailer,
		Location:        trigger.Location,
		GiftCardBrand:   trigger.GiftCardBrand,
		GiftCardType:    trigger.GiftCardType,
		RedemptionValue: trigger.RedemptionValue,
	}

	// Simplify purchase items - combine brand and name, drop other fields
	if len(trigger.Items) > 0 {
		simplified.Items = make([]SimplifiedPurchaseItem, len(trigger.Items))
		for i, item := range trigger.Items {
			// Combine brand and name into a single item string
			itemString := item.Name
			if item.Brand != "" && item.Brand != item.Name {
				itemString = item.Brand + " " + item.Name
			}
			simplified.Items[i] = SimplifiedPurchaseItem{
				Item: itemString,
			}
		}
	}

	return simplified
}
