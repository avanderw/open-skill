# Deployment Tasks

Target: static app on GitHub Pages (or Cloudflare Pages) → Go service on Cloud Run → Turso DB, with real auth.
Each task is independently testable; complete in order where dependencies are noted.

---

## 1. Configurable API base URL (app)

- [x] Add `VITE_API_URL` (default empty → relative `/api` for dev proxy)
- [x] Prefix request paths in `req()` in `app/src/lib/api.ts`
- [x] Add `app/.env.example` documenting `VITE_API_URL`

**Test**
- `npm run dev` with no env → requests still go through Vite proxy to `localhost:8080`
- `VITE_API_URL=http://localhost:8080 npm run build && npm run preview` → network tab shows requests hitting `localhost:8080` directly (CORS succeeds)
- `npm test` passes

---

## 2. GitHub Pages compatibility (app)

- [ ] Add empty `app/static/.nojekyll` (Pages/Jekyll ignores `_app/`)
- [ ] Change adapter `fallback` to `404.html` so deep links load
- [ ] Decide hosting URL:
  - [ ] Custom domain or `<user>.github.io` repo → no base path needed, **or**
  - [ ] Project page `<user>.github.io/open-skill` → set `kit.paths.base` from env and replace hard-coded `href="/..."` / `goto('/...')` with `base`-prefixed paths (incl. `auth.ts` redirects)

**Test**
- `npm run build` → `build/.nojekyll` and `build/404.html` exist
- Serve `build/` under the base path locally; navigate directly to `/profile/<id>` and `/leaderboard` → page loads, links stay under base

---

## 3. GitHub Actions deploy for app

- [ ] Add `.github/workflows/deploy-app.yml` (build `app/`, `actions/upload-pages-artifact`, `actions/deploy-pages`)
- [ ] Pass `VITE_API_URL` from a repo variable
- [ ] Enable Pages → Source: GitHub Actions in repo settings

**Test**
- Push to `main` → workflow green → site loads at Pages URL
- Deep link refresh works (depends on task 2)

---

## 4. Turso database support (service)

- [x] Add `github.com/tursodatabase/libsql-client-go`
- [x] `store.OpenRemote` used when `OPEN_SKILL_DB_URL` is set (token from `OPEN_SKILL_DB_TOKEN`); otherwise local `modernc` SQLite file via `-db`
- [x] Ensure schema migration runs against Turso (applied on startup; verified against `open-skill` DB)
- [x] Enable `foreign_keys` on each remote connection (verify with gated test below)
- [x] Review hot paths for N+1 queries (each query is now a network round-trip)
  - Match replay (edit/delete) batched: ~310 → 10–13 round trips for 30 matches, constant in match count
  - Report match: 19 → 8; membership/role checks: 2–7 → 1–2
  - Budgets enforced by `TestRoundTripBudgets` (`internal/store/roundtrips_test.go`)
  - Per-request `EnsureDemoIdentity` left for task 6 (replaced by Firebase sign-in)

**Test**
- `go test ./...` still passes against local SQLite (remote test skips)
- Copy `service/.env.example` to `service/.env` (git-ignored), paste the dev token, then from `service/`:
  - `./with-env.ps1 go test ./internal/store -run Remote -v` → passes (FKs enforced)
  - `./with-env.ps1 go run ./cmd/open-skill-service -seed` → app works against it; restart without `-seed` → data persists
- Verify in the Turso dashboard SQL console: `select count(*) from players`

---

## 5. Cloud Run deployment (service)

- [ ] Confirm `service/Dockerfile` builds a small static image and listens on `$PORT` (default 8080)
- [ ] Create GCP project, enable billing + **budget alert**
- [ ] Store `OPEN_SKILL_DB_TOKEN` in Secret Manager
- [ ] Deploy: `gcloud run deploy open-skill-service --source service --region europe-west1 --allow-unauthenticated --set-env-vars OPEN_SKILL_DB_URL=...,OPEN_SKILL_ALLOWED_ORIGINS=https://<user>.github.io --set-secrets OPEN_SKILL_DB_TOKEN=...`
- [ ] Optional: `.github/workflows/deploy-service.yml` using Workload Identity Federation

**Test**
- `curl https://<service-url>/api/health` → `{"ok":true}`
- Preflight: `curl -i -X OPTIONS -H "Origin: https://<user>.github.io" <service-url>/api/communities` → correct `Access-Control-Allow-Origin`
- Request from a non-allowed origin → no ACAO header

---

## 6. Firebase Authentication (app + service) — required before public launch

Providers: Google (built in), GitHub (built in). Discord is not a built-in Firebase provider — see 6d.

### 6a. Firebase project setup
- [ ] Add Firebase to the existing GCP project (same project as Cloud Run)
- [ ] Authentication → enable Google and GitHub providers (GitHub needs an OAuth App client id/secret)
- [ ] Authorized domains: add `<user>.github.io` (and custom domain if used); `localhost` for dev
- [ ] Register a Web app; record `apiKey`, `authDomain`, `projectId`, `appId` (public config, not secrets)

**Test**
- Firebase console shows providers enabled and authorized domains listed

### 6b. App sign-in
- [ ] Add `firebase` npm package; init from `VITE_FIREBASE_*` env vars
- [ ] Replace demo login in `app/src/lib/auth.ts` with `signInWithPopup` (Google/GitHub) and `signOut`
- [ ] Drive auth state from `onAuthStateChanged` instead of `localStorage` flag; keep protected-route redirects
- [ ] In `req()`, send `Authorization: Bearer ${await auth.currentUser.getIdToken()}` (SDK auto-refreshes)
- [ ] Debug profiles / `X-Demo-User` only when `import.meta.env.DEV` (never in prod builds)
- [ ] Add `VITE_FIREBASE_*` to `app/.env.example` and the Pages deploy workflow

**Test**
- Sign in with Google and GitHub locally → protected routes accessible; sign out → redirected to `/login`
- Network tab shows `Authorization: Bearer ...` on API calls; prod build has no `X-Demo-User`

### 6c. Service token verification
- [ ] Add `firebase.google.com/go/v4`; verify ID tokens with `auth.Client.VerifyIDToken` (uses public certs, no secret needed; on Cloud Run uses default credentials, locally set `GOOGLE_CLOUD_PROJECT`)
- [ ] Middleware: missing/invalid token on protected endpoints → 401; resolve Firebase `uid` → user id
- [ ] Add `firebase_uid` mapping (unique) and create user/player on first sign-in, seeding handle/avatar from token claims
- [ ] Only honour `X-Demo-User` when `OPEN_SKILL_DEMO_AUTH=1` (dev/tests); off by default
- [ ] Remove `X-Demo-User` from CORS allowed headers unless demo auth is enabled
- [ ] Keep tests working via an injectable token verifier (fake in tests)

**Test**
- No token → 401 on protected endpoints; public endpoints (health, public leaderboard) still work
- Forged `X-Demo-User` with demo auth off → ignored / 401
- Expired or tampered token → 401
- First sign-in creates the user; second sign-in reuses it
- Two signed-in users see only their own memberships/permissions
- `go test ./...` and `npm test` pass

### 6d. Discord sign-in (optional, via custom token)
- [ ] Register a Discord application; store client secret in Secret Manager
- [ ] Service: `/api/auth/discord/login` + `/callback` (authorization code + PKCE + `state`), fetch Discord user
- [ ] Mint Firebase custom token (`auth.Client.CustomToken`, uid `discord:<id>`); requires service account with `iam.serviceAccounts.signBlob` on Cloud Run
- [ ] Redirect back to app with a one-time code; app exchanges it and calls `signInWithCustomToken`

**Test**
- Discord sign-in completes and app receives a Firebase session
- Invalid/missing `state` → rejected
- Discord user is a distinct account from Google/GitHub users (unless linked)

---

## 7. End-to-end verification

- [ ] Pages app → Cloud Run → Turso with real sign-in
- [ ] Create community, join from second account, report match, ratings update
- [ ] Restart/redeploy Cloud Run → data persists
- [ ] Cold start after idle is acceptable

---

## Optional / later

- [ ] Evaluate Cloudflare Pages instead of GitHub Pages (native SPA fallback, no base path issues, preview deploys)
- [ ] Custom domain for app and API
- [ ] Turso backups / point-in-time restore check
- [ ] Basic rate limiting on write endpoints
