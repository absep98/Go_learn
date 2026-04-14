<#
.SYNOPSIS
    Post a markdown article to all platforms (dev.to, Hashnode, Substack) in one command.

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
Write-Host "[1/3] dev.to..." -ForegroundColor Yellow -NoNewline
$devtoDir = Join-Path $scriptRoot "post-to-devto"
$devtoEnv = Join-Path $devtoDir ".env"

if (Test-Path $devtoEnv) {
    try {
        Push-Location $devtoDir
        $output = & go run main.go $Mode $File 2>&1 | Out-String
        Pop-Location

        if ($LASTEXITCODE -eq 0) {
            Write-Host " OK" -ForegroundColor Green
            $results += [PSCustomObject]@{ Platform = "dev.to"; Status = "Success"; Detail = ($output -split "`n" | Where-Object { $_ -match "URL|ID|Draft" } | Select-Object -First 1).Trim() }
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
Write-Host "[2/3] Hashnode..." -ForegroundColor Yellow -NoNewline
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
Write-Host "[3/3] Substack..." -ForegroundColor Yellow -NoNewline
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

# ── Summary ───────────────────────────────────────────────────
Write-Host ""
Write-Host "=== Results ===" -ForegroundColor Cyan
$results | Format-Table -AutoSize
