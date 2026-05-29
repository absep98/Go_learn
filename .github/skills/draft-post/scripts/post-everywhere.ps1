<#
.SYNOPSIS
    Post a markdown article to all platforms (dev.to, Hashnode, Substack, Reddit, Bluesky) in one command.

.PARAMETER File
    Path to markdown file with YAML frontmatter (title, tags, etc.)

.PARAMETER Mode
    "draft" or "publish" (default: draft)

.EXAMPLE
    .\post-everywhere.ps1 -File my-article.md -Mode draft
    .\post-everywhere.ps1 -File my-article.md -Mode publish
#>

param(
    [Parameter(Mandatory = $true)]
    [string]$File,

    [ValidateSet("draft", "publish")]
    [string]$Mode = "draft"
)

$ErrorActionPreference = "Continue"
$scriptRoot = Split-Path -Parent $MyInvocation.MyCommand.Path

# Resolve the markdown file to absolute path
if (-not [System.IO.Path]::IsPathRooted($File)) {
    $File = Join-Path (Get-Location) $File
}

if (-not (Test-Path $File)) {
    Write-Host "Error: File not found: $File" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "=== Post Everywhere ===" -ForegroundColor Cyan
Write-Host "File: $File"
Write-Host "Mode: $Mode"
Write-Host ""

$results = @()

# ── dev.to ────────────────────────────────────────────────────
Write-Host "[1/6] dev.to..." -ForegroundColor Yellow -NoNewline
$devtoDir = Join-Path $scriptRoot "post-to-devto"
$devtoEnv = Join-Path $devtoDir ".env"
$devtoArticleURL = ""

if (Test-Path $devtoEnv) {
    try {
        Push-Location $devtoDir
        $output = & go run main.go $Mode $File 2>&1 | Out-String
        Pop-Location

        if ($LASTEXITCODE -eq 0) {
            Write-Host " OK" -ForegroundColor Green
            $detail = ($output -split "`n" | Where-Object { $_ -match "URL|ID|Draft" } | Select-Object -First 1).Trim()
            $results += [PSCustomObject]@{ Platform = "dev.to"; Status = "Success"; Detail = $detail }
            # Extract the dev.to URL to pass to Bluesky as a link card
            $urlLine = ($output -split "`n" | Where-Object { $_ -match "^URL: https://dev\.to" } | Select-Object -First 1)
            if ($urlLine) { $devtoArticleURL = ($urlLine -replace "^URL: ", "").Trim() }
        }
        else {
            Write-Host " FAILED" -ForegroundColor Red
            $results += [PSCustomObject]@{ Platform = "dev.to"; Status = "Failed"; Detail = ($output -split "`n" | Select-Object -Last 2 | Select-Object -First 1).Trim() }
        }
    }
    catch {
        Pop-Location -ErrorAction SilentlyContinue
        Write-Host " ERROR" -ForegroundColor Red
        $results += [PSCustomObject]@{ Platform = "dev.to"; Status = "Error"; Detail = $_.Exception.Message }
    }
}
else {
    Write-Host " SKIPPED (no .env)" -ForegroundColor DarkGray
    $results += [PSCustomObject]@{ Platform = "dev.to"; Status = "Skipped"; Detail = "No .env configured" }
}

# ── Hashnode ──────────────────────────────────────────────────
Write-Host "[2/6] Hashnode..." -ForegroundColor Yellow -NoNewline
$hashnodeDir = Join-Path $scriptRoot "post-to-hashnode"
$hashnodeEnv = Join-Path $hashnodeDir ".env"

if (Test-Path $hashnodeEnv) {
    try {
        Push-Location $hashnodeDir
        $output = & go run main.go $Mode $File 2>&1 | Out-String
        Pop-Location

        if ($LASTEXITCODE -eq 0) {
            Write-Host " OK" -ForegroundColor Green
            $results += [PSCustomObject]@{ Platform = "Hashnode"; Status = "Success"; Detail = ($output -split "`n" | Where-Object { $_ -match "URL|ID|Draft" } | Select-Object -First 1).Trim() }
        }
        else {
            Write-Host " FAILED" -ForegroundColor Red
            $results += [PSCustomObject]@{ Platform = "Hashnode"; Status = "Failed"; Detail = ($output -split "`n" | Select-Object -Last 2 | Select-Object -First 1).Trim() }
        }
    }
    catch {
        Pop-Location -ErrorAction SilentlyContinue
        Write-Host " ERROR" -ForegroundColor Red
        $results += [PSCustomObject]@{ Platform = "Hashnode"; Status = "Error"; Detail = $_.Exception.Message }
    }
}
else {
    Write-Host " SKIPPED (no .env)" -ForegroundColor DarkGray
    $results += [PSCustomObject]@{ Platform = "Hashnode"; Status = "Skipped"; Detail = "No .env configured" }
}

# ── Substack ─────────────────────────────────────────────────
Write-Host "[3/6] Substack..." -ForegroundColor Yellow -NoNewline
$substackDir = Join-Path $scriptRoot "post-to-substack"
$substackEnv = Join-Path $substackDir ".env"

if (Test-Path $substackEnv) {
    try {
        Push-Location $substackDir
        $output = & node dist/main.js $Mode $File 2>&1 | Out-String
        Pop-Location

        if ($LASTEXITCODE -eq 0) {
            Write-Host " OK" -ForegroundColor Green
            $results += [PSCustomObject]@{ Platform = "Substack"; Status = "Success"; Detail = ($output -split "`n" | Where-Object { $_ -match "URL|ID|Edit" } | Select-Object -First 1).Trim() }
        }
        else {
            Write-Host " FAILED" -ForegroundColor Red
            $results += [PSCustomObject]@{ Platform = "Substack"; Status = "Failed"; Detail = ($output -split "`n" | Select-Object -Last 2 | Select-Object -First 1).Trim() }
        }
    }
    catch {
        Pop-Location -ErrorAction SilentlyContinue
        Write-Host " ERROR" -ForegroundColor Red
        $results += [PSCustomObject]@{ Platform = "Substack"; Status = "Error"; Detail = $_.Exception.Message }
    }
}
else {
    Write-Host " SKIPPED (no .env)" -ForegroundColor DarkGray
    $results += [PSCustomObject]@{ Platform = "Substack"; Status = "Skipped"; Detail = "No .env configured" }
}

# ── Reddit ────────────────────────────────────────────────────
Write-Host "[4/6] Reddit..." -ForegroundColor Yellow -NoNewline
$redditDir = Join-Path $scriptRoot "post-to-reddit"
$redditEnv = Join-Path $redditDir ".env"

if (Test-Path $redditEnv) {
    try {
        Push-Location $redditDir
        $output = & go run main.go $Mode $File 2>&1 | Out-String
        Pop-Location

        if ($LASTEXITCODE -eq 0) {
            Write-Host " OK" -ForegroundColor Green
            $results += [PSCustomObject]@{ Platform = "Reddit"; Status = "Success"; Detail = ($output -split "`n" | Where-Object { $_ -match "URL|ID|Draft|dry-run" } | Select-Object -First 1).Trim() }
        }
        else {
            Write-Host " FAILED" -ForegroundColor Red
            $results += [PSCustomObject]@{ Platform = "Reddit"; Status = "Failed"; Detail = ($output -split "`n" | Select-Object -Last 2 | Select-Object -First 1).Trim() }
        }
    }
    catch {
        Pop-Location -ErrorAction SilentlyContinue
        Write-Host " ERROR" -ForegroundColor Red
        $results += [PSCustomObject]@{ Platform = "Reddit"; Status = "Error"; Detail = $_.Exception.Message }
    }
}
else {
    Write-Host " SKIPPED (no .env)" -ForegroundColor DarkGray
    $results += [PSCustomObject]@{ Platform = "Reddit"; Status = "Skipped"; Detail = "No .env configured" }
}

# ── Bluesky ───────────────────────────────────────────────────
Write-Host "[5/6] Bluesky..." -ForegroundColor Yellow -NoNewline
$blueskyDir = Join-Path $scriptRoot "post-to-bluesky"
$blueskyEnv = Join-Path $blueskyDir ".env"

if (Test-Path $blueskyEnv) {
    try {
        # Auto-inject dev.to URL captured earlier as the link card URL
        if ($devtoArticleURL -ne "") { $env:BSKY_DEVTO_URL = $devtoArticleURL }

        # Try Go binary first, fall back to PowerShell script if blocked by Application Control
        $bskyPS = Join-Path $scriptRoot "post-to-bluesky.ps1"
        Push-Location $blueskyDir
        $output = & go run main.go $Mode $File 2>&1 | Out-String
        $goExitCode = $LASTEXITCODE
        Pop-Location

        if ($goExitCode -ne 0 -and $output -match "Application Control") {
            # Go binary blocked — use PowerShell fallback
            $output = & $bskyPS -Mode $Mode -File $File 2>&1 | Out-String
            $goExitCode = $LASTEXITCODE
        }

        $env:BSKY_DEVTO_URL = ""  # Reset after use

        if ($goExitCode -eq 0) {
            Write-Host " OK" -ForegroundColor Green
            $results += [PSCustomObject]@{ Platform = "Bluesky"; Status = "Success"; Detail = ($output -split "`n" | Where-Object { $_ -match "URL|ID|Draft|dry-run" } | Select-Object -First 1).Trim() }
        }
        else {
            Write-Host " FAILED" -ForegroundColor Red
            $results += [PSCustomObject]@{ Platform = "Bluesky"; Status = "Failed"; Detail = ($output -split "`n" | Select-Object -Last 2 | Select-Object -First 1).Trim() }
        }
    }
    catch {
        Pop-Location -ErrorAction SilentlyContinue
        Write-Host " ERROR" -ForegroundColor Red
        $results += [PSCustomObject]@{ Platform = "Bluesky"; Status = "Error"; Detail = $_.Exception.Message }
    }
}
else {
    Write-Host " SKIPPED (no .env)" -ForegroundColor DarkGray
    $results += [PSCustomObject]@{ Platform = "Bluesky"; Status = "Skipped"; Detail = "No .env configured" }
}
# ── LinkedIn ──────────────────────────────────────────────────
Write-Host "[6/6] LinkedIn..." -ForegroundColor Yellow -NoNewline
$linkedinDir = Join-Path $scriptRoot "post-to-linkedin"
$linkedinEnv = Join-Path $linkedinDir ".env"

if (Test-Path $linkedinEnv) {
    # Check if access token is configured
    $hasToken = (Get-Content $linkedinEnv | Where-Object { $_ -match '^LINKEDIN_ACCESS_TOKEN=.+' -and $_ -notmatch 'your-' })
    if ($hasToken) {
        try {
            # Auto-inject dev.to URL for link card
            if ($devtoArticleURL -ne "") { $env:LINKEDIN_DEVTO_URL = $devtoArticleURL }
            Push-Location $linkedinDir
            $output = & go run main.go $Mode $File 2>&1 | Out-String
            Pop-Location
            $env:LINKEDIN_DEVTO_URL = ""  # Reset after use

            if ($LASTEXITCODE -eq 0) {
                Write-Host " OK" -ForegroundColor Green
                $results += [PSCustomObject]@{ Platform = "LinkedIn"; Status = "Success"; Detail = ($output -split "`n" | Where-Object { $_ -match "URL|ID|Draft|Posted" } | Select-Object -First 1).Trim() }
            }
            else {
                Write-Host " FAILED" -ForegroundColor Red
                $results += [PSCustomObject]@{ Platform = "LinkedIn"; Status = "Failed"; Detail = ($output -split "`n" | Select-Object -Last 2 | Select-Object -First 1).Trim() }
            }
        }
        catch {
            Pop-Location -ErrorAction SilentlyContinue
            Write-Host " ERROR" -ForegroundColor Red
            $results += [PSCustomObject]@{ Platform = "LinkedIn"; Status = "Error"; Detail = $_.Exception.Message }
        }
    }
    else {
        Write-Host " SKIPPED (run 'go run main.go login' first)" -ForegroundColor DarkGray
        $results += [PSCustomObject]@{ Platform = "LinkedIn"; Status = "Skipped"; Detail = "No access token. Run login first." }
    }
}
else {
    Write-Host " SKIPPED (no .env)" -ForegroundColor DarkGray
    $results += [PSCustomObject]@{ Platform = "LinkedIn"; Status = "Skipped"; Detail = "No .env configured" }
}
# ── Summary ───────────────────────────────────────────────────
Write-Host ""
Write-Host "=== Results ===" -ForegroundColor Cyan
$results | Format-Table -AutoSize

# ── AksExplorer (manual) ──────────────────────────────────────
# Parse title from frontmatter for the manual post text
$fileContent = Get-Content $File -Raw
$postTitle = ""
if ($fileContent -match '(?m)^title:\s*(.+)$') { $postTitle = $Matches[1].Trim() }
$articleURL = if ($devtoArticleURL -ne "") { $devtoArticleURL } else { "https://dev.to" }

$aksExplorerText = @"
New post: $postTitle

$(if ($postTitle -match "Part (\d+)|#(\d+)") { "Part of the 'Learning Go in Public' series." } else { "From the 'Learning Go in Public' series." })

Full article: $articleURL

#golang #buildinpublic #backend #learning
"@

Write-Host ""
Write-Host "=== AksExplorer (manual paste) ===" -ForegroundColor Magenta
Write-Host "Go to: https://www.linkedin.com/company/aksexplorer/posts/" -ForegroundColor DarkGray
Write-Host "Click 'Start a post' and paste:" -ForegroundColor DarkGray
Write-Host ""
Write-Host $aksExplorerText -ForegroundColor White
Write-Host ""
Write-Host "--- (copied to clipboard) ---" -ForegroundColor DarkGray
$aksExplorerText | Set-Clipboard

# ── Substack (monetization for Indian creators) ───────────────
Write-Host ""
Write-Host "=== Substack Paid Setup (earns money - works in India!) ===" -ForegroundColor Green
Write-Host "Medium's Partner Program doesn't pay Indian creators. Use Substack instead:" -ForegroundColor DarkGray
Write-Host ""
Write-Host "ONE-TIME SETUP (if not done yet):" -ForegroundColor Yellow
Write-Host "1. Go to: https://substack.com/publish/payments" -ForegroundColor DarkGray
Write-Host "2. Enable paid subscriptions (Stripe India supported)" -ForegroundColor DarkGray
Write-Host "3. Set a monthly price (e.g. $5/month)" -ForegroundColor DarkGray
Write-Host ""
Write-Host "FOR THIS POST — share the Substack link on X/Twitter:" -ForegroundColor Yellow

# Extract substack URL from results if available
$substackResult = $results | Where-Object { $_.Platform -eq "Substack" -and $_.Status -eq "Success" }
$substackURL = ""
if ($substackResult -and $substackResult.Detail -match "(https://\S+substack\S+)") {
    $substackURL = $Matches[1]
}

if ($substackURL) {
    Write-Host "   $substackURL" -ForegroundColor White
    Write-Host "(Use this URL in your X/Twitter thread instead of the dev.to link)" -ForegroundColor DarkGray
} else {
    Write-Host "   https://abhisheksharma539226.substack.com" -ForegroundColor White
    Write-Host "(Check Substack dashboard for the direct post URL)" -ForegroundColor DarkGray
}
