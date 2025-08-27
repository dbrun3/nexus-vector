package model

import "math/rand/v2"

type Page struct {
	Id       string   `json:"id"`
	Layout   string   `json:"layout"`
	Type     string   `json:"type"`
	Category string   `json:"category"`
	Title    []string `json:"title"`
	SubTitle []string `json:"subTitle"`
}

// CreateRandomPage generates a randomized Page with predefined values
func CreateRandomPage(seed uint64) Page {
	rng := rand.New(rand.NewPCG(seed, seed))

	// Predefined value lists
	layouts := []string{"card", "banner", "list", "grid", "carousel", "modal"}
	types := []string{"offer", "reward", "recommendation", "notification", "promotion", "survey"}
	categories := []string{"groceries", "electronics", "clothing", "restaurants", "beauty", "home", "automotive", "health", "books", "sports"}

	titles := []string{
		"Exclusive Offer!", "Limited Time Deal", "Special Reward", "Just For You",
		"Don't Miss Out", "Trending Now", "Popular Choice", "Best Value",
		"New Arrival", "Customer Favorite", "Flash Sale", "Premium Selection",
	}

	subtitles := []string{
		"Save big on your favorites", "Earn extra rewards", "Top-rated products",
		"Based on your preferences", "While supplies last", "Free shipping included",
		"Member exclusive", "Highly recommended", "Perfect for you", "Great value",
		"Limited quantity", "Act fast", "Don't wait", "Available now",
	}

	// Generate random titles (1-3 titles)
	titleCount := rng.IntN(3) + 1
	selectedTitles := make([]string, 0, titleCount)
	usedTitles := make(map[string]bool)

	for len(selectedTitles) < titleCount {
		title := titles[rng.IntN(len(titles))]
		if !usedTitles[title] {
			selectedTitles = append(selectedTitles, title)
			usedTitles[title] = true
		}
	}

	// Generate random subtitles (1-2 subtitles)
	subtitleCount := rng.IntN(2) + 1
	selectedSubtitles := make([]string, 0, subtitleCount)
	usedSubtitles := make(map[string]bool)

	for len(selectedSubtitles) < subtitleCount {
		subtitle := subtitles[rng.IntN(len(subtitles))]
		if !usedSubtitles[subtitle] {
			selectedSubtitles = append(selectedSubtitles, subtitle)
			usedSubtitles[subtitle] = true
		}
	}

	return Page{
		Layout:   layouts[rng.IntN(len(layouts))],
		Type:     types[rng.IntN(len(types))],
		Category: categories[rng.IntN(len(categories))],
		Title:    selectedTitles,
		SubTitle: selectedSubtitles,
	}
}
