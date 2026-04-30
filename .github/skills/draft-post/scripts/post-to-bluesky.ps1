<#
.SYNOPSIS
    Post to Bluesky using PowerShell (no Go binary needed).
    Replicates the Go-based post-to-bluesky/main.go logic.

.PARAMETER Mode
    "publish" or "draft"

.PARAMETER File
    Path to markdown file with YAML frontmatter

.EXAMPLE
    .\post-to-bluesky.ps1 -Mode publish -File ..\posts\post-12-config-resilience.md
#>

param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("publish","draft")]
    [string]$Mode,

    [Parameter(Mandatory=$true)]
    [string]$File
)

# Load .env from the Go project's directory (same credentials)
$envFile = Join-Path $PSScriptRoot "post-to-bluesky\.env"
if (Test-Path $envFile) {
    Get-Content $envFile | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
            [System.Environment]::SetEnvironmentVariable($Matches[1].Trim(), $Matches[2].Trim(), "Process")
        }
    }
}

$handle = $env:BSKY_HANDLE
$appPassword = $env:BSKY_APP_PASSWORD
$devtoURL = $env:BSKY_DEVTO_URL

if (-not $handle -or -not $appPassword) {
    Write-Error "BSKY_HANDLE and BSKY_APP_PASSWORD must be set in post-to-bluesky\.env"
    exit 1
}

# Resolve file path
if (-not [System.IO.Path]::IsPathRooted($File)) {
    $File = Join-Path (Get-Location) $File
}
if (-not (Test-Path $File)) {
    Write-Error "File not found: $File"
    exit 1
}

# Parse frontmatter
$content = Get-Content $File -Raw
$title = ""
$tags = @()
if ($content -match '(?s)^---\s*\n(.+?)\n---') {
    $frontmatter = $Matches[1]
    foreach ($line in $frontmatter -split "`n") {
        $line = $line.Trim()
        if ($line -match '^title:\s*(.+)$') { $title = $Matches[1].Trim() }
        if ($line -match '^tags:\s*(.+)$') { $tags = $Matches[1] -split ',' | ForEach-Object { $_.Trim() } }
    }
}

# Build post text
$postText = "New post: $title`n`nRead the full article below"

# Add hashtags (respect 300 grapheme limit)
$hashtagLine = ""
foreach ($t in $tags) {
    $candidate = $hashtagLine + "#$t "
    if (($postText + "`n`n" + $candidate.Trim()).Length -gt 280) { break }
    $hashtagLine = $candidate
}
if ($hashtagLine) { $postText += "`n`n" + $hashtagLine.Trim() }

# Add link label
$linkLabel = "Read here"
if ($devtoURL) {
    $postText += "`n`n$linkLabel"
}

if ($Mode -eq "draft") {
    Write-Host "=== DRY RUN (would post) ==="
    Write-Host "Text ($($postText.Length) chars):"
    Write-Host $postText
    Write-Host "Draft ID: dry-run"
    exit 0
}

# Authenticate
$bskyAPI = "https://bsky.social/xrpc"
$authBody = @{ identifier = $handle; password = $appPassword } | ConvertTo-Json
try {
    $session = Invoke-RestMethod -Uri "$bskyAPI/com.atproto.server.createSession" -Method Post -Body $authBody -ContentType "application/json"
} catch {
    Write-Error "Auth failed: $($_.Exception.Message)"
    exit 1
}

Write-Host "Authenticated as @$($session.handle)" -ForegroundColor Green

# Build record
$textBytes = [System.Text.Encoding]::UTF8.GetBytes($postText.Replace($linkLabel, ""))
$byteStart = $textBytes.Length
$linkLabelBytes = [System.Text.Encoding]::UTF8.GetBytes($linkLabel)
$byteEnd = $byteStart + $linkLabelBytes.Length

$record = @{
    '$type'   = "app.bsky.feed.post"
    text      = $postText
    createdAt = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
}

if ($devtoURL) {
    $record.embed = @{
        '$type'  = "app.bsky.embed.external"
        external = @{
            uri         = $devtoURL
            title       = $title
            description = "Read the full article on dev.to"
        }
    }
    $record.facets = @(
        @{
            index    = @{ byteStart = $byteStart; byteEnd = $byteEnd }
            features = @(
                @{ '$type' = "app.bsky.richtext.facet#link"; uri = $devtoURL }
            )
        }
    )
}

$payload = @{
    repo       = $session.did
    collection = "app.bsky.feed.post"
    record     = $record
} | ConvertTo-Json -Depth 10

$headers = @{ Authorization = "Bearer $($session.accessJwt)" }

try {
    $result = Invoke-RestMethod -Uri "$bskyAPI/com.atproto.repo.createRecord" -Method Post -Body $payload -ContentType "application/json" -Headers $headers
    $postID = ($result.uri -split "/")[-1]
    $postURL = "https://bsky.app/profile/$($session.handle)/post/$postID"
    Write-Host "URL: $postURL" -ForegroundColor Green
} catch {
    $errBody = $_.ErrorDetails.Message
    Write-Error "Post failed: $errBody"
    exit 1
}
