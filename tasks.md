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

- [x] Add empty `app/static/.nojekyll` (Pages/Jekyll ignores `_app/`)
- [x] Change adapter `fallback` to `404.html` so deep links load
- [x] Hosting URL decided: custom domain cascades to project pages as a subpath → `https://avanderw.co.za/open-skill`. `VITE_BASE_PATH=/open-skill` sets `kit.paths.base` (`app/vite.config.ts`); all hard-coded `href="/..."`, `goto('/...')` and `redirect(307, '/...')` now use `` `${base}/...` `` (incl. `auth.ts` `isProtectedPath`/`loginPath`)

**Test**
- `npm run build` (no `VITE_BASE_PATH`) → `build/.nojekyll` and `build/404.html` exist; asset links are root-relative (`/_app/...`) — unchanged for Docker/nginx
- `VITE_BASE_PATH=/open-skill npm run build` → `build/404.html` asset links are `/open-skill/_app/...`
- `npm test` and `npm run check` pass

---

## 3. GitHub Actions deploy for app

- [x] Add `.github/workflows/deploy-app.yml` (build `app/`, `actions/upload-pages-artifact`, `actions/deploy-pages`)
- [ ] Add repo variables: `VITE_API_URL` (Cloud Run URL, after task 5), `VITE_BASE_PATH=/open-skill`, `VITE_FIREBASE_API_KEY`, `VITE_FIREBASE_AUTH_DOMAIN`, `VITE_FIREBASE_PROJECT_ID`, `VITE_FIREBASE_APP_ID`
- [ ] Enable Pages → Source: GitHub Actions in repo settings
- [ ] Do **not** add a `CNAME` file to this repo — the custom domain is configured on the `avanderw.github.io` user-site repo, which GitHub automatically cascades to project sites as a subpath
- [ ] Add `https://avanderw.co.za` to `OPEN_SKILL_ALLOWED_ORIGINS` (service) and Firebase authorized domains (task 6a)

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

- [x] Confirm `service/Dockerfile` builds a small static image and listens on `$PORT` (default 8080)
- [ ] GCP Console (no gcloud CLI needed): enable APIs, Artifact Registry repo, Workload Identity Federation, deploy service account, IAM bindings — see walkthrough below
- [ ] Billing on `open-skill-25b83` (Firebase Blaze) + **budget alert**
- [ ] Production Turso DB + token (separate from dev) → Secret Manager `turso_token`; grant the Cloud Run runtime service account `roles/secretmanager.secretAccessor`
- [ ] Add GitHub repo secrets (`GCP_WORKLOAD_IDENTITY_PROVIDER`, `GCP_DEPLOY_SERVICE_ACCOUNT`) and variables (`OPEN_SKILL_DB_URL`, `OPEN_SKILL_ALLOWED_ORIGINS`)
- [x] `.github/workflows/deploy-service.yml` — builds, pushes to Artifact Registry, deploys via Workload Identity Federation (keyless)

**Test**
- Push to `main` touching `service/**` → workflow green
- `curl https://<service-url>/api/health` → `{"ok":true}`
- `curl https://<service-url>/api/communities` → 401; with `-H "X-Demo-User: u_demo"` → still 401
- Preflight: `curl -i -X OPTIONS -H "Origin: https://<user>.github.io" <service-url>/api/communities` → correct `Access-Control-Allow-Origin`, no `X-Demo-User` in allowed headers
- Request from a non-allowed origin → no ACAO header

### GCP Console walkthrough (no gcloud CLI)
1. **Enable APIs** — console.cloud.google.com/apis/library, enable: Cloud Run Admin API, Artifact Registry API, Secret Manager API, IAM Service Account Credentials API.
2. **Artifact Registry repo** — console.cloud.google.com/artifacts → Create Repository → name `open-skill`, format Docker, region `europe-west1`.
3. **Deploy service account** — IAM & Admin → Service Accounts → Create: `github-deployer`. Grant roles: Artifact Registry Writer, Cloud Run Admin, Service Account User.
4. **Workload Identity Federation** — IAM & Admin → Workload Identity Federation → Create Pool (`github-pool`) → Add provider (OIDC, issuer `https://token.actions.githubusercontent.com`) → attribute mapping `google.subject = assertion.sub`, `attribute.repository = assertion.repository` → attribute condition `assertion.repository == "avanderw/open-skill"`.
5. **[x] Allow the pool to impersonate the service account** — granted `principalSet://iam.googleapis.com/projects/54185533602/locations/global/workloadIdentityPools/github-pool/attribute.repository/avanderw/open-skill` the Workload Identity User role on `github-deployer`.
6. **Secret** — Secret Manager → Create Secret `turso_token`, paste the production Turso token. Grant `secretAccessor` to both `github-deployer` and the default compute service account (Cloud Run runtime identity).
7. **GitHub repo secrets** (Settings → Secrets and variables → Actions → Secrets): `GCP_WORKLOAD_IDENTITY_PROVIDER` (full provider resource name from step 4), `GCP_DEPLOY_SERVICE_ACCOUNT` (`github-deployer@open-skill-25b83.iam.gserviceaccount.com`).
8. **GitHub repo variables** (same page → Variables): `OPEN_SKILL_DB_URL` (production Turso URL), `OPEN_SKILL_ALLOWED_ORIGINS` (start with `http://localhost:5173`, add the Pages origin after task 3).

---

## 6. Firebase Authentication (app + service) — required before public launch

Providers: Google (built in), GitHub (built in). Discord is not a built-in Firebase provider — see 6d.

### 6a. Firebase project setup
- [ ] Add Firebase to the existing GCP project (same project as Cloud Run)
- [ ] Authentication → enable Google and GitHub providers (GitHub needs an OAuth App client id/secret)
- [ ] Authorized domains: add `<user>.github.io` (and custom domain if used); `localhost` for dev
- [x] Register a Web app; record `apiKey`, `authDomain`, `projectId`, `appId` (public config, not secrets) — project `open-skill-25b83`

**Test**
- Firebase console shows providers enabled and authorized domains listed

### 6b. App sign-in
- [x] Add `firebase` npm package; init from `VITE_FIREBASE_*` env vars (local values in git-ignored `app/.env`)
- [x] Replace demo login in `app/src/lib/auth.ts` with `signInWithPopup` (Google/GitHub) and `signOut`
- [x] Drive auth state from Firebase (`authStateReady` + `onAuthStateChanged` → `auth-change` event); keep protected-route redirects
- [x] In `req()`, send `Authorization: Bearer <id token>` (SDK auto-refreshes)
- [x] Debug profiles / `X-Demo-User` only when `import.meta.env.DEV && VITE_DEMO_AUTH=1` (verified absent from prod build); tests use `app/.env.test`
- [x] Add `VITE_FIREBASE_*` to `app/.env.example`
- [ ] Add `VITE_FIREBASE_*` to the Pages deploy workflow (task 3, as repo variables)
- [ ] Optional: restrict the browser API key to HTTP referrers (`localhost`, Pages domain) in GCP Console → Credentials

**Test**
- Sign in with Google and GitHub locally → protected routes accessible; sign out → redirected to `/login`
- Network tab shows `Authorization: Bearer ...` on API calls; prod build has no `X-Demo-User`

### 6c. Service token verification
- [x] Add `firebase.google.com/go/v4`; verify ID tokens with `VerifyIDToken` (public certs, no credentials: `OPEN_SKILL_FIREBASE_PROJECT_ID` only)
- [x] Missing token on protected endpoints → 401; invalid token → 401 everywhere; anonymous allowed on leaderboard, modes, community lookup/membership, browse
- [x] User id `fb_<uid>` / player `p_<uid>` created on first sign-in; handle from display name (never email), random suffix if taken; cached so later requests skip the DB
- [x] Only honour `X-Demo-User` when `OPEN_SKILL_DEMO_AUTH=1` (dev/tests); service refuses to start with no auth configured
- [x] Remove `X-Demo-User` from CORS allowed headers unless demo auth is enabled
- [x] Injectable `TokenVerifier` (fake in `internal/api/auth_test.go`)
- [ ] Never set `FIREBASE_AUTH_EMULATOR_HOST` in production (the SDK then skips signature checks)

**Test**
- [x] No token → 401 on protected endpoints; public endpoints (health, public leaderboard) still work (unit tests + smoke test with real verifier)
- [x] Forged `X-Demo-User` with demo auth off → 401
- [x] Tampered/garbage token → 401
- [x] First sign-in creates the user; second sign-in reuses it
- [ ] Two real signed-in users see only their own memberships/permissions (manual, browser)
- [x] `go test ./...` and `npm test` pass

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
