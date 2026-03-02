# MOTORIA — Cars Viewer

A sleek, production-quality car explorer built with **Go** (backend) and vanilla **HTML/CSS/JS** (frontend).

---

## Project Overview

MOTORIA fetches and displays car model data from a provided Cars API resource. It features:

- **Fleet Gallery** — browse all 10 car models with live search & filtering
- **Car Detail Modal** — click any card to load full specs + manufacturer info from the server
- **Side-by-Side Comparison** — compare up to 3 vehicles with highlighted best/worst values
- **Manufacturers View** — browse makers and see their model lineup
- **Personalised Recommendations** — get AI-scored suggestions based on your HP/category preferences and past views

---

## Setup & Installation

### Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- The Cars API (`api/` folder from the original zip, running separately OR just its `data.json`)

### Installation

```bash
# Clone/copy this project directory
cd cars-viewer

# Build the binary
go build -o cars-viewer .

# Run (pointing to the data.json from the Cars API)
./cars-viewer ../api/data.json
```

Or use the Makefile:

```bash
make run        # uses ../api/data.json on port 8080
make dev        # same, with custom PORT
PORT=9000 make dev
```

Then open: **http://localhost:8080**

---

## Project Structure

```
cars-viewer/
├── main.go              # Go backend — routes, handlers, data loading
├── go.mod               # Module file (no external dependencies)
├── Makefile             # Build & run shortcuts
├── templates/
│   └── index.html       # Main HTML template
├── static/
│   ├── css/style.css    # All styles (industrial-luxury aesthetic)
│   ├── js/app.js        # Frontend logic (search, modal, compare, recs)
│   └── img/             # Car images (copy from api/img/)
└── README.md
```

> **Note:** Copy the `img/` folder from the API's directory into `static/img/`:
> ```bash
> cp -r ../api/img/ static/img/
> ```

---

## Usage Guide

### Fleet (Gallery)
- Browse all car models as cards with key specs
- **Search** by name, engine, manufacturer, or drivetrain
- **Filter** by category, manufacturer, sort order
- **HP slider** to filter by horsepower range
- Click any card to open the detail modal (fetches fresh data from server)

### Compare
- Click **+ COMPARE** on any card to add it to the comparison queue (up to 3)
- Switch to the **COMPARE** tab and click **COMPARE NOW**
- The table highlights the best value (green) and worst (red) for HP and year

### Makers
- Click any manufacturer to see their profile and models
- Click a model to open its detail modal

### For You (Recommendations)
- Set your preferred category and HP range
- Click **FIND MY CARS** to get personalised, ranked suggestions
- Score is based on: category match, HP range fit, model year, and view history

---

## API Endpoints (served by Go)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/models` | All car models |
| GET | `/api/models/{id}` | Single car with manufacturer & category |
| GET | `/api/manufacturers` | All manufacturers |
| GET | `/api/manufacturers/{id}` | Manufacturer + their models |
| GET | `/api/categories` | All categories |
| GET | `/api/categories/{id}` | Category + their models |
| GET | `/api/search` | Filtered/sorted car search |
| GET | `/api/compare?ids=1,2,3` | Compare multiple cars |
| GET | `/api/recommendations` | Personalised recommendations |

### Search query params
- `q` — text search
- `category` — category name
- `manufacturer` — manufacturer name
- `minHP` / `maxHP` — horsepower range
- `minYear` / `maxYear` — year range
- `sort` — `hp_desc`, `hp_asc`, `year_desc`, `year_asc`, `name`

---

## Extra Features Implemented

- ✅ **Advanced Filtering** — full-text search + category/manufacturer/HP/year/sort filters
- ✅ **Comparisons** — side-by-side table with up to 3 cars, visual best/worst highlighting
- ✅ **Personalised Recommendations** — scored algorithm using view history, category match, HP fit, and year

---

## Technology Stack

- **Backend**: Go standard library only (`net/http`, `encoding/json`, `html/template`) — zero external dependencies
- **Frontend**: Vanilla HTML + CSS + JS — no frameworks, no build step
- **Fonts**: Bebas Neue (display), DM Sans (body), DM Mono (data) via Google Fonts
- **Data**: JSON served from Go's in-memory parsed data structure
