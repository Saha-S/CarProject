# MOTORIA — Cars Viewer

A car explorer web app built with Go (backend) and vanilla HTML/CSS/JS (frontend).

## Overview

The app loads car data from a local `data.json` file and serves:

- Fleet gallery with search and filters
- Car detail modal with manufacturer and category details
- Side-by-side comparison (up to 3 cars)
- Manufacturer profiles with their models
- Preference-based recommendations

## Requirements

- Go 1.21+
- A valid data file (default: `data.json` in this repository)
- Optional: car images inside `static/img/`

## Run

From the project root:

```bash
go run .
```

The server starts on `http://localhost:8080`.

### Optional run modes

```bash
# Run with a custom data file path
go run . ./data.json

# Run on a custom port
PORT=9000 go run .

# Build and run binary
go build -o cars-viewer .
./cars-viewer
```

## Project Structure

```
.
├── main.go               # bootstrap: startup, template load, route wiring
├── models.go             # data models + shared app state
├── data.go               # data.json loading
├── helpers.go            # lookup helpers (car/manufacturer/category)
├── handlers_api.go       # all /api handlers
├── handlers_page.go      # page/template handlers
├── go.mod
├── data.json
├── templates/
│   └── index.html
└── static/
    ├── css/style.css
    ├── js/app.js
    └── img/
```

### Backend file layout

- `main.go`: app entrypoint and HTTP route registration
- `models.go`: structs (`CarModel`, `Manufacturer`, `Category`) and globals
- `data.go`: `loadData()` for JSON parsing
- `helpers.go`: reusable ID lookup helpers
- `handlers_api.go`: JSON API endpoints (`/api/*`)
- `handlers_page.go`: HTML page handler (`/`)

## API Endpoints

| Method | Path | Description |
|---|---|---|
| GET | `/api/models` | List all car models |
| GET | `/api/models/{id}` | Get one model with manufacturer and category |
| GET | `/api/manufacturers` | List manufacturers |
| GET | `/api/manufacturers/{id}` | Manufacturer with its models |
| GET | `/api/categories` | List categories |
| GET | `/api/categories/{id}` | Category with its models |
| GET | `/api/search` | Search/filter/sort models |
| GET | `/api/compare?ids=1,2,3` | Compare selected models |
| GET | `/api/recommendations` | Get top recommendations |

### `/api/search` query params

- `q`
- `category`
- `manufacturer`
- `minHP`, `maxHP`
- `minYear`, `maxYear`
- `sort`: `hp_desc`, `hp_asc`, `year_desc`, `year_asc`, `name`

## Notes

- If images are missing in `static/img/`, the UI still works and shows placeholders.
