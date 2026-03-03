package main

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	view := r.URL.Query().Get("view")
	if view == "" {
		view = "gallery"
	}

	data := map[string]interface{}{
		"View":           view,
		"Categories":    db.Categories,
		"Manufacturers": db.Manufacturers,
	}

	// Handle gallery/search view
	if view == "gallery" || view == "" {
		q := strings.ToLower(r.URL.Query().Get("q"))
		categoryStr := r.URL.Query().Get("category")
		manufacturerStr := r.URL.Query().Get("manufacturer")
		minHPStr := r.URL.Query().Get("minHP")
		maxHPStr := r.URL.Query().Get("maxHP")
		sortBy := r.URL.Query().Get("sort")

		type EnrichedCar struct {
			CarModel
			ManufacturerName string
			CategoryName     string
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
		data["Manufacturers"] = db.Manufacturers
		for i, m := range db.Manufacturers {
			var models []CarModel
			for _, c := range db.CarModels {
				if c.ManufacturerID == m.ID {
					models = append(models, c)
				}
			}
			db.Manufacturers[i].Models = models
		}
	}

	// Handle recommendations view
	if view == "recommendations" {
		categoryStr := r.URL.Query().Get("category")
		minHPStr := r.URL.Query().Get("minHP")
		maxHPStr := r.URL.Query().Get("maxHP")

		if categoryStr != "" || minHPStr != "" || maxHPStr != "" {
			type ScoredCar struct {
				Car    CarModel
				Score  float64
				Reason string
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
						reasons = append(reasons, "Matches category")
					}
				}

				if minHPStr != "" {
					minHP, _ := strconv.Atoi(minHPStr)
					if car.Specifications.Horsepower >= minHP {
						score += 5
					}
				}
				if maxHPStr != "" {
					maxHP, _ := strconv.Atoi(maxHPStr)
					if car.Specifications.Horsepower <= maxHP {
						score += 3
					}
				}

				score += float64(car.Year-2020) * 0.5
				if score < 0.1 {
					score = 0.1
				}

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
			data["Recommendations"] = top
		}

		data["SelectedCategory"] = categoryStr
		data["SelectedMinHP"] = minHPStr
		data["SelectedMaxHP"] = maxHPStr
	}

	// Handle compare view
	if view == "compare" {
		idsStr := r.URL.Query().Get("ids")
		if idsStr != "" {
			parts := strings.Split(idsStr, ",")
			type CompareEntry struct {
				CarModel
				Manufacturer *Manufacturer
				Category     *Category
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
			data["CompareResults"] = result
		}
		data["AllCars"] = db.CarModels
	}

	// Handle detail view
	if view == "detail" {
		idStr := r.URL.Query().Get("id")
		if idStr != "" {
			id, err := strconv.Atoi(idStr)
			if err == nil {
				car := carByID(id)
				if car != nil {
					viewCounts[id]++
					type DetailedCar struct {
						CarModel
						Manufacturer *Manufacturer
						Category     *Category
					}
					data["DetailCar"] = DetailedCar{
						CarModel:     *car,
						Manufacturer: manufacturerByID(car.ManufacturerID),
						Category:     categoryByID(car.CategoryID),
					}
				}
			}
		}
	}

	tmpl.ExecuteTemplate(w, "index.html", data)
}
