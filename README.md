# MOTORIA — Cars Viewer

Cars Viewer is a Go web app that renders HTML templates and consumes car data from a separate API server.

## Project Structure

- `main.go` — app startup and route wiring
- `handlers_page.go` — page/template handlers
- `handlers_api.go` — app API endpoints
- `data.go` — loading data from external API + image proxy
- `templates/index.html` — main UI template
- `static/css/style.css` — styles
- `api/` — external data server (Node.js)
- `start.sh` — starts API + Go app together

## Requirements

- Go 1.21+
- Node.js 18+

## Run (Recommended)

From project root:

```bash
./start.sh
```

This starts:

- API server: `http://localhost:3000`
- Go app: `http://localhost:8080`

Stop both with `Ctrl+C`.

## Run Manually (Two Terminals)

Terminal 1 (API):

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

- `PORT` — Go app port (default `8080`)
- `API_BASE_URL` — API URL used by Go app (default `http://localhost:3000`)
- `API_PORT` — API server port when running `api/server.js` (default `3000`)

Examples:

```bash
PORT=9000 go run .
API_BASE_URL=http://localhost:3001 go run .
cd api && API_PORT=3001 npm start
```