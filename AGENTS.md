# Repository Guidelines

## Project Structure & Module Organization

ServerDock pairs a Go API with a Vue single-page application. In `backend/`, `cmd/server` is the executable entry point; `internal/{config,dto,handler,model,pkg,router,service}` contains the application layers. Keep Go tests beside their packages as `*_test.go`.

In `frontend/src/`, route screens belong in `views/`, shared UI in `components/`, requests in `api/`, authentication in `stores/`, stateful helpers in `composables/`, and pure helpers in `utils/`. Root Docker files define deployment. Never commit runtime `data/`.

## Build, Test, and Development Commands

Run `mise install` for pinned Go, Node, and npm versions. If mise is not activated, prefix tool commands with `mise exec --`.

- `./scripts/deploy.sh <hostname-or-ip>`: generate secrets and a TLS certificate, then deploy the production stack.
- `docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build`: run the HTTPS development frontend (`:3000`) and backend (`:8000`).
- `cd backend && mise exec -- go run ./cmd/server`: run the API.
- `cd backend && mise exec -- go test ./...`: run all backend tests.
- `cd frontend && mise exec -- npm ci`: install locked frontend dependencies.
- `cd frontend && mise exec -- npm run dev`: start Vite.
- `cd frontend && mise exec -- npm run build`: build `frontend/dist/`.

## Coding Style & Naming Conventions

Format Go with `gofmt`; use PascalCase for exports and camelCase internally. Preserve handler-service-model boundaries and return errors explicitly. Vue uses two-space indentation, `<script setup>`, single quotes, and no semicolons. Name components in PascalCase (`StatusBadge.vue`), screens `*View.vue`, and composables `useX.js`. No repository-wide linter exists; match nearby code.

## Testing Guidelines

Backend tests use Go's `testing` package, usually with in-memory SQLite. Name tests `TestType_Behavior` and benchmarks `BenchmarkBehavior`. Cover changed service rules, handlers, authentication, and persistence. No frontend runner or coverage threshold exists; require a successful build and manually verify affected loading, empty, error, and mobile states.

## Commit & Pull Request Guidelines

Follow Conventional Commits: `feat(frontend): ...`, `fix(backend): ...`, `refactor: ...`, or `docs: ...`. Keep commits small and imperative. Pull requests must summarize behavior, list validation commands, note configuration or migration effects, link issues, and include before/after screenshots for UI changes.

## Security & Configuration

Copy `.env.example` locally. Never commit `.env`, credentials, JWT secrets, SSH keys, or databases. Preserve REST and WebSocket contracts unless changing both clients and handlers together.
