# AI Guide and Repository Memory

## Local development on Windows

Use PowerShell from the repository root. Go and Node.js LTS must be installed.
After installing runtimes, refresh the terminal's PATH or restart the terminal.
If VS Code still has the old PATH, refresh it in each task:

```powershell
$env:Path = [Environment]::GetEnvironmentVariable('Path', 'Machine') + ';' + [Environment]::GetEnvironmentVariable('Path', 'User')
```

### Start the API

```powershell
.\service\with-env.ps1 go run .\cmd\open-skill-service -addr 127.0.0.1:8080
```

The helper loads the git-ignored `service\.env` and runs the command from
`service\`. Never print or commit environment secrets. Configuration may point
to a remote database: do not assume local development uses disposable data.
Do not seed data unless explicitly requested.

### Start the app

```powershell
npm.cmd --prefix app run dev -- --host 127.0.0.1 --port 5173 --strictPort
```

If frontend dependencies are missing, restore the lockfile dependencies:

```powershell
npm.cmd --prefix app ci
```

### Keep logs accessible

Run both processes in separate, persistent task terminals or background sessions
whose output remains accessible to the assistant. Capture stdout and stderr for
each process. If terminal output cannot be retrieved reliably, pipe it to
`Tee-Object -FilePath <log-path>` using separate log files outside version control.
Do not finish startup until both servers respond successfully. When troubleshooting,
inspect frontend/browser errors as well as API logs.

Check for existing listeners before starting another instance. Do not stop an
unrelated process just to free a port.

### Browser URL and authentication

**Use http://localhost:5173/login for testing Google sign-in, not
http://127.0.0.1:5173/login.**

Binding the server to `127.0.0.1` is fine; browse using `localhost`. Firebase treats
these as different hostnames. Testing with `127.0.0.1` produced the confirmed error
`auth/unauthorized-domain`. If `localhost` is also rejected, check Firebase Console
under Authentication > Settings > Authorized domains. Do not bypass authentication
or switch to demo auth to hide this configuration error.

Use port 5173 for normal development. A production preview on another port is
temporary and should be stopped after validation.

### Verify readiness

```powershell
Invoke-WebRequest -UseBasicParsing http://localhost:5173/login
Invoke-WebRequest -UseBasicParsing http://localhost:8080/api/health
Invoke-WebRequest -UseBasicParsing http://localhost:5173/api/health
```

Expect HTTP 200. Both health endpoints should return
`{"ok":true,"service":"open-skill"}`. The app forwards `/api` requests through
Vite to the API on port 8080.

## Focused frontend validation

```powershell
Set-Location app
npm.cmd run check
npm.cmd test -- src\lib\auth.test.ts src\lib\api.test.ts
npm.cmd run build
```

Choose tests relevant to the change rather than always running every command.
Checks and builds can regenerate SvelteKit files and reload the development
browser; do not run them concurrently with browser interaction tests.

Login behavior and browser performance measurements are documented in
[app/README.md](app/README.md). Local desktop timings are not evidence of
real-device mobile performance or authenticated token-refresh latency.

## Maintaining this guide

Keep entries actionable, verified, and repository-specific. Do not record
credentials, personal account details, transient process IDs, or session-specific
log paths. Update commands and constraints when the verified workflow changes.
