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
- SQLite uses the stdlib-compatible `modernc.org/sqlite` driver.
- Authentication is a demo identity stub; real authentication is not implemented.
- Rating mode is derived from submitted participants: `1v1`, `team`, or `ffa`.
- Ratings are scoped by `community_id`, `player_id`, and mode.
- A participant is eligible when the owning user is a member of the reporting community.
- Community leaderboard reads return the top rows plus the current user's row when available.

## Validation

- Backend: `cd service && go test ./...`
- Frontend: `cd app && npm test -- --run src/lib/api.test.ts`
- Frontend checks: `cd app && npm run check`
- Compose syntax: `docker-compose config`