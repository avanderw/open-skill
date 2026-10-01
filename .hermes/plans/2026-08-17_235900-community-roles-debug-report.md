# Remaining Community Roles, Debug Identity, and Report Match Plan

> **For Hermes:** Use this plan task-by-task with strict TDD and repository-hygiene practices.

**Goal:** Finish the three remaining product limitations: complete community admin roles and permissions, make the selected debug profile authoritative across the backend, and replace the legacy Report Match UI with a friendly team-count/search/assignment/review flow.

**Architecture:** Keep the current SvelteKit SPA and Go/SQLite service. Treat `communityId` and the active debug user identity as explicit request context. Keep authorization server-side even in the no-auth demo, with the backend accepting a controlled demo profile identifier rather than trusting UI-only state. Normalize report submissions into a single request model with explicit team count and participant assignments; retain backend-derived rating mode as an internal compatibility projection.

**Tech Stack:** SvelteKit 5 runes, TypeScript, Go `net/http`, SQLite via `modernc.org/sqlite`, existing Glicko-2 rating package.

---

## Current baseline and constraints

- Active remaining-work board: `COMMUNITY-UX-REMAINING-KANBAN.md`.
- Current backend has `communities.owner_id` and `community_members.role`, but role mutation and authorization are incomplete.
- Current backend uses fixed `u_demo` in `service/internal/api/api.go`.
- Browser debug profiles exist in `app/src/lib/debug-profiles.ts`, but the selected profile is not sent to or resolved by the backend.
- Report Match currently has legacy `reportMode` and raw player-ID-oriented state, with team-count groundwork but no complete drag/drop/search/review redesign.
- Do not introduce real OAuth, passwords, tokens, or external provider calls.
- Preserve community-scoped leaderboard, history, rating, and team-builder behavior.

---

# Track A — Community admin roles and permissions

## A1. Define role and authorization contract

**Objective:** Establish the exact role rules before changing endpoints.

**Files:**
- Modify: `service/internal/model/model.go`
- Modify: `service/internal/store/communities.go`
- Create: `service/internal/store/admin_roles_test.go`

**Rules:**

- Roles are `owner`, `admin`, `member`.
- `communities.owner_id` is the immutable owner identity.
- The owner is always treated as an administrator.
- Owner can promote/demote admins and remove members/admins.
- Admin can configure community settings and manage ordinary members.
- Admin cannot remove/demote the owner.
- Admin cannot promote/demote/remove another admin unless the acting user is the owner.
- Member and non-member cannot mutate settings, roles, integrations, or join requests.

**TDD:**

1. Write failing store tests for every permission rule.
2. Run:
   ```bash
   go test ./internal/store -run 'Test.*Role|Test.*Permission' -count=1
   ```
3. Implement a small role resolver such as:
   ```go
   func (s *Store) CommunityRole(ctx context.Context, communityID, userID string) (string, error)
   ```
4. Add a reusable authorization helper returning a typed forbidden error.
5. Re-run focused tests, then `go test ./...`.

**Acceptance:** Unauthorized operations return a typed forbidden error and never mutate SQLite.

## A2. Add role mutation store operations

**Objective:** Persist promote, demote, and remove actions transactionally.

**Files:**
- Modify: `service/internal/store/communities.go`
- Modify: `service/internal/store/schema.go` only if a constraint/index is required
- Test: `service/internal/store/admin_roles_test.go`

**Operations:**

- `PromoteCommunityAdmin(ctx, communityID, actorID, targetUserID)`
- `DemoteCommunityAdmin(ctx, communityID, actorID, targetUserID)`
- `RemoveCommunityMember(ctx, communityID, actorID, targetUserID)`
- `ListCommunityMembers(ctx, communityID, actorID)` returning users, player handles, role, owner flag, and join-request status.

**Implementation details:**

- Verify the community exists.
- Resolve actor role before mutation.
- Reject target owner mutation.
- Use SQLite transactions where membership and `community_players` must change together.
- Removing a member must remove their `community_players` rows but not global player ratings/history.
- Promote/demote must be idempotent and return the resulting member record.

## A3. Add authenticated API context for the demo

**Objective:** Ensure every protected mutation is authorized using an explicit active demo user.

**Files:**
- Modify: `service/internal/api/api.go`
- Modify: `service/internal/api/helpers.go` if present
- Modify: `app/src/lib/api.ts`
- Test: `service/internal/api/admin_api_test.go`

**Contract:**

- Accept a controlled header such as:
  ```text
  X-Demo-User: u_demo
  ```
- If absent, preserve current default `u_demo` behavior for compatibility.
- Only allow known demo users from the seeded/demo profile set.
- Never accept arbitrary user IDs without resolving them from `users`.
- Centralize request identity resolution so all handlers use the same actor.

**TDD:**

- Test missing header default.
- Test valid profile header.
- Test unknown profile rejection.
- Test member cannot call admin endpoint.
- Test admin cannot mutate owner.

## A4. Add admin endpoints and client methods

**Objective:** Expose role management and member listing over the API.

**Endpoints:**

```text
GET  /api/communities/{id}/members
POST /api/communities/{id}/members/{userId}/approve
POST /api/communities/{id}/members/{userId}/promote
POST /api/communities/{id}/members/{userId}/demote
DELETE /api/communities/{id}/members/{userId}
```

**Files:**
- Modify: `service/internal/api/api.go`
- Modify: `app/src/lib/api.ts`
- Test: `service/internal/api/admin_api_test.go`

**Response shape:**

```json
{
  "userId": "u_demo",
  "playerId": "p_001",
  "handle": "IronCrest",
  "role": "owner",
  "membership": "member",
  "isOwner": true
}
```

**Acceptance:** API returns `403` for unauthorized mutations, `404` for unknown community/user, `409` for invalid role transitions where useful, and a consistent JSON error body `{ "error": "..." }`.

## A5. Add owner/admin UI and settings cog

**Objective:** Make admin access visible and usable without exposing controls to members.

**Files:**
- Modify: `app/src/routes/leaderboard/+page.svelte`
- Modify: `app/src/routes/communities/[id]/settings/+page.svelte`
- Modify: `app/src/lib/api.ts`
- Optional create: `app/src/lib/components/AdminCog.svelte`, `MemberRow.svelte`

**UI behavior:**

- Show a labeled cog/button only when membership is `owner` or `admin`.
- Use accessible label `Open community settings`.
- Settings page Members tab shows role badges and owner badge.
- Owner row disables demote/remove controls with explanation.
- Owner sees promote/demote/remove actions according to rules.
- Admin sees ordinary-member management only.
- After mutation, refresh members and membership state from the API rather than relying only on optimistic state.
- Use `aria-live="polite"` for successful role changes and `role="alert"` for errors.

## A6. Admin-role verification

Run:

```bash
go test ./... -count=1
go vet ./...
npm run check
npm run build
```

Browser cases:

- Owner sees cog and all permitted controls.
- Admin sees cog but cannot demote/remove owner.
- Member sees no cog.
- Non-member sees join/request only.
- Owner remains present after all attempted destructive actions.

---

# Track B — Debug profile backend identity

## B1. Define demo profile fixtures

**Objective:** Make the three debug profiles real backend users/players with deterministic memberships.

**Profiles:**

| ID | Player | Role intent |
|---|---|---|
| `u_demo` | `p_001` / IronCrest | owner/admin of seeded communities |
| `u_member` | `p_member` / CommunityMember | member of one seeded community, non-member of another |
| `u_guest` | `p_guest` / GuestPlayer | non-member/pending test profile |

**Files:**
- Modify: `service/internal/seed/seed.go`
- Create: `service/internal/store/debug_profiles_test.go`
- Modify: `app/src/lib/debug-profiles.ts` only if fields differ

**Acceptance:** A clean temp database contains all three users and players without changing existing seeded leaderboard fixtures.

## B2. Thread active identity through the frontend client

**Objective:** Send the selected debug profile on every API request.

**Files:**
- Modify: `app/src/lib/api.ts`
- Modify: `app/src/lib/debug-profiles.ts`

**Implementation:**

- Add `activeDebugUserId()`.
- In `req()`, add:
  ```ts
  'X-Demo-User': activeDebugUserId()
  ```
- Keep the header absent or defaulted safely during SSR if localStorage is unavailable.
- Dispatch a custom event after switching profiles so active pages can refresh their data.

**Test:** Verify a client request includes the selected profile in a browser/CDP smoke script.

## B3. Resolve active identity in Go

**Objective:** Replace hard-coded `stubUser` behavior with controlled request identity.

**Files:**
- Modify: `service/internal/api/api.go`
- Modify: `service/internal/api/helpers.go`
- Test: `service/internal/api/debug_identity_test.go`

**Implementation:**

- Add `actorID(r *http.Request) (string, error)`.
- Validate `X-Demo-User` against `users`.
- Fall back to `u_demo` when the header is absent for backward compatibility.
- Ensure list communities, browse, membership, leaderboard `isYou`, create, join, report, history, and admin operations all use the actor.
- `/api/me` must return the active profile, not always IronCrest.

## B4. Add profile switcher UX

**Objective:** Make debug profile switching clearly visible and functional.

**Files:**
- Modify: `app/src/routes/+layout.svelte`
- Modify: `app/src/lib/debug-profiles.ts`

**Behavior:**

- Show `Debug profile` label in account menu.
- Show three profiles with labels.
- On selection, update localStorage, dispatch auth/profile event, close menu, and invalidate/reload current route data.
- Profile link uses the selected player ID.
- Keep a `Reset to IronCrest` action.
- Clearly mark this as development-only preview behavior.

## B5. Debug profile verification

Browser cases:

- IronCrest sees owner/admin cog on owned community.
- CommunityMember sees member state and no admin cog.
- GuestPlayer sees Join/Request action and no admin cog.
- Refresh preserves the selected profile.
- `/api/me` changes with the selected profile.
- My Communities and Browse membership labels change with the selected profile.

---

# Track C — Full Report Match redesign

## C1. Normalize the request model around team count

**Objective:** Remove the legacy user-facing mode concept while preserving backend rating projections.

**Files:**
- Modify: `service/internal/store/matches.go`
- Modify: `app/src/lib/api.ts`
- Test: `service/internal/store/report_team_count_test.go`

**Request model:**

```ts
interface ReportMatchRequest {
  communityId: string;
  teamCount: 1 | 2 | 3 | 4;
  participants: {
    playerId: string;
    teamNumber: number;
    score: number;
    finish: number | null;
  }[];
  notes?: string;
}
```

**Rules:**

- `teamCount=1`: placement/FFA projection; unique finish positions required.
- `teamCount>=2`: team projection; exactly `teamCount` numbered teams, even distribution, valid scores.
- Backend may derive internal `1v1` when teamCount=2 and each team has one player, otherwise `team`.
- Reject legacy/missing team-count requests after a compatibility migration period is explicitly tested.

## C2. Build friendly player-card primitives

**Objective:** Stop exposing raw IDs in selection/review UI.

**Files:**
- Create: `app/src/lib/components/PlayerCard.svelte`
- Create: `app/src/lib/components/PlayerPicker.svelte`
- Modify: `app/src/lib/api.ts`

**Player card must show:**

- Avatar/initials
- Handle
- Rating
- Match count
- Selected team/placement
- Remove/move action

**Accessibility:**

- Keyboard-operable selection and move controls.
- Drag/drop is an enhancement, not the only interaction.
- Clear `aria-label` values such as `Assign IronCrest to Team 2`.

## C3. Implement search and drag/drop assignment

**Objective:** Let users find players and assign them to numbered teams.

**Files:**
- Modify: `app/src/routes/report/+page.svelte`
- Create/use: `PlayerPicker.svelte`

**Behavior:**

- Search filters by friendly handle.
- Available players list excludes assigned players or clearly indicates assignment.
- Drag from available list into Team 1–4 buckets.
- Click/keyboard move controls provide equivalent assignment.
- Team 1 mode uses a placement list with move-up/move-down controls.
- Changing team count clears only invalid assignments and explains the reset.
- Prevent duplicate participant assignments.

## C4. Replace legacy mode UI and improve review

**Objective:** Make the report flow understandable to a non-developer.

**Files:**
- Modify: `app/src/routes/report/+page.svelte`

**UI sections:**

1. Community selector and metadata.
2. Number of teams selector.
3. Player search and available players.
4. Numbered team/placement columns.
5. Score/finish controls.
6. Friendly review panel.
7. Submit/confirm state.

**Review must show:**

- Community name
- Join mode and draw policy
- Team count/placement format
- Friendly handles and avatars
- Team assignments
- Scores or placements
- Plain-language explanation: `Ratings will update after submission.`

No raw IDs should appear except in an optional developer-only disclosure.

## C5. Redesign report result

**Objective:** Make success and errors actionable.

**Behavior:**

- Success shows `Match recorded` with community name, participants, rating changes, and timestamp.
- Links go to selected community leaderboard and selected player history.
- Failure preserves entered assignments and explains the validation error.
- Submit button is disabled while request is active.
- Result uses live-region semantics.

## C6. Report redesign verification

- Backend tests for team-count normalization, duplicate assignment, invalid team count, finish order, draw policy, and community scope.
- Browser tests for search, drag/drop, keyboard assignment, team count changes, friendly review, error recovery, and success links.
- Verify no visible `p_001`/`p_000` in normal client UI.

---

# Track D — Team Builder crash and context cleanup

## D1. Fix null data crash

**Files:**
- Modify: `app/src/routes/team/+page.ts`
- Modify: `app/src/routes/team/+page.svelte`
- Test: route smoke against `team?communityId=c_bd5fd2f791f1a953`

**Rules:**

- Normalize response arrays with `Array.isArray(...) ? value : []`.
- Never call `.slice`, `.filter`, or `.map` on nullable data.
- Invalid community → friendly error/empty state.
- Empty community → `No players available` state.

## D2. Contextual links and select theme

- Team back link preserves `communityId`.
- Report back/cancel links preserve `communityId`.
- All selects use dark background, text, border, option, and focus colors.
- Add route checks for the exact reported community URL.

---

# Track E — Final consistency and archive cleanup

## E1. Route/link matrix

Verify:

| Source | Expected target |
|---|---|
| `/` | `/communities` |
| My Communities card | `/leaderboard?communityId=...` |
| Browse card | `/leaderboard?communityId=...` |
| Leaderboard Settings cog | `/communities/{id}/settings` |
| Leaderboard Report | `/report?communityId=...` |
| Leaderboard Team Builder | `/team?communityId=...` |
| Report success leaderboard | selected community leaderboard |
| Team handoff report | selected community + assignments |
| Profile history | player + optional community |
| Settings back | selected community leaderboard |

## E2. Verification commands

```bash
cd service && go test ./... -count=1 && go vet ./...
cd ../app && npm run check && npm run build
cd .. && git diff --check -- app service COMMUNITY-UX-REMAINING-KANBAN.md
```

Browser matrix:

- Three debug profiles
- Open join persistence after refresh
- Approval pending state
- Owner/admin cog and protected settings
- Non-member join action
- Report team-count/search/assignment/review
- Team Builder exact reported URL
- Back/forward/refresh across all community-aware routes

## Archive rule

When R1–R5 are complete and verified:

- Keep `COMMUNITY-UX-REMAINING-KANBAN.md` as the completed historical summary.
- Archive this plan to `archive/` with the verification evidence.
- Do not delete prior boards without preserving a concise archive pointer.

---

Created 2026-08-17 for later execution.
