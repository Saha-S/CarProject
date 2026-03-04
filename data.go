package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

func loadDataFromAPI(baseURL string) error {
	baseURL = strings.TrimRight(baseURL, "/")

	loaded := Database{}
	if err := fetchJSONFromCandidates(baseURL, []string{"/api/manufacturers", "/manufacturers"}, &loaded.Manufacturers); err != nil {
		return fmt.Errorf("fetch manufacturers: %w", err)
	}
	if err := fetchJSONFromCandidates(baseURL, []string{"/api/categories", "/categories"}, &loaded.Categories); err != nil {
		return fmt.Errorf("fetch categories: %w", err)
	}
	if err := fetchJSONFromCandidates(baseURL, []string{"/api/models", "/api/cars", "/models", "/cars"}, &loaded.CarModels); err != nil {
		return fmt.Errorf("fetch models: %w", err)
	}

	db = loaded
	return nil
}

func fetchJSONFromCandidates(baseURL string, candidates []string, target interface{}) error {
	var lastErr error
	for _, endpoint := range candidates {
		if err := fetchJSON(baseURL+endpoint, target); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}
	return lastErr
}

func fetchJSON(address string, target interface{}) error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(address)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("%s: status %d: %s", address, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("%s: decode json: %w", address, err)
	}

	return nil
}

func imageProxyHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/static/img/")
	if name == "" || strings.Contains(name, "..") {
		http.NotFound(w, r)
		return
	}

	escaped := url.PathEscape(path.Base(name))
	candidates := []string{
		"/api/images/" + escaped,
		"/images/" + escaped,
		"/static/img/" + escaped,
		"/img/" + escaped,
	}

	client := &http.Client{Timeout: 15 * time.Second}
	baseURL := strings.TrimRight(apiBaseURL, "/")
	for _, endpoint := range candidates {
		resp, err := client.Get(baseURL + endpoint)
		if err != nil {
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			continue
		}

		if contentType := resp.Header.Get("Content-Type"); contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		if cacheControl := resp.Header.Get("Cache-Control"); cacheControl != "" {
			w.Header().Set("Cache-Control", cacheControl)
		}
		_, _ = io.Copy(w, resp.Body)
		resp.Body.Close()
		return
	}

	http.NotFound(w, r)
}
