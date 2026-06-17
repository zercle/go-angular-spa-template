# syntax=docker/dockerfile:1

# --- Stage 1: build the Angular SPA with Bun ---
FROM oven/bun:1.3.8 AS frontend
WORKDIR /app/frontend
COPY frontend/package.json frontend/bun.lock ./
RUN bun install --frozen-lockfile
COPY frontend/ ./
# Outputs into /app/backend/web/dist (see angular.json outputPath).
RUN bun run build

# --- Stage 2: build the Go binary with the SPA embedded ---
FROM golang:1.26 AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY backend/ ./backend/
# Bring in the built SPA so //go:embed picks it up.
COPY --from=frontend /app/backend/web/dist ./backend/web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./backend/cmd/server

# --- Stage 3: minimal runtime image ---
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=backend /server /server
EXPOSE 3000
USER nonroot:nonroot
ENTRYPOINT ["/server"]
