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
	// Convert to semantic representation for better embeddings
	semanticSnapshot := convertUserSnapshotToSemantic(snapshot)

	jsonBytes, err := json.Marshal(semanticSnapshot)
	if err != nil {
		return "", err
	}

	return cleanJSONText(jsonBytes), nil
}

// SemanticUserSnapshot represents a user snapshot with semantic descriptions instead of raw numbers
type SemanticUserSnapshot struct {
	ID                   string   `json:"id,omitempty"`
	Gender               string   `json:"gender,omitempty"`
	AgeGroup             string   `json:"age_group,omitempty"`
	Location             string   `json:"location,omitempty"`
	RewardsLevel         string   `json:"rewards_level,omitempty"`
	SpendingLevel        string   `json:"spending_level,omitempty"`
	LastPurchaseCategory string   `json:"last_purchase_category,omitempty"`
	FavoriteCategories   []string `json:"favorite_categories,omitempty"`
	EngagementLevel      string   `json:"engagement_level,omitempty"`
	AppUsageFrequency    string   `json:"app_usage_frequency,omitempty"`
	PreferredOfferType   string   `json:"preferred_offer_type,omitempty"`
	SeasonalPreference   string   `json:"seasonal_preference,omitempty"`
	ShoppingTimePref     string   `json:"shopping_time_pref,omitempty"`
	Pricesensitivity     string   `json:"price_sensitivity,omitempty"`
	BrandLoyalty         string   `json:"brand_loyalty,omitempty"`
}

// convertUserSnapshotToSemantic converts numeric values to semantic descriptions
func convertUserSnapshotToSemantic(snapshot model.UserSnapshot) SemanticUserSnapshot {
	return SemanticUserSnapshot{
		ID:                   snapshot.ID,
		Gender:               snapshot.Gender,
		AgeGroup:             ageToSemantic(snapshot.Age),
		Location:             snapshot.Location,
		RewardsLevel:         rewardsToSemantic(snapshot.RewardsBalance),
		SpendingLevel:        spendToSemantic(snapshot.TotalSpend),
		LastPurchaseCategory: snapshot.LastPurchaseCategory,
		FavoriteCategories:   snapshot.FavoriteCategories,
		EngagementLevel:      snapshot.EngagementLevel,
		AppUsageFrequency:    snapshot.AppUsageFrequency,
		PreferredOfferType:   snapshot.PreferredOfferType,
		SeasonalPreference:   snapshot.SeasonalPreference,
		ShoppingTimePref:     snapshot.ShoppingTimePref,
		Pricesensitivity:     snapshot.Pricesensitivity,
		BrandLoyalty:         snapshot.BrandLoyalty,
	}
}

// ageToSemantic converts numeric age to semantic age groups
func ageToSemantic(age int) string {
	switch {
	case age < 18:
		return "minor"
	case age >= 18 && age <= 24:
		return "young adult"
	case age >= 25 && age <= 34:
		return "adult"
	case age >= 35 && age <= 44:
		return "middle aged"
	case age >= 45 && age <= 54:
		return "mature adult"
	case age >= 55 && age <= 64:
		return "senior"
	case age >= 65:
		return "elderly"
	default:
		return "unknown age"
	}
}

// rewardsToSemantic converts numeric rewards balance to semantic levels
func rewardsToSemantic(balance int) string {
	switch {
	case balance == 0:
		return "no rewards"
	case balance > 0 && balance <= 500:
		return "low rewards"
	case balance > 500 && balance <= 2000:
		return "moderate rewards"
	case balance > 2000 && balance <= 5000:
		return "high rewards"
	case balance > 5000:
		return "premium rewards"
	default:
		return "unknown rewards"
	}
}

// spendToSemantic converts numeric spending to semantic levels
func spendToSemantic(spend float64) string {
	switch {
	case spend == 0:
		return "no spending"
	case spend > 0 && spend <= 500:
		return "low spender"
	case spend > 500 && spend <= 2000:
		return "moderate spender"
	case spend > 2000 && spend <= 10000:
		return "high spender"
	case spend > 10000:
		return "premium spender"
	default:
		return "unknown spending"
	}
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
