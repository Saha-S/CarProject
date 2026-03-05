#!/usr/bin/env bash
set -euo pipefail

API_PORT="${API_PORT:-3000}"
PORT="${PORT:-8080}"
API_BASE_URL="${API_BASE_URL:-http://localhost:${API_PORT}}"

if [[ ! -d "api/node_modules" ]]; then
  echo "Installing API dependencies..."
  (cd api && npm install)
fi

echo "Starting API server on http://localhost:${API_PORT}"
(
  cd api
  API_PORT="${API_PORT}" npm start
) &
API_PID=$!

cleanup() {
  if kill -0 "${API_PID}" >/dev/null 2>&1; then
    kill "${API_PID}" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT INT TERM

# Wait for API to be ready
echo "Waiting for API server to be ready..."
max_attempts=30
attempt=0
while [ $attempt -lt $max_attempts ]; do
  if curl -s "http://localhost:${API_PORT}/api/health" > /dev/null 2>&1; then
    echo "API server is ready!"
    break
  fi
  attempt=$((attempt + 1))
  if [ $attempt -eq $max_attempts ]; then
    echo "Error: API server did not start in time"
    exit 1
  fi
  sleep 1
done

echo "Starting Go app on http://localhost:${PORT} (API_BASE_URL=${API_BASE_URL})"
API_BASE_URL="${API_BASE_URL}" PORT="${PORT}" go run .
