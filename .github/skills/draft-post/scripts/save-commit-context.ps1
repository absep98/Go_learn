# save-commit-context.ps1 — called by post-commit hook or manually
# Saves latest commit info for the /draft-post Copilot skill

$repoRoot = git rev-parse --show-toplevel 2>$null
if (-not $repoRoot) { $repoRoot = "c:\learning\Go_learn" }
$repoRoot = $repoRoot -replace '/', '\'

$contextFile = Join-Path $repoRoot ".github\skills\draft-post\scripts\.last-commit-context.json"

$hash = git rev-parse --short HEAD
$fullHash = git rev-parse HEAD
$message = git log -1 --pretty=format:"%s"
$body = git log -1 --pretty=format:"%b"
$author = git log -1 --pretty=format:"%an"
$date = git log -1 --pretty=format:"%ai"
$filesChanged = (git diff-tree --no-commit-id --name-only -r HEAD) -join ","
$stat = (git diff-tree --no-commit-id --stat -r HEAD | Select-Object -Last 1).Trim()

$context = @{
    hash          = $hash
    full_hash     = $fullHash
    message       = $message
    body          = $body
    author        = $author
    date          = $date
    files_changed = $filesChanged
    stat          = $stat
    timestamp     = (Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ")
} | ConvertTo-Json -Depth 2

$context | Set-Content -Path $contextFile -Encoding UTF8

Write-Host ""
Write-Host "[draft-post] Commit context saved." -ForegroundColor Cyan
Write-Host "  Open Copilot Chat and type: /draft-post" -ForegroundColor Yellow
Write-Host ""
