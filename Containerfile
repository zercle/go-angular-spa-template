# syntax=docker/dockerfile:1
# Build context is the repository root so both frontend/ and backend/ are available.

# -----------------------------------------------------------------------------
# Stage 1: build the Angular SPA with Bun (on Node 26 — Angular CLI 22 requires
# Node >= 24.15/26; the oven/bun image ships an older bundled Node).
# -----------------------------------------------------------------------------
FROM node:26-slim AS frontend
RUN npm install -g bun@1.3.8
WORKDIR /app/frontend
COPY frontend/package.json frontend/bun.lock ./
RUN bun install --frozen-lockfile
COPY frontend/ ./
# angular.json outputPath is ../backend/internal/web/dist -> /app/backend/internal/web/dist
RUN bun run build

# -----------------------------------------------------------------------------
# Stage 2: build the Go binary with the SPA embedded
# -----------------------------------------------------------------------------
FROM golang:1.26 AS builder
WORKDIR /build
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
# Bring in the built SPA so //go:embed picks it up.
COPY --from=frontend /app/backend/internal/web/dist ./internal/web/dist

ARG VERSION=dev
ARG COMMIT_SHA=unknown
ARG BUILD_TIME=unknown
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.CommitSHA=${COMMIT_SHA} -X main.BuildTime=${BUILD_TIME}" \
    -o /server ./cmd/server

# -----------------------------------------------------------------------------
# Stage 3: minimal runtime image
# -----------------------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder --chown=nonroot:nonroot /server /server
COPY --from=builder --chown=nonroot:nonroot /build/config.yaml /config.yaml
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
