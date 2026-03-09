// File purpose:
// This file implements a simple asynchronous event system for the application.
// It allows different parts of the app to record user actions
// (like viewing a car or performing a search) without slowing down HTTP requests.
//
// Responsibilities:
// - Define event types and the AppEvent data structure
// - Provide an event bus that queues and processes events asynchronously
// - Provide helper functions to record specific events (car views, searches)
//
// Used by:
// - handlers_page.go: records car view and search events from the HTML page handler
// - handlers_api.go: records car view and search events from the JSON API handler
// - models.go: defines viewCounts which the event bus updates on car_viewed events

package main

import (
	"log"
	"time"
)

// log is used to warn when the event queue is full and an event is dropped.
// This helps developers detect if the system is under heavy load.

// time is used to timestamp each event so we know exactly when it happened.
// All timestamps are stored in UTC to avoid timezone confusion.

// AppEventType is a string type that represents the kind of event.
// Using a named string type (instead of a plain string) makes the code
// more readable and prevents accidentally passing the wrong string value.
type AppEventType string

// EventCarViewed is fired when a user views a car's detail page.
// This event causes the view counter for that car to be incremented.
const EventCarViewed AppEventType = "car_viewed"

// EventSearchPerformed is fired when a user performs a search.
// This event is recorded for future use (e.g., tracking popular search terms).
const EventSearchPerformed AppEventType = "search_performed"

// AppEvent represents a single user action that happened in the app.
// Events are created and published by handler functions,
// then processed asynchronously by the event bus.
type AppEvent struct {
	Type      AppEventType // What kind of event this is (e.g., "car_viewed")
	CarID     int          // The ID of the car involved (only used for car_viewed events)
	Query     string       // The search term (only used for search_performed events)
	Source    string       // Where the event came from, e.g., "page:detail" or "api:search"
	Timestamp time.Time    // When the event happened (always in UTC)
}

// eventBus is a simple message queue for app events.
// It receives events from any part of the app and processes them
// in a separate goroutine so HTTP handlers do not have to wait.
//
// Why use an event bus instead of processing events directly in the handler?
// Processing events (like updating view counts) takes time.
// An event bus lets the HTTP handler return immediately
// and process the event in the background, keeping responses fast.
type eventBus struct {
	// eventsCh is a buffered channel that holds pending events.
	// Buffered means it can store events without a receiver being ready,
	// so publishers do not block (up to the buffer limit).
	eventsCh chan AppEvent
}

// newEventBus creates a new event bus and starts its background processing loop.
//
// Input:
// - buffer: how many events can be queued before new ones are dropped.
//   A larger buffer means the system handles traffic spikes better.
//
// Output:
// - Returns a pointer to the running event bus.
//
// Side effects:
// - Starts a background goroutine (bus.run) that processes events forever.
func newEventBus(buffer int) *eventBus {
	bus := &eventBus{
		// A buffered channel lets publishers send events without blocking,
		// as long as the queue is not full.
		eventsCh: make(chan AppEvent, buffer),
	}

	// Start the background processing goroutine asynchronously.
	// This goroutine runs for the entire lifetime of the application.
	go bus.run()
	return bus
}

// run is the internal processing loop of the event bus.
// It reads events from the channel one by one and handles them.
//
// Important:
// This function runs in its own goroutine and never stops.
// It is started by newEventBus.
//
// Side effects:
// - For car_viewed events: updates the view counter for the car.
// - For other event types: currently received but not processed further.
//
// TODO:
// Add processing for EventSearchPerformed (e.g., log popular searches
// or build a search analytics feature in the future).
func (b *eventBus) run() {
	// "range b.eventsCh" reads events from the channel one at a time.
	// It blocks and waits when there are no events to process.
	for ev := range b.eventsCh {
		// Only car_viewed events update the view counter.
		// Other event types (like search) are received but not acted on yet.
		if ev.Type == EventCarViewed {
			viewCounts.Increment(ev.CarID)
		}
	}
}

// Publish sends an event to the event bus for asynchronous processing.
// If the event queue is full, the event is dropped and a warning is logged.
//
// Input:
// - event: the AppEvent to publish (includes type, car ID, query, source, timestamp)
//
// Side effects:
// - Sends the event to the background goroutine for processing.
// - Logs a warning if the queue is full and the event had to be dropped.
//
// Design decision:
// We use a non-blocking select with a default case so that
// HTTP handlers are never blocked waiting for the event queue.
// Dropping events under heavy load is acceptable because
// view counts are approximate metrics, not critical business data.
func (b *eventBus) Publish(event AppEvent) {
	select {
	// Try to send the event to the queue without blocking.
	case b.eventsCh <- event:
	default:
		// Keep request path non-blocking if event queue is full.
		// Log the dropped event so developers can detect heavy load situations.
		log.Printf("event bus full, dropping event type=%s carID=%d source=%s", event.Type, event.CarID, event.Source)
	}
}

// recordView publishes a car-viewed event for the given car.
// Call this whenever a user views a car's detail page.
//
// Input:
// - carID: the unique ID of the car that was viewed
// - source: a label describing where the view came from (e.g., "page:detail")
//
// Side effects:
// - Publishes an EventCarViewed event to the event bus.
//   The event bus will update the view counter in the background.
func recordView(carID int, source string) {
	appEvents.Publish(AppEvent{
		Type:      EventCarViewed,
		CarID:     carID,
		Source:    source,
		Timestamp: time.Now().UTC(),
	})
}

// recordSearch publishes a search event for the given query.
// Call this whenever a user performs a search.
//
// Input:
// - query: the search term the user entered (can be an empty string)
// - source: a label describing where the search came from (e.g., "api:search")
//
// Side effects:
// - Publishes an EventSearchPerformed event to the event bus.
//   Currently the event bus does not act on search events,
//   but they are recorded here for future use.
func recordSearch(query, source string) {
	appEvents.Publish(AppEvent{
		Type:      EventSearchPerformed,
		Query:     query,
		Source:    source,
		Timestamp: time.Now().UTC(),
	})
}
