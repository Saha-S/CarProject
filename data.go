// File purpose:
// This file handles all data loading and image serving for the application.
//
// Responsibilities:
// - Fetch car data (manufacturers, categories, car models) from the external Node.js API
// - Try multiple API endpoint paths in case the API uses a different URL structure
// - Serve car images by proxying requests to the external API
//
// Used by:
// - main.go: calls loadDataFromAPI at startup to populate the in-memory database
// - main.go: registers imageProxyHandler for the /static/img/ URL path

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// encoding/json is used to decode JSON responses from the external API
// into Go structs (like []CarModel, []Manufacturer, []Category).

// fmt is used to build formatted error messages when something goes wrong.

// io is used to read HTTP response bodies and to copy image data to the response.

// log is used to log errors when image copying fails partway through.

// net/http is used to make HTTP GET requests to the external API.

// net/url is used to safely encode image file names in URLs,
// handling special characters like spaces or non-ASCII characters.

// path is used to extract just the file name from a URL path,
// preventing directory traversal issues.

// strings is used for URL string manipulation (e.g., TrimRight, TrimPrefix).

// time is used to set timeouts on HTTP requests to the external API.
// Without a timeout, a slow or unresponsive API could hang the app forever.

// loadDataFromAPI fetches all car data from the external API and stores it in db.
// This is called once at startup. If it fails, the application cannot start.
//
// Input:
// - baseURL: the base URL of the external API (e.g., "http://localhost:3000")
//
// Output:
// - Returns nil if all data was loaded successfully
// - Returns an error describing which fetch failed
//
// Side effects:
// - Populates the global db variable with manufacturers, categories, and car models.
//
// Design decision:
// We load all data into memory at startup instead of querying the API per request.
// This makes the app faster (no per-request API calls) and simpler,
// but means the data can become stale if the API data changes while the app is running.
func loadDataFromAPI(baseURL string) error {
	// Remove any trailing slash from the URL to avoid double slashes
	// when we append endpoint paths below (e.g., avoid "localhost:3000//api/models").
	baseURL = strings.TrimRight(baseURL, "/")

	loaded := Database{}

	// Try multiple endpoint paths for manufacturers because the API might
	// use either "/api/manufacturers" or "/manufacturers" depending on its version.
	if err := fetchJSONFromCandidates(baseURL, []string{"/api/manufacturers", "/manufacturers"}, &loaded.Manufacturers); err != nil {
		return fmt.Errorf("fetch manufacturers: %w", err)
	}

	// Same approach for categories.
	if err := fetchJSONFromCandidates(baseURL, []string{"/api/categories", "/categories"}, &loaded.Categories); err != nil {
		return fmt.Errorf("fetch categories: %w", err)
	}

	// Try multiple paths for car models. Different API versions may call them
	// "models", "cars", or a combination. We try all known options.
	if err := fetchJSONFromCandidates(baseURL, []string{"/api/models", "/api/cars", "/models", "/cars"}, &loaded.CarModels); err != nil {
		return fmt.Errorf("fetch models: %w", err)
	}

	// Important:
	// Only update the global db after all fetches succeed.
	// This prevents partially loading data (e.g., having manufacturers
	// but no car models because the second fetch failed).
	db = loaded
	return nil
}

// fetchJSONFromCandidates tries each URL candidate in order and returns
// the result from the first one that succeeds.
// This is used to support APIs that may have different URL structures.
//
// Input:
// - baseURL: the base URL of the API
// - candidates: a list of endpoint paths to try, in order
// - target: a pointer to the Go variable where the decoded JSON will be stored
//
// Output:
// - Returns nil if any candidate succeeded
// - Returns the error from the last failed candidate if all fail
func fetchJSONFromCandidates(baseURL string, candidates []string, target interface{}) error {
	var lastErr error
	// Try each candidate URL in order. Return immediately on first success.
	for _, endpoint := range candidates {
		if err := fetchJSON(baseURL+endpoint, target); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}
	// All candidates failed. Return the last error as a hint of what went wrong.
	return lastErr
}

// fetchJSON makes an HTTP GET request to the given address and decodes
// the JSON response body into the target variable.
//
// Input:
// - address: the full URL to fetch (e.g., "http://localhost:3000/api/models")
// - target: a pointer to a Go variable where decoded JSON will be stored
//
// Output:
// - Returns nil on success
// - Returns an error if the request fails, the server returns an error status,
//   or the response body cannot be decoded as JSON
//
// Side effects:
// - Makes a network request to the given address.
// - Uses a 10-second timeout to prevent hanging if the API is unresponsive.
func fetchJSON(address string, target interface{}) error {
	// Use a client with a timeout instead of the default http client.
	// Without a timeout, a slow or unresponsive API would hang the app forever.
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(address)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check that the API returned a success status code (200 OK).
	// Any other status means something went wrong on the API side.
	if resp.StatusCode != http.StatusOK {
		// Read up to 512 bytes of the error response body for context.
		// We limit to 512 bytes to avoid reading huge error pages into memory.
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		// Try to parse error as JSON to extract user-friendly message
		var errorResp map[string]interface{}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			if msg, ok := errorResp["message"].(string); ok && msg != "" {
				return fmt.Errorf("service error: %s", msg)
			}
		}
		// Fallback to generic error message if JSON parsing fails or has no message field.
		return fmt.Errorf("data service returned error %d", resp.StatusCode)
	}

	// Decode the JSON response body directly into the target variable.
	// Using a streaming decoder (NewDecoder) is more efficient than
	// reading the full body into a []byte first, especially for large datasets.
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("%s: decode json: %w", address, err)
	}

	return nil
}

// imageProxyHandler handles requests to /static/img/<filename>.
// Instead of serving images from a local folder, it fetches them
// from the external API and streams them back to the browser.
//
// Input:
// - w: the HTTP response writer to send the image data to
// - r: the incoming HTTP request containing the image filename in the URL path
//
// Why use a proxy instead of serving images directly?
// Car images are stored on the external API server, not in this app's file system.
// A proxy lets the browser think images are served locally while
// actually fetching them from the remote API.
//
// Side effects:
// - Makes one or more HTTP requests to the external API to find the image.
// - Copies the image data from the API response to the browser response.
// - Logs an error if copying the image data fails partway through.
func imageProxyHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the image filename from the URL path.
	name := strings.TrimPrefix(r.URL.Path, "/static/img/")

	// Validate the image name to prevent path traversal attacks.
	// An empty name or a name with ".." could allow an attacker to access
	// files outside the intended directory, which is a security risk.
	if name == "" || strings.Contains(name, "..") {
		http.NotFound(w, r)
		return
	}

	// URL-encode the filename to handle special characters safely.
	// path.Base extracts just the filename (removes any directory parts).
	// url.PathEscape ensures characters like spaces are properly encoded in the URL.
	escaped := url.PathEscape(path.Base(name))

	// Try multiple paths where the API might serve images.
	// Different API versions may use different URL patterns for images.
	candidates := []string{
		"/api/images/" + escaped,
		"/images/" + escaped,
		"/static/img/" + escaped,
		"/img/" + escaped,
	}

	// Use a longer timeout for images than for data requests (15s vs 10s).
	// Images can be large files that take more time to transfer.
	client := &http.Client{Timeout: 15 * time.Second}
	baseURL := strings.TrimRight(apiBaseURL, "/")

	// Try each candidate URL until we find one that returns a successful response.
	for _, endpoint := range candidates {
		resp, err := client.Get(baseURL + endpoint)
		if err != nil {
			// This candidate failed (e.g., connection error). Try the next one.
			continue
		}
		if resp.StatusCode != http.StatusOK {
			// The server responded, but the image was not found at this path.
			resp.Body.Close()
			continue
		}

		// Forward the Content-Type header so the browser knows what kind of
		// image it is receiving (e.g., "image/jpeg", "image/png").
		if contentType := resp.Header.Get("Content-Type"); contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}

		// Forward the Cache-Control header so the browser can cache the image.
		// Caching avoids repeated image downloads, making the app faster.
		if cacheControl := resp.Header.Get("Cache-Control"); cacheControl != "" {
			w.Header().Set("Cache-Control", cacheControl)
		}

		// Copy the image data from the API response directly to the browser response.
		if _, err := io.Copy(w, resp.Body); err != nil {
			// Log if copying fails (e.g., the client disconnected mid-download).
			// We cannot send an error response at this point because
			// we have already started writing the response body.
			log.Printf("Error copying image response: %v", err)
		}
		resp.Body.Close()
		return
	}

	// None of the candidate paths returned the image. Return 404 Not Found.
	http.NotFound(w, r)
}
