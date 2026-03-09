// File purpose:
// This is the starting point of the entire application.
// When you run "go run .", Go executes this file first.
//
// Responsibilities:
// - Read configuration from environment variables (port number, API URL)
// - Load all car data from the external Node.js API at startup
// - Register custom template functions used in HTML pages
// - Parse all HTML template files
// - Register all HTTP routes (URL paths and their handler functions)
// - Start the HTTP web server
//
// Used by:
// - Everything. This file is the entry point that wires together all other files.

package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
)

// html/template is used to safely render HTML pages from Go data.
// It automatically escapes values to prevent XSS (cross-site scripting) attacks,
// which makes it safer than building HTML strings manually.

// log is used to print messages to the terminal while the app is running.
// This helps developers see what is happening and debug problems.

// net/http is Go's built-in library for building web servers.
// It handles incoming HTTP requests, URL routing, and sending responses.

// os is used to read environment variables like PORT and API_BASE_URL.
// Environment variables allow changing settings without editing code.

// main is the entry point of the Go application.
// Go always starts executing from the main() function in the main package.
//
// Important:
// The steps below must happen in this exact order.
// Data must be loaded before routes are registered,
// and templates must be parsed before the server starts handling requests.
//
// Side effects:
// - Sets the global apiBaseURL variable used by data.go
// - Populates the global db variable with car data
// - Parses and stores HTML templates in the global tmpl variable
// - Starts the HTTP server (this call blocks forever)
func main() {
	// Read the API base URL from the environment variable.
	// This tells the app where to fetch car data from.
	// Default to localhost:3000 where the Node.js API runs locally.
	apiBaseURL = os.Getenv("API_BASE_URL")
	if apiBaseURL == "" {
		apiBaseURL = "http://localhost:3000"
	}

	// Load all car data from the API before starting the web server.
	// Important: The app cannot serve pages without this data,
	// so we stop immediately (Fatalf) if loading fails.
	if err := loadDataFromAPI(apiBaseURL); err != nil {
		log.Fatalf("Failed to load data from API (%s): %v", apiBaseURL, err)
	}

	var err error
	// Register custom helper functions available inside HTML templates.
	// "add" adds two integers (used for pagination calculations in templates).
	// "mul" multiplies two floats (used for score display in templates).
	// These functions must be registered before templates are parsed.
	funcs := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"mul": func(a, b float64) float64 {
			return a * b
		},
	}

	// Parse all HTML template files from the templates/ directory.
	// ParseGlob reads all files matching the pattern and stores them
	// so we can render them later by name (e.g., "index.html").
	// We stop the app if any template has a syntax error.
	tmpl, err = template.New("").Funcs(funcs).ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	// Read the port number from the environment variable.
	// Default to 8080 if not set. This allows changing the port
	// without modifying the code (useful in different environments).
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Serve static files (CSS, JavaScript, fonts) from the static/ directory.
	// StripPrefix removes "/static/" from the URL so the file server
	// can find files relative to the static/ folder.
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Use a special proxy handler for images instead of serving them directly.
	// This is because car images are hosted on the external API server,
	// not stored locally in the static/ folder.
	http.HandleFunc("/static/img/", imageProxyHandler)

	// Register all JSON API endpoints.
	// These routes return JSON data instead of HTML pages.
	// They are used by JavaScript on the frontend to fetch data dynamically.
	http.HandleFunc("/api/models", apiModels)
	http.HandleFunc("/api/models/", apiModelByID)
	http.HandleFunc("/api/manufacturers", apiManufacturers)
	http.HandleFunc("/api/manufacturers/", apiManufacturerByID)
	http.HandleFunc("/api/categories", apiCategories)
	http.HandleFunc("/api/categories/", apiCategoryByID)
	http.HandleFunc("/api/search", apiSearch)
	http.HandleFunc("/api/compare", apiCompare)
	http.HandleFunc("/api/recommendations", apiRecommendations)

	// Register the main page handler.
	// The "/" route matches all requests that don't match the routes above.
	// It renders the HTML page for gallery, detail, compare, and recommendations views.
	http.HandleFunc("/", indexHandler)

	// Wrap all handlers with panic recovery middleware.
	// If any handler crashes (panics), this middleware catches the crash,
	// logs it, and returns a 500 error page instead of crashing the whole server.
	handler := recoveryMiddleware(http.DefaultServeMux)

	// Log a startup message so the developer knows the server is running.
	log.Printf("🚗 Cars Viewer running on http://localhost:%s (API: %s)", port, apiBaseURL)

	// Start the HTTP server. This call blocks forever (or until the server crashes).
	// log.Fatal logs the error and exits the program if the server fails to start.
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
