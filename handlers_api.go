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

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func apiModels(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, db.CarModels)
}

func apiModelByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/models/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	car := carByID(id)
	if car == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	viewCounts[id]++

	type DetailedCar struct {
		CarModel
		Manufacturer *Manufacturer `json:"manufacturer"`
		Category     *Category     `json:"category"`
	}
	jsonResponse(w, DetailedCar{
		CarModel:     *car,
		Manufacturer: manufacturerByID(car.ManufacturerID),
		Category:     categoryByID(car.CategoryID),
	})
}

func apiManufacturers(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, db.Manufacturers)
}

func apiManufacturerByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/manufacturers/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	m := manufacturerByID(id)
	if m == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var models []CarModel
	for _, c := range db.CarModels {
		if c.ManufacturerID == id {
			models = append(models, c)
		}
	}
	jsonResponse(w, map[string]interface{}{"manufacturer": m, "models": models})
}

func apiCategories(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, db.Categories)
}

func apiCategoryByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/categories/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	cat := categoryByID(id)
	if cat == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	var models []CarModel
	for _, c := range db.CarModels {
		if c.CategoryID == id {
			models = append(models, c)
		}
	}
	jsonResponse(w, map[string]interface{}{"category": cat, "models": models})
}

func apiCompare(w http.ResponseWriter, r *http.Request) {
	idsStr := r.URL.Query().Get("ids")
	if idsStr == "" {
		http.Error(w, "ids param required", http.StatusBadRequest)
		return
	}
	parts := strings.Split(idsStr, ",")
	type CompareEntry struct {
		CarModel
		Manufacturer *Manufacturer `json:"manufacturer"`
		Category     *Category     `json:"category"`
	}
	var result []CompareEntry
	for _, p := range parts {
		id, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			continue
		}
		car := carByID(id)
		if car == nil {
			continue
		}
		result = append(result, CompareEntry{
			CarModel:     *car,
			Manufacturer: manufacturerByID(car.ManufacturerID),
			Category:     categoryByID(car.CategoryID),
		})
	}
	jsonResponse(w, result)
}

func apiRecommendations(w http.ResponseWriter, r *http.Request) {
	categoryStr := r.URL.Query().Get("category")
	minHPStr := r.URL.Query().Get("minHP")
	maxHPStr := r.URL.Query().Get("maxHP")

	type ScoredCar struct {
		Car    CarModel `json:"car"`
		Score  float64  `json:"score"`
		Reason string   `json:"reason"`
	}

	var scored []ScoredCar
	for _, car := range db.CarModels {
		score := 0.0
		reasons := []string{}

		if vc, ok := viewCounts[car.ID]; ok {
			score += float64(vc) * 2
			if vc > 0 {
				reasons = append(reasons, "Popular choice")
			}
		}

		if categoryStr != "" {
			cat := categoryByID(car.CategoryID)
			if cat != nil && strings.EqualFold(cat.Name, categoryStr) {
				score += 10
				reasons = append(reasons, fmt.Sprintf("Matches %s category", cat.Name))
			}
		}

		if minHPStr != "" {
			minHP, _ := strconv.Atoi(minHPStr)
			if car.Specifications.Horsepower >= minHP {
				score += 5
				reasons = append(reasons, fmt.Sprintf("%d HP meets minimum", car.Specifications.Horsepower))
			}
		}
		if maxHPStr != "" {
			maxHP, _ := strconv.Atoi(maxHPStr)
			if car.Specifications.Horsepower <= maxHP {
				score += 3
			}
		}

		score += float64(car.Year-2020) * 0.5
		score = math.Max(score, 0.1)

		reason := "Based on your preferences"
		if len(reasons) > 0 {
			reason = strings.Join(reasons, " · ")
		}

		scored = append(scored, ScoredCar{Car: car, Score: score, Reason: reason})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	top := scored
	if len(top) > 4 {
		top = top[:4]
	}
	jsonResponse(w, top)
}

func apiSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(r.URL.Query().Get("q"))
	categoryStr := r.URL.Query().Get("category")
	manufacturerStr := r.URL.Query().Get("manufacturer")
	minHPStr := r.URL.Query().Get("minHP")
	maxHPStr := r.URL.Query().Get("maxHP")
	minYearStr := r.URL.Query().Get("minYear")
	maxYearStr := r.URL.Query().Get("maxYear")
	sortBy := r.URL.Query().Get("sort")

	type EnrichedCar struct {
		CarModel
		ManufacturerName string `json:"manufacturerName"`
		CategoryName     string `json:"categoryName"`
	}

	var results []EnrichedCar
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

		if q != "" {
			haystack := strings.ToLower(car.Name + " " + mfrName + " " + catName +
				" " + car.Specifications.Engine + " " + car.Specifications.Transmission)
			if !strings.Contains(haystack, q) {
				continue
			}
		}

		if categoryStr != "" && !strings.EqualFold(catName, categoryStr) {
			continue
		}

		if manufacturerStr != "" && !strings.EqualFold(mfrName, manufacturerStr) {
			continue
		}

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

	switch sortBy {
	case "hp_desc":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Specifications.Horsepower > results[j].Specifications.Horsepower
		})
	case "hp_asc":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Specifications.Horsepower < results[j].Specifications.Horsepower
		})
	case "year_desc":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Year > results[j].Year
		})
	case "year_asc":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Year < results[j].Year
		})
	case "name":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Name < results[j].Name
		})
	}

	jsonResponse(w, results)
}
