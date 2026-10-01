# Architecture — Open-Skill

## Repo layout

| Path | Owner | Purpose |
|---|---|---|
| `app/` | SvelteKit web app (static adapter) | The player-facing UI |
| `design/` | Figma `design` exports (dc-html) + Capitec design system + scripts + design tokens | Source of truth for theme & components |
| `mockup/` | Hand-linked dark-theme HTML prototypes (12 pages) | Pagerless spec for Svelte port; ALSO the homepage for the visual design |
| `service/` | Go backend | Ranking API, match reporting, SQLite persistence, auth stub |
| Root `MIGRATION-*.md` and `kanban-*.md` | Planning | Historical implementation plans |

## Ports/contracts

- `app` speaks **SvelteKit**; build static via adapter-static (`prerender=false`).
- `app` calls `service` at **`/api/**`** — in dev, SvelteKit `server.proxy` (vite.config.ts) proxies `/api` → `localhost:8080`; in prod, nginx serves `app/build` and reverse-proxies `/api` → `service:8080`.
- `VITE_API_URL` (build-time, see `app/.env.example`) sets the service origin prefixed to `/api/**`. Empty (default) keeps requests same-origin; set it when the app and service are hosted separately (e.g. static host + Cloud Run), and add the app origin to `OPEN_SKILL_ALLOWED_ORIGINS`.

## Service

- `service/go.mod`
- `service/cmd/open-skill-service/main.go` — `:8080` entrypoint
- `service/internal/api` — HTTP handlers and demo identity headers
- `service/internal/store` — SQLite schema, communities, history, matches, and ratings
- SQLite via `modernc.org/sqlite` (stdlib driver, no cgo)

## Runtime Decisions

- Development `/api` requests use the Vite proxy to `localhost:8080`.
- Production nginx serves `app/build` and proxies `/api` to `service:8080`.
- Separately hosted app builds set `VITE_API_URL` to the service origin and call it cross-origin.
- The Go module path is `open-skill/service`.
- SQLite uses `modernc.org/sqlite` without cgo.
- Setting `OPEN_SKILL_DB_URL` (e.g. `libsql://...turso.io`) with `OPEN_SKILL_DB_TOKEN` switches the store to a remote libSQL/Turso database via `libsql-client-go` (also cgo-free); otherwise the `-db` file is used.
- Invite codes use the documented human-readable code generator in the API.
- The app authenticates with Firebase Auth and sends an ID token as `Authorization: Bearer`; `VITE_DEMO_AUTH=1` (dev builds only) sends the `X-Demo-User` demo identity instead. The service verifies Firebase ID tokens (`OPEN_SKILL_FIREBASE_PROJECT_ID`) and maps them to users `fb_<uid>`, created on first sign-in; `X-Demo-User` is only trusted with `OPEN_SKILL_DEMO_AUTH=1`. Leaderboard, community lookup/membership and browse allow anonymous callers; other endpoints require sign-in. Allowed CORS origins are configured with `OPEN_SKILL_ALLOWED_ORIGINS`.

## Domain Rules

- Ratings are keyed by `community_id`, `player_id`, and mode.
- Supported modes are `1v1`, `team`, and `ffa`.
- Mode is derived from participant/team configuration when a match is reported.
- A player is eligible when the player's owning user is in `community_members` for the match community.
- The match store validates reporter membership and participant eligibility inside the write transaction.
- There is no separate `community_players` projection.

Open questions and historical implementation details live in the investigation and migration documents linked from [README.md](README.md).