package main

import (
	"fmt"
	"log"
	"net/http"
)

// serveError renders an error page with the given status code and optional message
func serveError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	data := map[string]interface{}{
		"Message": message,
	}
	if err := tmpl.ExecuteTemplate(w, "error.html", data); err != nil {
		log.Printf("Error rendering error template: %v", err)
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Internal Server Error\nSomething went wrong on the server.")
	}
}

// recoveryMiddleware wraps handlers with panic recovery
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC recovered: %v", err)
				serveError(w, http.StatusInternalServerError, "")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func manufacturerByID(id int) *Manufacturer {
	for i := range db.Manufacturers {
		if db.Manufacturers[i].ID == id {
			return &db.Manufacturers[i]
		}
	}
	return nil
}

func categoryByID(id int) *Category {
	for i := range db.Categories {
		if db.Categories[i].ID == id {
			return &db.Categories[i]
		}
	}
	return nil
}

func carByID(id int) *CarModel {
	for i := range db.CarModels {
		if db.CarModels[i].ID == id {
			return &db.CarModels[i]
		}
	}
	return nil
}
