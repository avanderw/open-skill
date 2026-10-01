# Runs a command with variables from service/.env (git-ignored), e.g.
#   ./with-env.ps1 go run ./cmd/open-skill-service -seed
#   ./with-env.ps1 go test ./internal/store -run Remote -v
$envFile = Join-Path $PSScriptRoot '.env'
if (-not (Test-Path $envFile)) { throw "Missing $envFile - copy .env.example to .env and fill it in." }
if ($args.Count -eq 0) { throw 'Usage: ./with-env.ps1 <command> [args...]' }

$names = @()
foreach ($line in Get-Content $envFile) {
	if ($line -match '^\s*(#|$)') { continue }
	$name, $value = $line -split '=', 2
	$name = $name.Trim()
	Set-Item "env:$name" $value.Trim()
	$names += $name
}

Push-Location $PSScriptRoot
try {
	& $args[0] @($args | Select-Object -Skip 1)
} finally {
	Pop-Location
	# Don't leave the token in the calling shell.
	$names | ForEach-Object { Remove-Item "env:$_" -ErrorAction SilentlyContinue }
}
