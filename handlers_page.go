// File purpose:
// This file contains the main HTML page handler for the application.
// It renders the web interface that users see in their browser.
//
// Responsibilities:
// - Handle all page views through a single URL route ("/")
// - Determine which view to render based on the "view" query parameter
// - Assemble the data needed for each view and pass it to the HTML template
//
// Views handled:
// - gallery: shows a searchable and filterable grid of cars
// - detail: shows the full details of a single car
// - compare: shows two or more cars side by side for comparison
// - recommendations: shows personalized car suggestions
// - manufacturers: shows a list of all car brands with their models
//
// Used by:
// - main.go registers indexHandler for the "/" URL route

package main

import (
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// log is used to record template rendering errors without exposing them to users.

// net/http provides types for HTTP request and response handling.

// sort is used to sort filtered car results by the user's chosen sort order.

// strconv is used to parse integer values from URL query parameters (which are strings).

// strings is used for case-insensitive comparisons, text searching, and URL parsing.

// indexHandler handles all requests to the "/" route and renders the HTML page.
// The specific view is determined by the "view" query parameter in the URL.
//
// Input:
// - w: the HTTP response writer used to send the HTML page to the browser
// - r: the incoming HTTP request (query parameters control which view is shown)
//
//   Query parameters:
//   - view: which section to show ("gallery", "detail", "compare",
//           "recommendations", "manufacturers"). Defaults to "gallery".
//   - (gallery view) q, category, manufacturer, minHP, maxHP, sort: search/filter options
//   - (detail view) id: the ID of the car to show
//   - (compare view) ids: comma-separated car IDs to compare
//   - (recommendations view) category, minHP, maxHP: preference filters
//
// Output:
// - 200 OK with the rendered HTML page
// - 500 Internal Server Error if the template fails to render
//
// Side effects:
// - Calls recordSearch when a gallery search or filter is performed
// - Calls recordView when a car detail page is viewed
func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Read the "view" parameter to decide which section to render.
	// Default to "gallery" if no view is specified (e.g., the home page).
	view := r.URL.Query().Get("view")
	if view == "" {
		view = "gallery"
	}

	// data is the map of values passed to the HTML template.
	// The template reads these values to render the correct content.
	data := map[string]interface{}{
		"View":          view,
		"Categories":    db.Categories,    // Used by filter dropdowns in the template
		"Manufacturers": db.Manufacturers, // Used by filter dropdowns in the template
	}

	// Handle gallery/search view
	if view == "gallery" || view == "" {
		// Read all search and filter parameters from the URL.
		// q is lowercased immediately so all text comparisons can be case-insensitive.
		q := strings.ToLower(r.URL.Query().Get("q"))
		categoryStr := r.URL.Query().Get("category")
		manufacturerStr := r.URL.Query().Get("manufacturer")
		minHPStr := r.URL.Query().Get("minHP")
		maxHPStr := r.URL.Query().Get("maxHP")
		sortBy := r.URL.Query().Get("sort")

		// Only record a search event if the user provided at least one filter.
		// We do not record the default empty gallery load as a "search".
		if q != "" || categoryStr != "" || manufacturerStr != "" || minHPStr != "" || maxHPStr != "" {
			recordSearch(q, "page:gallery")
		}

		// Set default HP range values so the filter always has a valid range.
		// These defaults represent the full possible range (0 to 500 HP).
		if minHPStr == "" {
			minHPStr = "0"
		}
		if maxHPStr == "" {
			maxHPStr = "500"
		}

		// EnrichedCar adds human-readable manufacturer and category names
		// to each car so the template does not need to do ID lookups.
		type EnrichedCar struct {
			CarModel
			ManufacturerName string
			CategoryName     string
		}

		var results []EnrichedCar

		// Loop through all cars and apply each filter in sequence.
		// A car is only included if it passes ALL active filters.
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

			// Text search: check if the query appears in any relevant car field.
			// We combine multiple fields into one string to search all at once.
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

			// Manufacturer filter: skip cars not from the selected brand.
			if manufacturerStr != "" && !strings.EqualFold(mfrName, manufacturerStr) {
				continue
			}

			// Horsepower range filter: skip cars outside the HP slider range.
			// We parse the values here (not earlier) because they are guaranteed
			// to have defaults set above, so Atoi will always succeed.
			minHP, _ := strconv.Atoi(minHPStr)
			if car.Specifications.Horsepower < minHP {
				continue
			}
			maxHP, _ := strconv.Atoi(maxHPStr)
			if car.Specifications.Horsepower > maxHP {
				continue
			}

			results = append(results, EnrichedCar{car, mfrName, catName})
		}

		// Apply the user's chosen sort order to the filtered results.
		switch sortBy {
		case "hp_desc":
			// Sort cars from most powerful to least powerful.
			sort.Slice(results, func(i, j int) bool {
				return results[i].Specifications.Horsepower > results[j].Specifications.Horsepower
			})
		case "hp_asc":
			// Sort cars from least powerful to most powerful.
			sort.Slice(results, func(i, j int) bool {
				return results[i].Specifications.Horsepower < results[j].Specifications.Horsepower
			})
		case "year_desc":
			// Sort cars from newest to oldest model year.
			sort.Slice(results, func(i, j int) bool {
				return results[i].Year > results[j].Year
			})
		case "year_asc":
			// Sort cars from oldest to newest model year.
			sort.Slice(results, func(i, j int) bool {
				return results[i].Year < results[j].Year
			})
		case "name":
			// Sort cars alphabetically by name (A to Z).
			sort.Slice(results, func(i, j int) bool {
				return results[i].Name < results[j].Name
			})
		}

		// Pass all results and filter state to the template.
		// The template uses these to show the correct cars and keep filters selected.
		data["Cars"] = results
		data["Query"] = q
		data["SelectedCategory"] = categoryStr
		data["SelectedManufacturer"] = manufacturerStr
		data["MinHP"] = minHPStr
		data["MaxHP"] = maxHPStr
		data["Sort"] = sortBy
	}

	// Handle manufacturers view
	if view == "manufacturers" {
		// Make a copy of the manufacturers slice so we can attach models to each.
		// We copy to avoid modifying the original db.Manufacturers data.
		manufacturers := make([]Manufacturer, len(db.Manufacturers))
		copy(manufacturers, db.Manufacturers)

		// For each manufacturer, find all car models that belong to it.
		// This allows the template to show a manufacturer and its cars together.
		for i, m := range manufacturers {
			var models []CarModel
			for _, c := range db.CarModels {
				if c.ManufacturerID == m.ID {
					models = append(models, c)
				}
			}
			manufacturers[i].Models = models
		}

		// Override the Manufacturers value in data with the enriched version
		// that includes each manufacturer's car models.
		data["Manufacturers"] = manufacturers
	}

	// Handle recommendations view
	if view == "recommendations" {
		categoryStr := r.URL.Query().Get("category")
		minHPStr := r.URL.Query().Get("minHP")
		maxHPStr := r.URL.Query().Get("maxHP")

		// Only calculate recommendations if the user provided at least one preference.
		// Without any preference, we have nothing to score cars against.
		if categoryStr != "" || minHPStr != "" || maxHPStr != "" {
			// ScoredCar wraps a car with a numeric score and a human-readable reason.
			// The score determines the order. The reason is shown to the user.
			type ScoredCar struct {
				Car    CarModel
				Score  float64
				Reason string
			}

			// This operation reads all view counts from the concurrency-safe store.
			// We take the snapshot once here, before the scoring loop,
			// to avoid repeated blocking calls inside the loop.
			countsSnapshot := viewCounts.Snapshot()

			var scored []ScoredCar
			for _, car := range db.CarModels {
				score := 0.0
				reasons := []string{}

				// Popularity bonus: cars with more views get a higher score.
				if vc, ok := countsSnapshot[car.ID]; ok {
					score += float64(vc) * 2
					if vc > 0 {
						reasons = append(reasons, "Popular choice")
					}
				}

				// Category match bonus: +10 points if car matches the requested category.
				if categoryStr != "" {
					cat := categoryByID(car.CategoryID)
					if cat != nil && strings.EqualFold(cat.Name, categoryStr) {
						score += 10
						reasons = append(reasons, "Matches category")
					}
				}

				// Minimum HP bonus: +5 points if car meets the minimum HP requirement.
				if minHPStr != "" {
					minHP, _ := strconv.Atoi(minHPStr)
					if car.Specifications.Horsepower >= minHP {
						score += 5
					}
				}
				// Maximum HP bonus: +3 points if car is within the HP ceiling.
				if maxHPStr != "" {
					maxHP, _ := strconv.Atoi(maxHPStr)
					if car.Specifications.Horsepower <= maxHP {
						score += 3
					}
				}

				// Recency bonus: newer cars (after 2020) get a small extra score.
				score += float64(car.Year-2020) * 0.5

				// Ensure the score is always at least 0.1 so every car stays eligible.
				if score < 0.1 {
					score = 0.1
				}

				reason := "Based on your preferences"
				if len(reasons) > 0 {
					reason = strings.Join(reasons, " · ")
				}

				scored = append(scored, ScoredCar{Car: car, Score: score, Reason: reason})
			}

			// Sort recommended cars from highest score to lowest.
			sort.Slice(scored, func(i, j int) bool {
				return scored[i].Score > scored[j].Score
			})

			// Return only the top 4 recommendations to avoid overwhelming the user.
			top := scored
			if len(top) > 4 {
				top = top[:4]
			}
			data["Recommendations"] = top
		}

		// Pass the user's filter selections back to the template
		// so the form fields can show the previously selected values.
		data["SelectedCategory"] = categoryStr
		data["SelectedMinHP"] = minHPStr
		data["SelectedMaxHP"] = maxHPStr
	}

	// Handle compare view
	if view == "compare" {
		// Read the list of car IDs to compare from the query string.
		// IDs can be passed as ?ids=1,2 or ?ids=1&ids=2.
		idsValues := r.URL.Query()["ids"]
		if len(idsValues) > 0 {
			// CompareEntry adds manufacturer and category pointers to each car
			// so the template can show brand and type information for each car.
			type CompareEntry struct {
				CarModel
				Manufacturer *Manufacturer
				Category     *Category
			}
			var result []CompareEntry

			// Parse each value which may contain multiple comma-separated IDs.
			for _, raw := range idsValues {
				for _, p := range strings.Split(raw, ",") {
					id, err := strconv.Atoi(strings.TrimSpace(p))
					if err != nil {
						// Skip non-integer values silently (e.g., garbage input).
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
			data["CompareResults"] = result
		}
		// Pass all cars to the template so the user can pick cars to add to the comparison.
		data["AllCars"] = db.CarModels
	}

	// Handle detail view
	if view == "detail" {
		// Read the car ID from the query string (e.g., ?view=detail&id=42).
		idStr := r.URL.Query().Get("id")
		if idStr != "" {
			id, err := strconv.Atoi(idStr)
			if err == nil {
				car := carByID(id)
				if car != nil {
					// DetailedCar enriches the car with manufacturer and category data
					// so the template can display the brand name and car type.
					type DetailedCar struct {
						CarModel
						Manufacturer *Manufacturer
						Category     *Category
					}
					// Record this car view asynchronously to update the popularity counter.
					recordView(id, "page:detail")
					data["DetailCar"] = DetailedCar{
						CarModel:     *car,
						Manufacturer: manufacturerByID(car.ManufacturerID),
						Category:     categoryByID(car.CategoryID),
					}
				}
			}
		}
	}

	// Render the HTML page using the index.html template and the assembled data.
	if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		// Log the error but do not expose internal details to the user.
		// Template errors can reveal code structure, which is a security risk.
		log.Printf("Error rendering template: %v", err)
		serveError(w, http.StatusInternalServerError, "")
	}
}
