package model

import (
	"math/rand/v2"
)

type TriggerType string

const (
	PostSnapTrigger     TriggerType = "snap"
	PostEreceiptTrigger TriggerType = "ereceipt"
	PostRedemption      TriggerType = "redeem"
)

type Trigger struct {
	TriggerType TriggerType `json:"trigger_type"`

	// Common fields
	Amount   float64 `json:"amount,omitempty"`
	Category string  `json:"category,omitempty"`

	// Snap & E-receipt specific fields
	Items    []PurchaseItem `json:"items,omitempty"`
	Retailer string         `json:"retailer,omitempty"`
	Location string         `json:"location,omitempty"`

	// Redemption specific fields
	GiftCardBrand   string  `json:"gift_card_brand,omitempty"`
	GiftCardType    string  `json:"gift_card_type,omitempty"` // physical, digital
	RedemptionValue float64 `json:"redemption_value,omitempty"`
}

type PurchaseItem struct {
	Name     string  `json:"name"`
	Brand    string  `json:"brand,omitempty"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

// CreateRandomTrigger generates a randomized TriggerProfile with predefined values
func CreateRandomTrigger(seed uint64) Trigger {
	rng := rand.New(rand.NewPCG(seed, seed))
	// Predefined value lists
	triggers := []TriggerType{PostSnapTrigger, PostEreceiptTrigger, PostRedemption}
	retailers := []string{"Walmart", "Target", "Kroger", "Costco", "Amazon", "Best Buy", "CVS", "Walgreens", "Home Depot", "Starbucks"}
	locations := []string{"New York, NY", "Los Angeles, CA", "Chicago, IL", "Houston, TX", "Phoenix, AZ", "Philadelphia, PA", "San Antonio, TX", "San Diego, CA", "Dallas, TX", "San Jose, CA"}
	categories := []string{"groceries", "electronics", "clothing", "restaurants", "beauty", "home", "automotive", "health", "books", "sports"}

	// Items for purchase triggers
	itemNames := []string{"Organic Bananas", "iPhone 15", "Nike Sneakers", "Coffee Beans", "Laptop Charger", "Shampoo", "Bread", "Protein Bars", "Wireless Headphones", "Pasta"}
	brands := []string{"Apple", "Nike", "Samsung", "Coca-Cola", "Pepsi", "McDonald's", "Starbucks", "Amazon", "Google", "Microsoft"}

	// Gift card data for redemption triggers
	giftCardBrands := []string{"Amazon", "Starbucks", "Target", "Best Buy", "iTunes", "Google Play", "Netflix", "Uber", "DoorDash", "Steam"}
	giftCardTypes := []string{"digital", "physical"}

	triggerType := triggers[rng.IntN(len(triggers))]

	switch triggerType {
	case PostSnapTrigger, PostEreceiptTrigger:
		// Generate random items (1-5 items)
		itemCount := rng.IntN(5) + 1
		items := make([]PurchaseItem, 0, itemCount)
		totalAmount := 0.0

		for i := 0; i < itemCount; i++ {
			price := float64(rng.IntN(10000)) / 100 // $0.01 - $99.99
			quantity := rng.IntN(3) + 1             // 1-3 items

			item := PurchaseItem{
				Name:     itemNames[rng.IntN(len(itemNames))],
				Brand:    brands[rng.IntN(len(brands))],
				Category: categories[rng.IntN(len(categories))],
				Price:    price,
				Quantity: quantity,
			}

			items = append(items, item)
			totalAmount += price * float64(quantity)
		}

		return Trigger{
			TriggerType: triggerType,
			Amount:      totalAmount,
			Items:       items,
			Retailer:    retailers[rng.IntN(len(retailers))],
			Location:    locations[rng.IntN(len(locations))],
		}

	case PostRedemption:
		redemptionValue := float64([]int{5, 10, 15, 20, 25, 50, 100}[rng.IntN(7)]) // Common gift card values

		return Trigger{
			TriggerType:     PostRedemption,
			GiftCardBrand:   giftCardBrands[rng.IntN(len(giftCardBrands))],
			GiftCardType:    giftCardTypes[rng.IntN(len(giftCardTypes))],
			RedemptionValue: redemptionValue,
			Category:        categories[rng.IntN(len(categories))],
		}

	default:
		// Fallback to snap trigger
		return Trigger{
			TriggerType: PostSnapTrigger,
		}
	}
}

// CreateSimilarTrigger creates a new trigger by randomly modifying only one field from the input
func CreateSimilarTrigger(seed uint64, original Trigger) Trigger {
	rng := rand.New(rand.NewPCG(seed, seed))

	// Predefined value lists (same as CreateRandomTrigger)
	triggers := []TriggerType{PostSnapTrigger, PostEreceiptTrigger, PostRedemption}
	retailers := []string{"Walmart", "Target", "Kroger", "Costco", "Amazon", "Best Buy", "CVS", "Walgreens", "Home Depot", "Starbucks"}
	locations := []string{"New York, NY", "Los Angeles, CA", "Chicago, IL", "Houston, TX", "Phoenix, AZ", "Philadelphia, PA", "San Antonio, TX", "San Diego, CA", "Dallas, TX", "San Jose, CA"}
	categories := []string{"groceries", "electronics", "clothing", "restaurants", "beauty", "home", "automotive", "health", "books", "sports"}
	itemNames := []string{"Organic Bananas", "iPhone 15", "Nike Sneakers", "Coffee Beans", "Laptop Charger", "Shampoo", "Bread", "Protein Bars", "Wireless Headphones", "Pasta"}
	brands := []string{"Apple", "Nike", "Samsung", "Coca-Cola", "Pepsi", "McDonald's", "Starbucks", "Amazon", "Google", "Microsoft"}
	giftCardBrands := []string{"Amazon", "Starbucks", "Target", "Best Buy", "iTunes", "Google Play", "Netflix", "Uber", "DoorDash", "Steam"}
	giftCardTypes := []string{"digital", "physical"}

	// Create a copy of the original
	modified := original

	// Choose which field to modify based on trigger type
	var fieldOptions []int

	// Common fields available for all trigger types
	fieldOptions = append(fieldOptions, 0, 1) // TriggerType, Amount

	if original.Category != "" {
		fieldOptions = append(fieldOptions, 2) // Category
	}

	// Type-specific fields
	switch original.TriggerType {
	case PostSnapTrigger, PostEreceiptTrigger:
		if len(original.Items) > 0 {
			fieldOptions = append(fieldOptions, 3) // Items
		}
		if original.Retailer != "" {
			fieldOptions = append(fieldOptions, 4) // Retailer
		}
		if original.Location != "" {
			fieldOptions = append(fieldOptions, 5) // Location
		}
	case PostRedemption:
		if original.GiftCardBrand != "" {
			fieldOptions = append(fieldOptions, 6) // GiftCardBrand
		}
		if original.GiftCardType != "" {
			fieldOptions = append(fieldOptions, 7) // GiftCardType
		}
		if original.RedemptionValue != 0 {
			fieldOptions = append(fieldOptions, 8) // RedemptionValue
		}
	}

	// Randomly select a field to modify
	fieldToModify := fieldOptions[rng.IntN(len(fieldOptions))]

	switch fieldToModify {
	case 0: // TriggerType
		newType := triggers[rng.IntN(len(triggers))]
		modified.TriggerType = newType

		// Adjust fields based on new trigger type
		switch newType {
		case PostSnapTrigger, PostEreceiptTrigger:
			if len(original.Items) == 0 {
				// Generate random items if switching to purchase trigger
				itemCount := rng.IntN(3) + 1
				items := make([]PurchaseItem, 0, itemCount)
				totalAmount := 0.0

				for range itemCount {
					price := float64(rng.IntN(10000)) / 100
					quantity := rng.IntN(3) + 1

					item := PurchaseItem{
						Name:     itemNames[rng.IntN(len(itemNames))],
						Brand:    brands[rng.IntN(len(brands))],
						Category: categories[rng.IntN(len(categories))],
						Price:    price,
						Quantity: quantity,
					}

					items = append(items, item)
					totalAmount += price * float64(quantity)
				}

				modified.Items = items
				modified.Amount = totalAmount
				modified.Retailer = retailers[rng.IntN(len(retailers))]
				modified.Location = locations[rng.IntN(len(locations))]
			}
		case PostRedemption:
			if original.GiftCardBrand == "" {
				modified.GiftCardBrand = giftCardBrands[rng.IntN(len(giftCardBrands))]
				modified.GiftCardType = giftCardTypes[rng.IntN(len(giftCardTypes))]
				modified.RedemptionValue = float64([]int{5, 10, 15, 20, 25, 50, 100}[rng.IntN(7)])
			}
		}

	case 1: // Amount
		if original.TriggerType == PostRedemption {
			modified.RedemptionValue = float64([]int{5, 10, 15, 20, 25, 50, 100}[rng.IntN(7)])
		} else {
			modified.Amount = float64(rng.IntN(50000)) / 100
		}

	case 2: // Category
		modified.Category = categories[rng.IntN(len(categories))]

	case 3: // Items
		itemCount := rng.IntN(5) + 1
		items := make([]PurchaseItem, 0, itemCount)
		totalAmount := 0.0

		for range itemCount {
			price := float64(rng.IntN(10000)) / 100
			quantity := rng.IntN(3) + 1

			item := PurchaseItem{
				Name:     itemNames[rng.IntN(len(itemNames))],
				Brand:    brands[rng.IntN(len(brands))],
				Category: categories[rng.IntN(len(categories))],
				Price:    price,
				Quantity: quantity,
			}

			items = append(items, item)
			totalAmount += price * float64(quantity)
		}

		modified.Items = items
		modified.Amount = totalAmount

	case 4: // Retailer
		modified.Retailer = retailers[rng.IntN(len(retailers))]

	case 5: // Location
		modified.Location = locations[rng.IntN(len(locations))]

	case 6: // GiftCardBrand
		modified.GiftCardBrand = giftCardBrands[rng.IntN(len(giftCardBrands))]

	case 7: // GiftCardType
		modified.GiftCardType = giftCardTypes[rng.IntN(len(giftCardTypes))]

	case 8: // RedemptionValue
		modified.RedemptionValue = float64([]int{5, 10, 15, 20, 25, 50, 100}[rng.IntN(7)])
	}

	return modified
}
