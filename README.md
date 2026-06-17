# GoFiber + Angular SPA — Template

A clean, modern full-stack template pairing a **[GoFiber v3](https://gofiber.io/) (Go 1.26)** backend with an **[Angular 22](https://angular.dev/)** SPA, deployable as a **single self-contained binary**. It ships one end-to-end vertical slice — a `tasks` CRUD resource — that exercises every layer so you can clone it and start building.

## Stack

| Layer | Choice |
| --- | --- |
| Backend | GoFiber v3, clean architecture (domain / usecase / delivery / infrastructure) |
| Database | Postgres + [sqlc](https://sqlc.dev/) (type-safe queries) + [pgx/v5](https://github.com/jackc/pgx) |
| Migrations | [goose](https://github.com/pressly/goose) — embedded, auto-applied on startup |
| Frontend | Angular 22 (standalone, signals, zoneless), Angular Material, Vitest |
| Tooling | [Bun](https://bun.sh/) (frontend), [air](https://github.com/air-verse/air) (Go hot reload), Make, Docker |
| Serving | Angular built by Bun → embedded into the Go binary via `embed.FS` |

## Architecture

In **development**, two processes run: Angular's dev server (`:4200`) proxies `/api` to Fiber (`:3000`).
In **production**, `bun run build` emits the SPA into `backend/web/dist/`, Go embeds it with `embed.FS`, and the single binary serves both the SPA and the API.

```
Browser ──▶ Fiber (:3000)
              ├── /api/v1/*  → handler → service → repository (sqlc) → Postgres
              └── /*         → embedded Angular SPA (index.html fallback for client routes)
```

## Prerequisites

- Go 1.26+, Bun 1.3+, Docker (for Postgres)
- Dev tools: `make tools` (installs air, sqlc, goose, golangci-lint)

## Quick start

```bash
cp .env.example .env          # defaults work with the bundled docker-compose
make tools                    # one-time: install Go dev tools
make db-up                    # start Postgres
make dev                      # Fiber (:3000) + Angular (:4200)
# open http://localhost:4200
```

To run the production single binary:

```bash
make run                      # builds the SPA, embeds it, runs the binary
# open http://localhost:3000
```

## Project layout

```
backend/
  cmd/server/            # composition root (main.go)
  internal/
    config/              # env config
    domain/              # entities + repository interface (no framework imports)
    usecase/             # business logic (depends only on domain)
    delivery/rest/       # Fiber v3 handlers + SPA serving
    repository/postgres/ # sqlc queries, goose migrations, pgx adapter
  web/                   # embed.FS for the built SPA
frontend/                # Angular 22 app (Bun)
  src/app/core/          # Task model + signal-based TaskStore
  src/app/features/tasks # task-list + task-form (Material)
sqlc.yaml, Makefile, docker-compose.yml, Dockerfile, .github/workflows/ci.yml
```

## API

Base path `/api/v1`. Success responses are wrapped as `{ "data": ... }`; errors as `{ "error": "..." }`.

| Method | Path | Description |
| --- | --- | --- |
| GET | `/api/v1/health` | Health check |
| GET | `/api/v1/tasks` | List tasks |
| POST | `/api/v1/tasks` | Create `{ "title": "..." }` |
| GET | `/api/v1/tasks/:id` | Get one |
| PUT | `/api/v1/tasks/:id` | Update `{ "title": "...", "done": bool }` |
| DELETE | `/api/v1/tasks/:id` | Delete |

## Common commands

```bash
make dev          # run backend + frontend for development
make build        # build the single self-contained binary
make test         # backend (go test) + frontend (Vitest)
make lint         # golangci-lint + eslint
make sqlc         # regenerate sqlc code after editing queries/migrations
make migrate      # apply migrations manually (also auto-run on startup)
make docker       # build the production image
```

## Deployment

```bash
docker build -t gofiber-angular-spa .
docker run -p 3000:3000 -e DATABASE_URL=postgres://... gofiber-angular-spa
```

The multi-stage `Dockerfile` builds the SPA with Bun, compiles a static Go binary with the SPA embedded, and ships a minimal distroless image.

## Adding a new resource

1. Add a goose migration in `backend/internal/repository/postgres/migrations/`.
2. Add queries in `.../queries/` and run `make sqlc`.
3. Define the entity + repository interface in `domain/`, the logic in `usecase/`, the adapter in `repository/postgres/`, and handlers in `delivery/rest/`.
4. On the frontend, generate artifacts with the CLI (`bunx ng generate ...`) and wire a signal store + components.

## Next steps (intentionally out of scope)

- **Authentication / authorization** (JWT or sessions) — the tasks API is currently open.
- Pagination, request rate limiting, structured error codes, OpenAPI docs.
