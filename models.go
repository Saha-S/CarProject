// File purpose:
// This file defines all the data models (structs) used throughout the application.
// It also contains the concurrent view counter store that safely tracks
// how many times each car has been viewed.
//
// Responsibilities:
// - Define data structures for car models, manufacturers, categories, and the database
// - Provide a concurrency-safe view counter using Go channels
// - Declare global variables shared across the entire application
//
// Used by:
// - All other files in the project use these structs and global variables

package main

import "html/template"

// html/template is imported here because the global tmpl variable
// (which holds parsed HTML templates) needs this type.
// Templates are used by handlers to render HTML pages.

// viewCounterStore is a concurrency-safe counter that tracks
// how many times each car (identified by its ID) has been viewed.
//
// Why use channels instead of a simple map with a mutex lock?
// Channels are a Go-idiomatic way to share data safely between goroutines
// (lightweight threads). Only one goroutine (the run() loop) ever reads
// or writes the counts map, which avoids data races entirely.
type viewCounterStore struct {
	// incrementCh receives car IDs to increment their view count.
	// Any goroutine can safely send a car ID here.
	incrementCh chan int

	// snapshotCh receives requests for a copy of all view counts.
	// The response is sent back through a channel inside the request.
	snapshotCh chan viewCountSnapshotReq
}

// viewCountSnapshotReq is a request message asking for a snapshot
// of the current view counts.
//
// Why use a channel for the response?
// Because the run() goroutine needs to send the snapshot back to
// the caller. A response channel is the standard Go pattern for this.
type viewCountSnapshotReq struct {
	// resp is the channel where the view counts snapshot will be sent back.
	resp chan map[int]int
}

// newViewCounterStore creates and starts a new viewCounterStore.
//
// Output:
// - Returns a pointer to a new, running viewCounterStore
//
// Side effects:
// - Starts a background goroutine (store.run) that processes view count updates.
//   This goroutine runs for the lifetime of the application.
func newViewCounterStore() *viewCounterStore {
	store := &viewCounterStore{
		incrementCh: make(chan int),
		snapshotCh:  make(chan viewCountSnapshotReq),
	}

	// Start the background processing loop in a separate goroutine.
	// This allows the store to handle requests concurrently
	// without blocking the callers.
	go store.run()
	return store
}

// run is the internal processing loop of the viewCounterStore.
// It runs in its own goroutine and is the ONLY place that reads
// or writes the counts map. This design avoids race conditions.
//
// Important:
// This function runs forever. It is started by newViewCounterStore
// and never stops while the application is running.
//
// Side effects:
// - Updates the counts map when increment requests arrive
// - Sends a copy of the counts map when snapshot requests arrive
func (s *viewCounterStore) run() {
	// counts is a map from car ID to number of views.
	// Only this goroutine accesses this map, so no locking is needed.
	counts := make(map[int]int)

	// Wait for incoming requests forever using a select statement.
	// select picks whichever channel has a message ready.
	for {
		select {
		// An increment request: increase the view count for the given car ID.
		case id := <-s.incrementCh:
			counts[id]++

		// A snapshot request: send a copy of all counts back to the requester.
		// We copy the map so the caller gets a stable snapshot
		// that won't change even if new views come in afterward.
		case req := <-s.snapshotCh:
			snapshot := make(map[int]int, len(counts))
			for id, c := range counts {
				snapshot[id] = c
			}
			req.resp <- snapshot
		}
	}
}

// Increment records one new view for the car with the given ID.
//
// Input:
// - id: the unique identifier of the car that was viewed
//
// Side effects:
// - Sends the car ID to the background goroutine to update the count.
//   This call blocks briefly until the run() goroutine is ready to receive.
func (s *viewCounterStore) Increment(id int) {
	s.incrementCh <- id
}

// Snapshot returns a copy of all current view counts as a map.
//
// Output:
// - Returns a map where keys are car IDs and values are view counts.
//   The map is a copy, so modifying it does not affect the stored counts.
//
// Side effects:
// - Blocks briefly until the background goroutine processes the request.
//   This is safe to call from multiple goroutines at the same time.
func (s *viewCounterStore) Snapshot() map[int]int {
	// Create a response channel to receive the snapshot from the run() goroutine.
	resp := make(chan map[int]int)
	s.snapshotCh <- viewCountSnapshotReq{resp: resp}
	return <-resp
}

// Specifications holds the technical details of a car model.
// These fields come directly from the external API and are
// stored using JSON tags so they map correctly to the API field names.
type Specifications struct {
	Engine       string `json:"engine"`       // Engine type, e.g., "V8 4.0L"
	Horsepower   int    `json:"horsepower"`   // Power output in horsepower (HP)
	Transmission string `json:"transmission"` // e.g., "Automatic" or "Manual"
	Drivetrain   string `json:"drivetrain"`   // e.g., "AWD", "RWD", "FWD"
}

// CarModel represents a single car model in the system.
// Each car belongs to one manufacturer and one category.
// The json tags control how the fields are named in JSON responses.
type CarModel struct {
	ID             int            `json:"id"`             // Unique identifier for this car
	Name           string         `json:"name"`           // Model name, e.g., "Mustang"
	ManufacturerID int            `json:"manufacturerId"` // Links to a Manufacturer by its ID
	CategoryID     int            `json:"categoryId"`     // Links to a Category by its ID
	Year           int            `json:"year"`           // Model year, e.g., 2023
	Specifications Specifications `json:"specifications"` // Technical details of the car
	Image          string         `json:"image"`          // Image file name (served via the image proxy)
}

// Manufacturer represents a car brand or company.
// Examples: Toyota, Ford, BMW.
//
// The Models field is optional (omitempty means it is left out of JSON
// if empty). It is only populated when the API needs to return
// a manufacturer along with all its car models.
type Manufacturer struct {
	ID           int        `json:"id"`
	Name         string     `json:"name"`             // Brand name, e.g., "Toyota"
	Country      string     `json:"country"`          // Country where the brand is from
	FoundingYear int        `json:"foundingYear"`     // Year the company was founded
	Models       []CarModel `json:"models,omitempty"` // Cars made by this manufacturer (optional)
}

// Category represents a type or class of car.
// Examples: SUV, Sedan, Sports Car, Truck.
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"` // Category name, e.g., "SUV"
}

// Database holds all the data loaded from the external API at startup.
// It acts as an in-memory store for all cars, manufacturers, and categories.
//
// Assumption:
// Data is loaded once at startup and does not change while the app is running.
// If the data needs to stay up-to-date, a periodic reload would be needed.
//
// TODO:
// Consider adding a periodic data refresh mechanism if the API data
// changes frequently while the app is running.
type Database struct {
	Manufacturers []Manufacturer `json:"manufacturers"` // All car brands
	Categories    []Category     `json:"categories"`    // All car categories
	CarModels     []CarModel     `json:"carModels"`     // All car models
}

// Global variables shared across all files in the application.
// These are initialized at startup and used throughout the app's lifetime.

// viewCounts tracks how many times each car has been viewed.
// It uses a concurrency-safe store so multiple requests can update it safely.
var viewCounts = newViewCounterStore()

// appEvents is the async event bus used to record user actions
// like viewing a car or performing a search.
// Using a buffer of 256 means up to 256 events can be queued
// before any new events are dropped.
var appEvents = newEventBus(256)

// db holds all the car data loaded from the external API.
// It is populated once at startup by loadDataFromAPI in data.go.
var db Database

// tmpl holds all parsed HTML templates.
// It is populated at startup in main.go and used by handlers to render pages.
var tmpl *template.Template

// apiBaseURL is the base URL of the external Node.js API service.
// It is set at startup from the API_BASE_URL environment variable.
var apiBaseURL string
