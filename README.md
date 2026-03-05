# MOTORIA - Cars Viewer

MOTORIA is a Go web application that renders car data in HTML views and also exposes JSON endpoints.
It loads data from a separate Node.js API service.

## Usage Guide

Once the app is running on `http://localhost:8080`, you can explore these features:

### **🚗 Fleet Gallery** (Default View)
Browse all available car models in a grid layout with filtering options:
- **Search**: Find cars by model name, engine type, or drivetrain
- **Filter by Category**: Select vehicle type (e.g., SUV, Sedan, Truck)
- **Filter by Manufacturer**: Choose a specific car maker
- **Filter by Horsepower**: Set min/max HP range to narrow results
- **Sort Options**: Sort by horsepower, year, or name
- Click any car to view full details

### **📊 Compare** 
Compare up to 3 vehicles side-by-side:
- Select 3 vehicles from the dropdowns
- View all specifications in a detailed comparison table
- See highlighted best values for easy comparison

### **🏭 Manufacturers**
Explore all car manufacturers:
- Browse a complete list of all manufacturers
- View info: Country and founding year
- See all models from each manufacturer with links to details

### **🎯 Personalized Recommendations**
Get cars tailored to your preferences:
- Select preferred category (optional)
- Set minimum horsepower requirement
- Set maximum horsepower budget
- Get ranked recommendations based on your filters

### **📋 Car Details**
View complete information about any vehicle:
- Full specifications (engine, transmission, drivetrain)
- Horsepower and year
- Manufacturer details
- High-quality car image

## How It Works

- The Go app starts on `PORT` (default `8080`).
- On startup, it fetches and caches manufacturers, categories, and car models from `API_BASE_URL`.
- UI pages are rendered from templates.
- API endpoints provide JSON for models, search, compare, and recommendations.

## Async Event System

The project now includes an asynchronous event pipeline:

- HTTP handlers create events such as `car_viewed` and `search_performed`.
- Handlers publish events into a buffered channel (`eventBus.eventsCh`).
- A dedicated background goroutine consumes events and processes them.
- View counts are updated in the background event consumer, not directly inside handlers.

This keeps request handlers lightweight and avoids shared-state writes from multiple request goroutines.

## Concurrency Safety

- `viewCounts` is managed by its own goroutine via channels in `viewCounterStore`.
- Request handlers never write directly to a shared map.
- Recommendations read counts via a snapshot channel API.
- Use `go test -race ./...` to check for race conditions during development.

## Project Structure

- `main.go`: app startup, template parsing, route wiring
- `handlers_page.go`: page handlers and template data assembly
- `handlers_api.go`: JSON API handlers
- `event_system.go`: async event types, bus, and publish helpers
- `models.go`: domain models and concurrent view counter store
- `data.go`: startup data loading and image proxy logic
- `helpers.go`: helper utilities and panic recovery middleware
- `templates/`: HTML templates
- `static/`: static assets (CSS, images)
- `api/`: external Node.js API service
- `start.sh`: starts API + Go app together

## Requirements

- Go 1.21+
- Node.js 18+

## Run (Recommended)

From the project root:

```bash
./start.sh
```

This starts:

- API server: `http://localhost:3000`
- Go app: `http://localhost:8080`

Stop both with `Ctrl+C`.

## Run Manually (Two Terminals)

Terminal 1 (API service):

```bash
cd api
npm install
npm start
```

Terminal 2 (Go app):

```bash
go run .
```

## Environment Variables

- `PORT`: Go app port (default `8080`)
- `API_BASE_URL`: base URL for data source API (default `http://localhost:3000`)
- `API_PORT`: API service port for `api/server.js` (default `3000`)

Examples:

```bash
PORT=9000 go run .
API_BASE_URL=http://localhost:3001 go run .
cd api && API_PORT=3001 npm start
```

## Main App Routes

- `GET /`: main page (supports gallery/detail/compare/recommendations views via query params)
- `GET /static/*`: static files
- `GET /static/img/*`: proxied remote images

## Main JSON Endpoints

- `GET /api/models`
- `GET /api/models/{id}`
- `GET /api/manufacturers`
- `GET /api/manufacturers/{id}`
- `GET /api/categories`
- `GET /api/categories/{id}`
- `GET /api/search`
- `GET /api/compare`
- `GET /api/recommendations`

## Troubleshooting

- If `./start.sh` fails, confirm Node and Go are installed and on `PATH`.
- If the API port is busy, run with another port:

```bash
API_PORT=3001 API_BASE_URL=http://localhost:3001 ./start.sh
```

- If static pages open but data is missing, verify the API server is running and reachable.