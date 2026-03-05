package main

import (
	"log"
	"time"
)

type AppEventType string

const EventCarViewed AppEventType = "car_viewed"
const EventSearchPerformed AppEventType = "search_performed"

type AppEvent struct {
	Type      AppEventType
	CarID     int
	Query     string
	Source    string
	Timestamp time.Time
}

type eventBus struct {
	eventsCh chan AppEvent
}

func newEventBus(buffer int) *eventBus {
	bus := &eventBus{
		eventsCh: make(chan AppEvent, buffer),
	}

	go bus.run()
	return bus
}

func (b *eventBus) run() {
	for ev := range b.eventsCh {
		if ev.Type == EventCarViewed {
			viewCounts.Increment(ev.CarID)
		}
	}
}

func (b *eventBus) Publish(event AppEvent) {
	select {
	case b.eventsCh <- event:
	default:
		// Keep request path non-blocking if event queue is full.
		log.Printf("event bus full, dropping event type=%s carID=%d source=%s", event.Type, event.CarID, event.Source)
	}
}

func recordView(carID int, source string) {
	appEvents.Publish(AppEvent{
		Type:      EventCarViewed,
		CarID:     carID,
		Source:    source,
		Timestamp: time.Now().UTC(),
	})
}

func recordSearch(query, source string) {
	appEvents.Publish(AppEvent{
		Type:      EventSearchPerformed,
		Query:     query,
		Source:    source,
		Timestamp: time.Now().UTC(),
	})
}
