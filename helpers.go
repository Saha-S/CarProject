// File purpose:
// This file contains helper utility functions used across the application.
//
// Responsibilities:
// - Render error pages with proper HTTP status codes
// - Provide panic recovery middleware to prevent server crashes
// - Provide lookup functions to find records by ID from the in-memory database
//
// Used by:
// - handlers_page.go: uses serveError and the lookup functions
// - handlers_api.go: uses the lookup functions
// - main.go: uses recoveryMiddleware to wrap all HTTP handlers

package main

import (
	"fmt"
	"log"
	"net/http"
)

// fmt is used to write a plain text fallback error response
// if the error HTML template itself fails to render.

// log is used to record unexpected errors (like a template rendering failure)
// so developers can find problems in the server logs.

// net/http provides types for HTTP request/response handling
// and standard HTTP status code constants like http.StatusInternalServerError.

// serveError renders an error page with the given status code and optional message.
// It is called whenever a request cannot be completed successfully.
//
// Input:
// - w: the HTTP response writer to send the error page to
// - statusCode: the HTTP status code to return (e.g., 404, 500)
// - message: an optional message shown on the error page (pass empty string for no message)
//
// Side effects:
// - Writes the HTTP status code to the response
// - Renders the error.html template with the message
// - If the template fails, logs the error and returns a plain text error instead
//   (This fallback prevents a situation where an error handler itself causes another crash)
func serveError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	data := map[string]interface{}{
		"Message": message,
	}
	// Try to render the HTML error page.
	// If the template fails (very rare), fall back to a plain text message.
	if err := tmpl.ExecuteTemplate(w, "error.html", data); err != nil {
		// Log the template error so developers can investigate.
		log.Printf("Error rendering error template: %v", err)
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Internal Server Error\nSomething went wrong on the server.")
	}
}

// recoveryMiddleware wraps an HTTP handler with panic recovery.
// If any handler panics (crashes unexpectedly), this middleware
// catches the panic, logs it, and returns a 500 Internal Server Error page
// instead of crashing the whole web server.
//
// Input:
// - next: the HTTP handler to wrap with panic recovery
//
// Output:
// - Returns a new HTTP handler that adds panic recovery around next
//
// Why use middleware?
// Instead of adding panic recovery to every single handler function,
// we wrap all handlers at once in main.go. This keeps code DRY (Don't Repeat Yourself).
//
// Side effects:
// - Logs the panic value to the terminal if a panic is caught
// - Sends a 500 error response to the client if a panic is caught
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// defer runs this function when the handler returns or panics.
		// recover() catches any panic and prevents the server from crashing.
		defer func() {
			if err := recover(); err != nil {
				// Log the panic so developers can investigate what went wrong.
				log.Printf("PANIC recovered: %v", err)
				// Return a generic 500 error page to the user.
				// We pass an empty message to avoid exposing internal details
				// to the end user, which could be a security risk.
				serveError(w, http.StatusInternalServerError, "")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// manufacturerByID finds a manufacturer in the in-memory database by its ID.
//
// Input:
// - id: the unique ID of the manufacturer to find
//
// Output:
// - Returns a pointer to the Manufacturer if found
// - Returns nil if no manufacturer with that ID exists
//
// Assumption:
// The db.Manufacturers slice is loaded at startup and does not change.
// This function is safe to call from multiple goroutines simultaneously
// because it only reads the data, never writes.
func manufacturerByID(id int) *Manufacturer {
	// Linear search through all manufacturers.
	// This is acceptable because the total number of manufacturers is small.
	// We return a pointer to the original item in the slice (not a copy)
	// so callers can attach extra data to it (e.g., its models list).
	for i := range db.Manufacturers {
		if db.Manufacturers[i].ID == id {
			return &db.Manufacturers[i]
		}
	}
	// Return nil to signal "not found" so callers can handle missing data gracefully.
	return nil
}

// categoryByID finds a category in the in-memory database by its ID.
//
// Input:
// - id: the unique ID of the category to find
//
// Output:
// - Returns a pointer to the Category if found
// - Returns nil if no category with that ID exists
//
// Assumption:
// The db.Categories slice is loaded at startup and does not change.
func categoryByID(id int) *Category {
	for i := range db.Categories {
		if db.Categories[i].ID == id {
			return &db.Categories[i]
		}
	}
	return nil
}

// carByID finds a car model in the in-memory database by its ID.
//
// Input:
// - id: the unique ID of the car to find
//
// Output:
// - Returns a pointer to the CarModel if found
// - Returns nil if no car with that ID exists
//
// Assumption:
// The db.CarModels slice is loaded at startup and does not change.
func carByID(id int) *CarModel {
	for i := range db.CarModels {
		if db.CarModels[i].ID == id {
			return &db.CarModels[i]
		}
	}
	return nil
}
