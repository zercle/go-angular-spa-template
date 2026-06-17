# go-angular-spa-template

A full-stack template: an **Echo v5 (Go 1.26)** backend with **gRPC, OpenTelemetry/Prometheus, Valkey, and Postgres**, paired with an **Angular 22** SPA — shipped as a **single self-contained binary** (the Angular build is embedded into the Go binary via `embed.FS`).

The backend architecture mirrors [`zercle/zercle-go-template`](https://github.com/zercle/zercle-go-template): feature-sliced clean architecture, `samber/do` dependency injection, viper config, zerolog, golang-migrate, sqlc, and UUID identifiers. The example vertical slice is a `tasks` CRUD resource exercised end-to-end by the Angular UI.

## Stack

| Layer | Choice |
| --- | --- |
| HTTP / gRPC | Echo v5 + gRPC (`google.golang.org/grpc`) |
| DI | `samber/do/v2` |
| Database | Postgres 18 + sqlc (pgx/v5) |
| Migrations | golang-migrate (embedded, run via `cmd/migrate`) |
| Cache / messaging | Valkey |
| Observability | OpenTelemetry traces + Prometheus metrics, `/healthz` `/readyz` `/metrics` |
| Config | viper (`config.yaml` + env), validated |
| Logging | zerolog |
| Frontend | Angular 22 (standalone, signals, zoneless), Angular Material, Vitest, Bun |
| Serving | Angular built by Bun → embedded into the Go binary; Echo serves SPA + API |

## Layout

```
.
├── backend/                     # Go module (github.com/zercle/go-angular-spa-template)
│   ├── cmd/{server,migrate}/
│   ├── api/{proto,pb}/tasks/v1/ # gRPC contract + generated code
│   ├── internal/
│   │   ├── app/                 # DI composition root
│   │   ├── config/              # viper config
│   │   ├── features/tasks/      # domain/dto/service/repository/handler/di
│   │   ├── infrastructure/      # db (pgx, sqlc, migrations), messaging/valkey
│   │   ├── shared/              # errors, middleware, server, telemetry
│   │   └── web/                 # embeds the built Angular SPA + Echo route
│   ├── config.yaml  sqlc.yaml  Taskfile.yml  .golangci.yml
├── frontend/                    # Angular 22 app (Bun, Material)
├── compose.yml                  # postgres, valkey, migrate, server, otel, prometheus, grafana
├── Containerfile[.migrate]      # multi-stage builds (root context: bun → go → distroless)
├── Taskfile.yml                 # root orchestrator (includes backend/Taskfile.yml)
└── deployments/kustomize/       # k8s manifests
```

## Prerequisites

- Go 1.26+, Bun 1.3+, Docker/Podman
- [Task](https://taskfile.dev), and (for codegen) `protoc`, `golang-migrate`; backend Go tools via `task backend:tools`

## Quick start

```bash
cp backend/.env.example backend/.env     # defaults match compose
task setup                               # frontend deps (bun install)
task dev                                 # Postgres+Valkey, Echo :8080, Angular :4200
# open http://localhost:4200  (the dev server proxies /api → :8080)
```

Single self-contained binary:

```bash
task build      # bun builds Angular → embedded → backend/bin/server
task run        # open http://localhost:8080 (SPA + API from one process)
```

Ports: HTTP **:8080**, gRPC **:50051**. Health/observability: `/healthz`, `/readyz`, `/metrics`.

## API

REST under `/api/v1` (the browser uses REST; gRPC is for backend consumers). Success responses are the resource DTO; the list endpoint returns `{ "tasks": [...] }`; errors are `{ "error": "CODE", "message": "..." }`.

| Method | Path | Description |
| --- | --- | --- |
| GET | `/api/v1/tasks` | List tasks |
| POST | `/api/v1/tasks` | Create `{ "title": "..." }` |
| GET | `/api/v1/tasks/:id` | Get one |
| PUT | `/api/v1/tasks/:id` | Update `{ "title": "...", "done": bool }` |
| DELETE | `/api/v1/tasks/:id` | Delete |

gRPC: `tasks.v1.TaskService` (Create/Get/List/Update/Delete) on `:50051`.

## Common tasks

```bash
task dev                 # backend + frontend dev servers
task build               # single binary with SPA embedded
task test                # backend unit tests + frontend (Vitest)
task lint                # golangci-lint + eslint
task migrate             # apply DB migrations
task backend:generate    # sqlc + protobuf + mocks
task docker              # build the production image
```

## Deployment

```bash
docker build -f Containerfile -t go-angular-spa-template .
docker run -p 8080:8080 -p 50051:50051 --env-file backend/.env go-angular-spa-template
```

The image builds the Angular SPA with Bun, compiles a static Go binary with the SPA embedded, and ships a minimal distroless image. `compose.yml` additionally runs Postgres, Valkey, migrations, and an observability stack (OTel Collector, Prometheus, Grafana). Kubernetes manifests live under `deployments/kustomize/`.

## Replacing the example feature

The `tasks` feature is a deletable stub. To add your own: add a migration + sqlc queries (`task backend:generate`), define the domain/dto/service/repository/handler under `internal/features/<name>/`, and wire it in `internal/app/app.go`.

## Next steps (intentionally out of scope)

- **Authentication / authorization** — the tasks API is currently open.
- grpc-web/gateway if the browser needs to call gRPC directly.
