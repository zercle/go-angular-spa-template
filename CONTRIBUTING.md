# Contributing

Thanks for your interest in improving this template. This guide covers local
setup, the dev/build/test/lint workflow, and our PR expectations.

## Prerequisites

- **Go 1.26+**
- **Bun 1.3+** (frontend package manager / runner)
- **Docker or Podman** (compose stack, image builds)
- **[Task](https://taskfile.dev)** (task runner / orchestrator)
- `golang-migrate` and backend Go tools (sqlc, mockgen, air, golangci-lint) —
  install via `task backend:tools`

## Repository layout

- `backend/` — Go module `github.com/zercle/go-angular-spa-template` (Echo v5,
  feature-sliced clean architecture). REST-only; no gRPC.
- `frontend/` — Angular 22 SPA (Bun, Angular Material). Built and embedded into
  the Go binary via `embed.FS`.
- `deployments/kustomize/` — Kubernetes manifests (base + overlays).
- `compose.yml`, `Containerfile[.migrate]` — local stack and image builds.

## Local development

```bash
cp backend/.env.example backend/.env   # defaults match compose
task setup                             # install frontend deps (bun install)
task backend:tools                     # install backend Go tools

task dev      # Postgres + Valkey, Echo :8080, Angular dev server :4200
              # open http://localhost:4200 (dev server proxies /api -> :8080)
```

## Build, test, lint

```bash
task build    # bun builds Angular -> embedded -> backend/bin/server (single binary)
task run      # build + run the single binary at http://localhost:8080
task test     # backend unit tests + frontend (Vitest)
task lint     # golangci-lint + eslint
task migrate  # apply DB migrations
task backend:generate   # sqlc + mocks (run after changing queries/interfaces)
task docker   # build the production server image
```

Before opening a PR, make sure the following pass locally (they are also
enforced in CI):

- `task lint`
- `task test`
- `gofmt -s -l backend/` reports no files
- If you changed SQL queries or mocked interfaces, run `task backend:generate`
  and commit the regenerated code (CI fails if it is out of sync).

## Commit style

We use **[Conventional Commits](https://www.conventionalcommits.org/)**:

```
type(scope): short description

- optional bullet point details
```

Common types: `feat`, `fix`, `docs`, `refactor`, `chore`, `test`, `ci`, `build`,
`perf`. Example: `feat(tasks): add pagination to list endpoint`.

## Pull requests

- Keep PRs focused and reasonably small.
- Fill out the PR template (what/why, how tested, checklist).
- Ensure CI is green (lint, unit, integration, build/image, frontend, security).
- Update docs (`README.md`, this file) when behavior or workflow changes.
- New behavior should come with tests; the backend has a 60% coverage gate.

## Optional: pre-commit hooks

We provide a `.pre-commit-config.yaml`. To use it:

```bash
pip install pre-commit   # or: brew install pre-commit
pre-commit install
```

This runs formatting and basic hygiene checks before each commit.
