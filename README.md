# Open-Skill

Open-Skill is a competitive ranking app for communities of sport, esports, or any activity that can be ranked.

## Start Here

| Document | Use it for |
|---|---|
| [ARCHITECTURE.md](ARCHITECTURE.md) | Runtime boundaries, deployment, and data rules |
| [REFACTOR-RESEARCH.md](REFACTOR-RESEARCH.md) | Completed technical cleanup and remaining research |
| [INVESTIGATION-match-rating-model.md](INVESTIGATION-match-rating-model.md) | Rating and match decisions |
| [INVESTIGATION-community-membership.md](INVESTIGATION-community-membership.md) | Membership and eligibility rules |
| [INVESTIGATION-identity-auth.md](INVESTIGATION-identity-auth.md) | Identity and authentication decisions |
| [INVESTIGATION-settings-roles.md](INVESTIGATION-settings-roles.md) | Community roles and permissions |
| [INVESTIGATION-api-client-contract.md](INVESTIGATION-api-client-contract.md) | Frontend API contract decisions |

Migration plans and Kanban files remain as implementation history:

- [2026-09-09-community-scoped-openskill-migration.md](2026-09-09-community-scoped-openskill-migration.md)
- [MIGRATION-community-membership.md](MIGRATION-community-membership.md)
- [MIGRATION-api-client-contract.md](MIGRATION-api-client-contract.md)
- [COMMUNITY-UX-REMAINING-KANBAN.md](COMMUNITY-UX-REMAINING-KANBAN.md)
- [kanban-mockup-v1.md](kanban-mockup-v1.md)

## Repository Map

| Path | Purpose |
|---|---|
| `app/` | SvelteKit 5 frontend, built as a static site |
| `service/` | Go HTTP API, SQLite store, membership, matches, and ratings |
| `design/` | Exported design system and page references |
| `mockup/` | Early linked HTML prototypes |

## Current Technical Decisions

- Development `/api` requests use the Vite proxy; production uses nginx to proxy to the Go service.
- Set `VITE_API_URL` at build time (see `app/.env.example`) to call a separately hosted service directly; leave it empty for same-origin `/api`. The service must list the app origin in `OPEN_SKILL_ALLOWED_ORIGINS`.
- SQLite uses the stdlib-compatible `modernc.org/sqlite` driver.
- Set `OPEN_SKILL_DB_URL` and `OPEN_SKILL_DB_TOKEN` to use a remote Turso (libSQL) database instead of the local `-db` file.
- The app signs in with Firebase Auth (Google/GitHub) and sends `Authorization: Bearer <id token>`; config comes from `VITE_FIREBASE_*` (see `app/.env.example`). Dev builds can set `VITE_DEMO_AUTH=1` to use the demo identity header instead. The service verifies tokens for `OPEN_SKILL_FIREBASE_PROJECT_ID` and only trusts `X-Demo-User` when `OPEN_SKILL_DEMO_AUTH=1` (local development only); it refuses to start with neither set.
- Rating mode is derived from submitted participants: `1v1`, `team`, or `ffa`.
- Ratings are scoped by `community_id`, `player_id`, and mode.
- A participant is eligible when the owning user is a member of the reporting community.
- Community leaderboard reads return the top rows plus the current user's row when available.

## Validation

- Backend: `cd service && go test ./...`
- Frontend: `cd app && npm test -- --run src/lib/api.test.ts`
- Frontend checks: `cd app && npm run check`
- Compose syntax: `docker-compose config`