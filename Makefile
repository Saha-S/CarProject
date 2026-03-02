APP=cars-viewer
DATA=../api/data.json
PORT?=8080

.PHONY: build run clean

build:
	go build -o $(APP) .

run:
	go run . $(DATA)

dev:
	PORT=$(PORT) go run . $(DATA)

clean:
	rm -f $(APP)
