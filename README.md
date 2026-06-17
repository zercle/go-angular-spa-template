# go-angular-spa-template

A full-stack template: an **Echo v5 (Go 1.26)** backend with **OpenTelemetry/Prometheus, Valkey, and Postgres**, paired with an **Angular 22** SPA — shipped as a **single self-contained binary** (the Angular build is embedded into the Go binary via `embed.FS`).

The backend architecture mirrors [`zercle/zercle-go-template`](https://github.com/zercle/zercle-go-template): feature-sliced clean architecture, `samber/do` dependency injection, viper config, zerolog, golang-migrate, sqlc, and UUID identifiers. The example vertical slice is a `tasks` CRUD resource exercised end-to-end by the Angular UI. (gRPC, present in the upstream template, is intentionally omitted here — a browser SPA can't use it natively, so the API is REST-only.)

## Stack

| Layer | Choice |
| --- | --- |
| HTTP | Echo v5 |
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
- [Task](https://taskfile.dev) and `golang-migrate`; backend Go tools (sqlc, mockgen, air, golangci-lint) via `task backend:tools`

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

Port: HTTP **:8080**. Health/observability: `/healthz`, `/readyz`, `/metrics`.

## API

REST under `/api/v1`. Success responses are the resource DTO; the list endpoint returns `{ "tasks": [...] }`; errors are `{ "error": "CODE", "message": "..." }`.

| Method | Path | Description |
| --- | --- | --- |
| GET | `/api/v1/tasks` | List tasks |
| POST | `/api/v1/tasks` | Create `{ "title": "..." }` |
| GET | `/api/v1/tasks/:id` | Get one |
| PUT | `/api/v1/tasks/:id` | Update `{ "title": "...", "done": bool }` |
| DELETE | `/api/v1/tasks/:id` | Delete |

## Caching & observability

The tasks repository is wrapped by a **read-through Valkey cache** decorator (`repository.CachedRepository`) that also emits OpenTelemetry spans and Prometheus metrics — so every layer of the stack is exercised on the request path:

- **Valkey**: `Get`/`List` are served from cache on a hit and populated on a miss; writes evict the task key and bump a list-generation counter to invalidate cached pages.
- **OpenTelemetry**: spans `tasks.repository.{Create,GetByID,List,Update,Delete}` with a `cache.hit` attribute (exported when `OTEL_EXPORTER=otlp`).
- **Prometheus** (`/metrics`): `tasks_cache_hits_total`, `tasks_cache_misses_total` (labelled `op=get|list`), and `tasks_created_total`.

## Common tasks

```bash
task dev                 # backend + frontend dev servers
task build               # single binary with SPA embedded
task test                # backend unit tests + frontend (Vitest)
task lint                # golangci-lint + eslint
task migrate             # apply DB migrations
task backend:generate    # sqlc + mocks
task docker              # build the production image
```

## Deployment

```bash
docker build -f Containerfile -t go-angular-spa-template .
docker run -p 8080:8080 --env-file backend/.env go-angular-spa-template
```

The image builds the Angular SPA with Bun, compiles a static Go binary with the SPA embedded, and ships a minimal distroless image. `compose.yml` additionally runs Postgres, Valkey, migrations, and an observability stack (OTel Collector, Prometheus, Grafana).

### Kubernetes (Kustomize)

Manifests live under `deployments/kustomize/`:

```
deployments/kustomize/
├── base/                 # Deployment, Service, Ingress, HPA, ConfigMap, Namespace
└── overlays/
    ├── dev/              # 1 replica, dev image tag, debug logging
    └── prod/             # 3 replicas, ingress host, prod settings
```

The base ships a hardened single-binary Deployment: liveness (`/healthz`) and readiness (`/readyz`) probes on the HTTP port, CPU/memory requests and limits, a `runAsNonRoot` / `readOnlyRootFilesystem` securityContext with all capabilities dropped (matching the distroless `nonroot` UID `65532`), a CPU-based `HorizontalPodAutoscaler`, and a path-based `Ingress` routing `/` to the service (the SPA and API are one process).

Render and apply an overlay:

```bash
kubectl kustomize deployments/kustomize/overlays/prod    # render to stdout
kubectl apply -k deployments/kustomize/overlays/prod     # apply
```

Set the real image tag in your release pipeline (the overlays use `images[].newTag`) and the public host/TLS in the prod overlay's Ingress patch.

#### Secrets

The Deployment reads non-secret config from the `server-config` ConfigMap and **secrets from a `server-secrets` Secret** (`DB_PASSWORD`, `VALKEY_PASSWORD`). That Secret is intentionally **not** part of `kustomization.yaml` and must never be committed. Supply it out-of-band via:

- an **external secret manager** — External Secrets Operator, Sealed Secrets, Vault Agent, or a cloud CSI secret driver (recommended), or
- an **overlay patch** that injects values from your secret store.

`base/secret.example.yaml` documents the expected keys (example values only — do not apply it as-is).

## Replacing the example feature

The `tasks` feature is a deletable stub. To add your own: add a migration + sqlc queries (`task backend:generate`), define the domain/dto/service/repository/handler under `internal/features/<name>/`, and wire it in `internal/app/app.go`.

## Next steps (intentionally out of scope)

- **Authentication / authorization** — the tasks API is currently open.
