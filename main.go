package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
)

func main() {
	dataPath := "data.json"
	if len(os.Args) > 1 {
		dataPath = os.Args[1]
	}

	if err := loadData(dataPath); err != nil {
		log.Fatalf("Failed to load data: %v", err)
	}

	var err error
	funcs := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"mul": func(a, b float64) float64 {
			return a * b
		},
	}
	
	tmpl, err = template.New("").Funcs(funcs).ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/api/models", apiModels)
	http.HandleFunc("/api/models/", apiModelByID)
	http.HandleFunc("/api/manufacturers", apiManufacturers)
	http.HandleFunc("/api/manufacturers/", apiManufacturerByID)
	http.HandleFunc("/api/categories", apiCategories)
	http.HandleFunc("/api/categories/", apiCategoryByID)
	http.HandleFunc("/api/search", apiSearch)
	http.HandleFunc("/api/compare", apiCompare)
	http.HandleFunc("/api/recommendations", apiRecommendations)

	http.HandleFunc("/", indexHandler)

	log.Printf("🚗 Cars Viewer running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
