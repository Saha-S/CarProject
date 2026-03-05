package main

import "html/template"

type viewCounterStore struct {
	incrementCh chan int
	snapshotCh  chan viewCountSnapshotReq
}

type viewCountSnapshotReq struct {
	resp chan map[int]int
}

func newViewCounterStore() *viewCounterStore {
	store := &viewCounterStore{
		incrementCh: make(chan int),
		snapshotCh:  make(chan viewCountSnapshotReq),
	}

	go store.run()
	return store
}

func (s *viewCounterStore) run() {
	counts := make(map[int]int)

	for {
		select {
		case id := <-s.incrementCh:
			counts[id]++
		case req := <-s.snapshotCh:
			snapshot := make(map[int]int, len(counts))
			for id, c := range counts {
				snapshot[id] = c
			}
			req.resp <- snapshot
		}
	}
}

func (s *viewCounterStore) Increment(id int) {
	s.incrementCh <- id
}

func (s *viewCounterStore) Snapshot() map[int]int {
	resp := make(chan map[int]int)
	s.snapshotCh <- viewCountSnapshotReq{resp: resp}
	return <-resp
}

type Specifications struct {
	Engine       string `json:"engine"`
	Horsepower   int    `json:"horsepower"`
	Transmission string `json:"transmission"`
	Drivetrain   string `json:"drivetrain"`
}

type CarModel struct {
	ID             int            `json:"id"`
	Name           string         `json:"name"`
	ManufacturerID int            `json:"manufacturerId"`
	CategoryID     int            `json:"categoryId"`
	Year           int            `json:"year"`
	Specifications Specifications `json:"specifications"`
	Image          string         `json:"image"`
}

type Manufacturer struct {
	ID           int        `json:"id"`
	Name         string     `json:"name"`
	Country      string     `json:"country"`
	FoundingYear int        `json:"foundingYear"`
	Models       []CarModel `json:"models,omitempty"`
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Database struct {
	Manufacturers []Manufacturer `json:"manufacturers"`
	Categories    []Category     `json:"categories"`
	CarModels     []CarModel     `json:"carModels"`
}

var viewCounts = newViewCounterStore()
var appEvents = newEventBus(256)
var db Database
var tmpl *template.Template
var apiBaseURL string
