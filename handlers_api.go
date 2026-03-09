// File purpose:
// This file contains all the JSON API handler functions for the application.
// These handlers respond to HTTP requests with JSON data instead of HTML pages.
// They are used by JavaScript on the frontend to fetch car data dynamically.
//
// Responsibilities:
// - Return lists of car models, manufacturers, and categories as JSON
// - Return a single item by its ID as JSON
// - Handle car search with multiple filters and sorting options
// - Handle car comparison (side-by-side data for multiple cars)
// - Handle car recommendations based on category and horsepower preferences
//
// Used by:
// - main.go registers all these handler functions for their respective URL routes
// - Frontend JavaScript calls these endpoints to update the page without full reloads

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// encoding/json is used to encode Go structs as JSON and write them to HTTP responses.

// fmt is used to build formatted error messages (e.g., "service error: <message>").

// math is used for math.Max to ensure recommendation scores never go below 0.1.

// net/http provides the HTTP handler types, status code constants, and utility functions.

// sort is used to sort search results and recommendation results by various criteria.

// strconv is used to convert URL query parameters (which are strings) to integers.

// strings is used for case-insensitive string comparisons and URL path manipulation.

// jsonResponse writes a Go value as a JSON HTTP response.
// It sets the Content-Type header to "application/json" before encoding.
//
// Input:
// - w: the HTTP response writer
// - data: any Go value to encode as JSON (struct, slice, map, etc.)
//
// Side effects:
// - Sets Content-Type header to "application/json"
// - Writes the JSON-encoded data to the response body
func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// requireGET checks that the incoming HTTP request uses the GET method.
// If it is not a GET request, it responds with "405 Method Not Allowed"
// and returns false so the caller knows to stop processing.
//
// Input:
// - w: the HTTP response writer
// - r: the incoming HTTP request
//
// Output:
// - Returns true if the method is GET (handler should continue)
// - Returns false if the method is not GET (handler should stop)
//
// Why validate the HTTP method?
// All our API endpoints are read-only (they do not change data),
// so we only allow GET requests. POST, PUT, DELETE etc. are rejected.
func requireGET(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

// apiModels handles GET /api/models
// Returns a JSON array of all car models in the database.
//
// Input:
// - w: the HTTP response writer
// - r: the incoming HTTP request
//
// Output:
// - 200 OK with a JSON array of all CarModel objects
// - 405 Method Not Allowed if the request is not a GET
func apiModels(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}

	jsonResponse(w, db.CarModels)
}

// apiModelByID handles GET /api/models/{id}
// Returns a single car model by its ID, enriched with manufacturer and category data.
//
// Input:
// - w: the HTTP response writer
// - r: the incoming HTTP request (the car ID is extracted from the URL path)
//
// Output:
// - 200 OK with a JSON object containing the car model, its manufacturer, and its category
// - 400 Bad Request if the ID in the URL is not a valid integer
// - 404 Not Found if no car with that ID exists
// - 405 Method Not Allowed if the request is not a GET
//
// Side effects:
// - Calls recordView to increment the view counter for the car asynchronously
func apiModelByID(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}

	// Extract the car ID from the URL path by removing the known prefix.
	// For example, "/api/models/42" becomes "42".
	idStr := strings.TrimPrefix(r.URL.Path, "/api/models/")

	// Validate that the ID is a valid integer before looking it up.
	// This avoids unnecessary database lookups with invalid input.
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	// DetailedCar embeds CarModel and adds manufacturer and category fields.
	// This gives the API caller all the information they need in one response
	// without having to make multiple API calls.
	type DetailedCar struct {
		CarModel
		Manufacturer *Manufacturer `json:"manufacturer"`
		Category     *Category     `json:"category"`
	}
	car := carByID(id)

	// Edge case: if the car ID does not exist, return 404 instead of crashing.
	if car == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Record the view asynchronously so this handler does not have to wait.
	// The "api:model_detail" source label helps us track where views come from.
	recordView(id, "api:model_detail")
	jsonResponse(w, DetailedCar{
		CarModel:     *car,
		Manufacturer: manufacturerByID(car.ManufacturerID),
		Category:     categoryByID(car.CategoryID),
	})
}

// apiManufacturers handles GET /api/manufacturers
// Returns a JSON array of all manufacturers in the database.
//
// Input:
// - w: the HTTP response writer
// - r: the incoming HTTP request
//
// Output:
// - 200 OK with a JSON array of all Manufacturer objects
// - 405 Method Not Allowed if the request is not a GET
func apiManufacturers(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}

	jsonResponse(w, db.Manufacturers)
}

// apiManufacturerByID handles GET /api/manufacturers/{id}
// Returns a single manufacturer by its ID, along with all car models by that manufacturer.
//
// Input:
// - w: the HTTP response writer
// - r: the incoming HTTP request (the manufacturer ID is extracted from the URL path)
//
// Output:
// - 200 OK with a JSON object containing the manufacturer and its list of car models
// - 400 Bad Request if the ID in the URL is not a valid integer
// - 404 Not Found if no manufacturer with that ID exists
// - 405 Method Not Allowed if the request is not a GET
func apiManufacturerByID(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}

	// Extract the manufacturer ID from the URL path.
	idStr := strings.TrimPrefix(r.URL.Path, "/api/manufacturers/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	m := manufacturerByID(id)

	// Edge case: return 404 if the manufacturer is not found.
	if m == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Build the list of car models that belong to this manufacturer.
	// We scan all cars and keep only those with a matching ManufacturerID.
	var models []CarModel
	for _, c := range db.CarModels {
		if c.ManufacturerID == id {
			models = append(models, c)
		}
	}
	jsonResponse(w, map[string]interface{}{"manufacturer": m, "models": models})
}

// apiCategories handles GET /api/categories
// Returns a JSON array of all car categories in the database.
//
// Input:
// - w: the HTTP response writer
// - r: the incoming HTTP request
//
// Output:
// - 200 OK with a JSON array of all Category objects
// - 405 Method Not Allowed if the request is not a GET
func apiCategories(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}

	jsonResponse(w, db.Categories)
}

// apiCategoryByID handles GET /api/categories/{id}
// Returns a single category by its ID, along with all car models in that category.
//
// Input:
// - w: the HTTP response writer
// - r: the incoming HTTP request (the category ID is extracted from the URL path)
//
// Output:
// - 200 OK with a JSON object containing the category and its list of car models
// - 400 Bad Request if the ID in the URL is not a valid integer
// - 404 Not Found if no category with that ID exists
// - 405 Method Not Allowed if the request is not a GET
func apiCategoryByID(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}

	// Extract the category ID from the URL path.
	idStr := strings.TrimPrefix(r.URL.Path, "/api/categories/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	cat := categoryByID(id)

	// Edge case: return 404 if the category is not found.
	if cat == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Build the list of car models that belong to this category.
	var models []CarModel
	for _, c := range db.CarModels {
		if c.CategoryID == id {
			models = append(models, c)
		}
	}
	jsonResponse(w, map[string]interface{}{"category": cat, "models": models})
}

// apiCompare handles GET /api/compare?ids=1,2,3
// Returns detailed data for multiple cars so they can be compared side by side.
//
// Input:
// - w: the HTTP response writer
// - r: the incoming HTTP request
//   Query parameters:
//   - ids: one or more comma-separated car IDs (e.g., ?ids=1,2 or ?ids=1&ids=2)
//
// Output:
// - 200 OK with a JSON array of CompareEntry objects (car + manufacturer + category)
// - 400 Bad Request if the ids parameter is missing
// - 405 Method Not Allowed if the request is not a GET
//
// Edge cases:
// - Invalid IDs (non-integer) are silently skipped
// - IDs that do not exist are silently skipped
// - Returns an empty array if no valid IDs were provided
func apiCompare(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}

	// Read all values of the "ids" query parameter.
	// The caller can pass IDs as ?ids=1,2,3 or as ?ids=1&ids=2&ids=3 (or both).
	idsValues := r.URL.Query()["ids"]
	if len(idsValues) == 0 {
		http.Error(w, "ids param required", http.StatusBadRequest)
		return
	}

	// CompareEntry enriches a car with its manufacturer and category data
	// so the frontend can display all information without extra API calls.
	type CompareEntry struct {
		CarModel
		Manufacturer *Manufacturer `json:"manufacturer"`
		Category     *Category     `json:"category"`
	}
	var result []CompareEntry

	// Parse each value of the "ids" parameter.
	// Each value may contain multiple comma-separated IDs (e.g., "1,2,3").
	for _, raw := range idsValues {
		for _, p := range strings.Split(raw, ",") {
			id, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil {
				// Skip non-integer values silently (e.g., "abc").
				continue
			}
			car := carByID(id)
			if car == nil {
				// Skip IDs that do not match any car silently.
				continue
			}
			result = append(result, CompareEntry{
				CarModel:     *car,
				Manufacturer: manufacturerByID(car.ManufacturerID),
				Category:     categoryByID(car.CategoryID),
			})
		}
	}
	jsonResponse(w, result)
}

// apiRecommendations handles GET /api/recommendations
// Returns up to 4 recommended cars based on view popularity,
// category preference, and horsepower range.
//
// Input:
// - w: the HTTP response writer
// - r: the incoming HTTP request
//   Optional query parameters:
//   - category: filter by category name (case-insensitive, e.g., "SUV")
//   - minHP: minimum horsepower (e.g., "200")
//   - maxHP: maximum horsepower (e.g., "400")
//
// Output:
// - 200 OK with a JSON array of up to 4 ScoredCar objects,
//   sorted by score from highest to lowest
// - 405 Method Not Allowed if the request is not a GET
//
// How scoring works:
// - Cars with more views get a higher score (popularity bonus)
// - Cars matching the requested category get +10 points
// - Cars meeting the minimum HP requirement get +5 points
// - Cars within the maximum HP limit get +3 points
// - Newer cars (after 2020) get a small bonus (0.5 points per year)
// - The minimum score is always 0.1 to keep all cars eligible
func apiRecommendations(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}

	categoryStr := r.URL.Query().Get("category")
	minHPStr := r.URL.Query().Get("minHP")
	maxHPStr := r.URL.Query().Get("maxHP")

	// ScoredCar wraps a car with a numeric score and a human-readable reason.
	// The score determines the order of recommendations.
	// The reason is shown to the user to explain why this car was recommended.
	type ScoredCar struct {
		Car    CarModel `json:"car"`
		Score  float64  `json:"score"`
		Reason string   `json:"reason"`
	}

	// This operation is relatively expensive because it reads all view counts.
	// We take a snapshot once here instead of calling Snapshot() inside the loop
	// to avoid repeated blocking calls.
	countsSnapshot := viewCounts.Snapshot()

	var scored []ScoredCar
	for _, car := range db.CarModels {
		score := 0.0
		reasons := []string{}

		// Award points based on how many times this car has been viewed.
		// Each view adds 2 points. Popular cars float to the top.
		if vc, ok := countsSnapshot[car.ID]; ok {
			score += float64(vc) * 2
			if vc > 0 {
				reasons = append(reasons, "Popular choice")
			}
		}

		// Award +10 points if the car matches the requested category.
		if categoryStr != "" {
			cat := categoryByID(car.CategoryID)
			if cat != nil && strings.EqualFold(cat.Name, categoryStr) {
				score += 10
				reasons = append(reasons, fmt.Sprintf("Matches %s category", cat.Name))
			}
		}

		// Award +5 points if the car meets the minimum horsepower requirement.
		if minHPStr != "" {
			minHP, _ := strconv.Atoi(minHPStr)
			if car.Specifications.Horsepower >= minHP {
				score += 5
				reasons = append(reasons, fmt.Sprintf("%d HP meets minimum", car.Specifications.Horsepower))
			}
		}
		// Award +3 points if the car is within the maximum horsepower limit.
		if maxHPStr != "" {
			maxHP, _ := strconv.Atoi(maxHPStr)
			if car.Specifications.Horsepower <= maxHP {
				score += 3
			}
		}

		// Award a small recency bonus for cars made after 2020.
		// Each year after 2020 adds 0.5 points (e.g., 2023 = +1.5 points).
		score += float64(car.Year-2020) * 0.5

		// Ensure the score never goes below 0.1 so all cars remain eligible.
		// math.Max is used here instead of an if statement for brevity.
		score = math.Max(score, 0.1)

		reason := "Based on your preferences"
		if len(reasons) > 0 {
			reason = strings.Join(reasons, " · ")
		}

		scored = append(scored, ScoredCar{Car: car, Score: score, Reason: reason})
	}

	// Sort all scored cars from highest score to lowest.
	// This ensures the best recommendations appear first.
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	// Return only the top 4 recommendations to keep the response small.
	top := scored
	if len(top) > 4 {
		top = top[:4]
	}
	jsonResponse(w, top)
}

// apiSearch handles GET /api/search
// Searches for cars matching the given filters and returns a JSON array.
//
// Input:
// - w: the HTTP response writer
// - r: the incoming HTTP request
//   Optional query parameters:
//   - q: text search query (matches car name, manufacturer, category, engine, transmission)
//   - category: filter by category name (case-insensitive)
//   - manufacturer: filter by manufacturer name (case-insensitive)
//   - minHP: minimum horsepower (integer)
//   - maxHP: maximum horsepower (integer)
//   - minYear: minimum model year (integer)
//   - maxYear: maximum model year (integer)
//   - sort: sort order ("hp_desc", "hp_asc", "year_desc", "year_asc", "name")
//
// Output:
// - 200 OK with a JSON array of EnrichedCar objects matching all filters
//   (empty array if no cars match, never nil)
// - 405 Method Not Allowed if the request is not a GET
//
// Side effects:
// - Calls recordSearch to log the search event asynchronously
//   (only if at least one search filter was provided)
func apiSearch(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}

	// Read all search filter parameters from the URL query string.
	// strings.ToLower is applied to q so the text search is case-insensitive.
	q := strings.ToLower(r.URL.Query().Get("q"))
	categoryStr := r.URL.Query().Get("category")
	manufacturerStr := r.URL.Query().Get("manufacturer")
	minHPStr := r.URL.Query().Get("minHP")
	maxHPStr := r.URL.Query().Get("maxHP")
	minYearStr := r.URL.Query().Get("minYear")
	maxYearStr := r.URL.Query().Get("maxYear")
	sortBy := r.URL.Query().Get("sort")

	// Only record a search event if the user actually provided search criteria.
	// We do not record empty searches to avoid polluting the event log.
	if q != "" || categoryStr != "" || manufacturerStr != "" || minHPStr != "" || maxHPStr != "" || minYearStr != "" || maxYearStr != "" {
		recordSearch(q, "api:search")
	}

	// EnrichedCar adds manufacturer and category names to the car model.
	// This avoids extra API calls from the frontend to look up names by ID.
	type EnrichedCar struct {
		CarModel
		ManufacturerName string `json:"manufacturerName"`
		CategoryName     string `json:"categoryName"`
	}

	var results []EnrichedCar

	// Loop through all cars and apply each filter in sequence.
	// A car is only included in results if it passes ALL filters.
	for _, car := range db.CarModels {
		mfr := manufacturerByID(car.ManufacturerID)
		cat := categoryByID(car.CategoryID)
		mfrName := ""
		catName := ""
		if mfr != nil {
			mfrName = mfr.Name
		}
		if cat != nil {
			catName = cat.Name
		}

		// Text search: check if the query appears anywhere in the combined fields.
		// We search across name, manufacturer, category, engine, and transmission
		// so the user can find cars by any relevant keyword.
		if q != "" {
			haystack := strings.ToLower(car.Name + " " + mfrName + " " + catName +
				" " + car.Specifications.Engine + " " + car.Specifications.Transmission)
			if !strings.Contains(haystack, q) {
				continue
			}
		}

		// Category filter: skip cars that do not match the selected category.
		if categoryStr != "" && !strings.EqualFold(catName, categoryStr) {
			continue
		}

		// Manufacturer filter: skip cars that do not match the selected manufacturer.
		if manufacturerStr != "" && !strings.EqualFold(mfrName, manufacturerStr) {
			continue
		}

		// Horsepower range filter: skip cars outside the requested HP range.
		if minHPStr != "" {
			minHP, _ := strconv.Atoi(minHPStr)
			if car.Specifications.Horsepower < minHP {
				continue
			}
		}
		if maxHPStr != "" {
			maxHP, _ := strconv.Atoi(maxHPStr)
			if car.Specifications.Horsepower > maxHP {
				continue
			}
		}

		// Year range filter: skip cars outside the requested year range.
		if minYearStr != "" {
			minYear, _ := strconv.Atoi(minYearStr)
			if car.Year < minYear {
				continue
			}
		}
		if maxYearStr != "" {
			maxYear, _ := strconv.Atoi(maxYearStr)
			if car.Year > maxYear {
				continue
			}
		}

		results = append(results, EnrichedCar{car, mfrName, catName})
	}

	// Apply the requested sort order to the filtered results.
	// If no sort is specified, results keep their original order (by ID).
	switch sortBy {
	case "hp_desc":
		// Sort by horsepower from highest to lowest.
		sort.Slice(results, func(i, j int) bool {
			return results[i].Specifications.Horsepower > results[j].Specifications.Horsepower
		})
	case "hp_asc":
		// Sort by horsepower from lowest to highest.
		sort.Slice(results, func(i, j int) bool {
			return results[i].Specifications.Horsepower < results[j].Specifications.Horsepower
		})
	case "year_desc":
		// Sort by model year from newest to oldest.
		sort.Slice(results, func(i, j int) bool {
			return results[i].Year > results[j].Year
		})
	case "year_asc":
		// Sort by model year from oldest to newest.
		sort.Slice(results, func(i, j int) bool {
			return results[i].Year < results[j].Year
		})
	case "name":
		// Sort alphabetically by car name (A to Z).
		sort.Slice(results, func(i, j int) bool {
			return results[i].Name < results[j].Name
		})
	}

	jsonResponse(w, results)
}
