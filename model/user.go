package model

import (
	"math/rand/v2"

	"github.com/google/uuid"
)

type UserSnapshot struct {
	ID                   string   `json:"id,omitempty"`
	Gender               string   `json:"gender,omitempty"`
	Age                  int      `json:"age,omitempty"`
	Location             string   `json:"location,omitempty"`
	RewardsBalance       int      `json:"rewards_balance,omitempty"`
	TotalSpend           float64  `json:"total_spend,omitempty"`
	LastPurchaseCategory string   `json:"last_purchase_category,omitempty"`
	FavoriteCategories   []string `json:"favorite_categories,omitempty"`
	EngagementLevel      string   `json:"engagement_level,omitempty"`     // high, medium, low
	AppUsageFrequency    string   `json:"app_usage_frequency,omitempty"`  // daily, weekly, monthly, rare
	PreferredOfferType   string   `json:"preferred_offer_type,omitempty"` // cashback, discount, freebie
	SeasonalPreference   string   `json:"seasonal_preference,omitempty"`  // spring, summer, fall, winter
	ShoppingTimePref     string   `json:"shopping_time_pref,omitempty"`   // morning, afternoon, evening, night
	Pricesensitivity     string   `json:"price_sensitivity,omitempty"`    // high, medium, low
	BrandLoyalty         string   `json:"brand_loyalty,omitempty"`        // high, medium, low
}

// CreateRandomSnapshot generates a randomized UserSnapshot with predefined values
func CreateRandomSnapshot(seed uint64) UserSnapshot {
	rng := rand.New(rand.NewPCG(seed, seed))

	// Predefined value lists
	genders := []string{"male", "female", "non-binary", "prefer-not-to-say"}
	locations := []string{"urban", "suburban", "rural"}
	categories := []string{"groceries", "electronics", "clothing", "restaurants", "beauty", "home", "automotive", "health", "books", "sports"}
	engagementLevels := []string{"high", "medium", "low"}
	frequencies := []string{"daily", "weekly", "monthly", "rare"}
	offerTypes := []string{"cashback", "discount", "freebie"}
	seasons := []string{"spring", "summer", "fall", "winter"}
	timePrefs := []string{"morning", "afternoon", "evening", "night"}
	sensitivities := []string{"high", "medium", "low"}

	// Generate random favorite categories (1-4 categories)
	favCatCount := rng.IntN(4) + 1
	favoriteCategories := make([]string, 0, favCatCount)
	usedCategories := make(map[string]bool)

	for len(favoriteCategories) < favCatCount {
		category := categories[rng.IntN(len(categories))]
		if !usedCategories[category] {
			favoriteCategories = append(favoriteCategories, category)
			usedCategories[category] = true
		}
	}

	return UserSnapshot{
		ID:                   uuid.New().String(),
		Gender:               genders[rng.IntN(len(genders))],
		Age:                  rng.IntN(65) + 18, // Age between 18-82
		Location:             locations[rng.IntN(len(locations))],
		RewardsBalance:       rng.IntN(10000),                              // 0-9999 points
		TotalSpend:           float64(rng.IntN(50000)) + rng.Float64()*100, // $0-$50,099.xx
		LastPurchaseCategory: categories[rng.IntN(len(categories))],
		FavoriteCategories:   favoriteCategories,
		EngagementLevel:      engagementLevels[rng.IntN(len(engagementLevels))],
		AppUsageFrequency:    frequencies[rng.IntN(len(frequencies))],
		PreferredOfferType:   offerTypes[rng.IntN(len(offerTypes))],
		SeasonalPreference:   seasons[rng.IntN(len(seasons))],
		ShoppingTimePref:     timePrefs[rng.IntN(len(timePrefs))],
		Pricesensitivity:     sensitivities[rng.IntN(len(sensitivities))],
		BrandLoyalty:         sensitivities[rng.IntN(len(sensitivities))], // reuse same levels
	}
}

// CreateSimilarSnapshot creates a new snapshot by randomly modifying only one field from the input
func CreateSimilarSnapshot(seed uint64, original UserSnapshot) UserSnapshot {
	rng := rand.New(rand.NewPCG(seed, seed))

	// Predefined value lists (same as CreateRandomSnapshot)
	genders := []string{"male", "female", "non-binary", "prefer-not-to-say"}
	locations := []string{"urban", "suburban", "rural"}
	categories := []string{"groceries", "electronics", "clothing", "restaurants", "beauty", "home", "automotive", "health", "books", "sports"}
	engagementLevels := []string{"high", "medium", "low"}
	frequencies := []string{"daily", "weekly", "monthly", "rare"}
	offerTypes := []string{"cashback", "discount", "freebie"}
	seasons := []string{"spring", "summer", "fall", "winter"}
	timePrefs := []string{"morning", "afternoon", "evening", "night"}
	sensitivities := []string{"high", "medium", "low"}

	// Create a copy of the original
	modified := original
	modified.ID = uuid.New().String() // Always assign new ID

	// Choose which field to modify (0-12, excluding ID which is already modified)
	fieldToModify := rng.IntN(12)

	switch fieldToModify {
	case 0: // Gender
		modified.Gender = genders[rng.IntN(len(genders))]
	case 1: // Age
		modified.Age = rng.IntN(65) + 18
	case 2: // Location
		modified.Location = locations[rng.IntN(len(locations))]
	case 3: // RewardsBalance
		modified.RewardsBalance = rng.IntN(10000)
	case 4: // TotalSpend
		modified.TotalSpend = float64(rng.IntN(50000)) + rng.Float64()*100
	case 5: // LastPurchaseCategory
		modified.LastPurchaseCategory = categories[rng.IntN(len(categories))]
	case 6: // FavoriteCategories
		favCatCount := rng.IntN(4) + 1
		favoriteCategories := make([]string, 0, favCatCount)
		usedCategories := make(map[string]bool)
		for len(favoriteCategories) < favCatCount {
			category := categories[rng.IntN(len(categories))]
			if !usedCategories[category] {
				favoriteCategories = append(favoriteCategories, category)
				usedCategories[category] = true
			}
		}
		modified.FavoriteCategories = favoriteCategories
	case 7: // EngagementLevel
		modified.EngagementLevel = engagementLevels[rng.IntN(len(engagementLevels))]
	case 8: // AppUsageFrequency
		modified.AppUsageFrequency = frequencies[rng.IntN(len(frequencies))]
	case 9: // PreferredOfferType
		modified.PreferredOfferType = offerTypes[rng.IntN(len(offerTypes))]
	case 10: // SeasonalPreference
		modified.SeasonalPreference = seasons[rng.IntN(len(seasons))]
	case 11: // ShoppingTimePref
		modified.ShoppingTimePref = timePrefs[rng.IntN(len(timePrefs))]
	case 12: // Pricesensitivity
		modified.Pricesensitivity = sensitivities[rng.IntN(len(sensitivities))]
	case 13: // BrandLoyalty
		modified.BrandLoyalty = sensitivities[rng.IntN(len(sensitivities))]
	}

	return modified
}
