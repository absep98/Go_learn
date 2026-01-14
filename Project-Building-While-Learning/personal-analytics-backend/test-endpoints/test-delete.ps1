# Test DELETE Endpoint (DELETE /entries?id=X)
# Day 15 - Week 3
# Tests: Delete own entry, verify deleted, non-existent entry, without token, invalid ID

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  DELETE ENDPOINT TESTS (DELETE /entries)" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:8080"
$testsPassed = 0
$testsFailed = 0

# Helper function to run tests
function Test-Endpoint {
    param(
        [string]$Name,
        [string]$Method,
        [string]$Url,
        [hashtable]$Headers = @{},
        [string]$Body = $null,
        [int]$ExpectedStatus,
        [string]$ExpectedMessage = $null
    )

    Write-Host "TEST: $Name" -ForegroundColor Yellow

    try {
        $params = @{
            Uri         = $Url
            Method      = $Method
            ContentType = "application/json"
            Headers     = $Headers
        }

        if ($Body) {
            $params.Body = $Body
        }

        $response = Invoke-RestMethod @params

        # If we expected an error status but got success, that's a failure
        if ($ExpectedStatus -ge 400) {
            Write-Host "  FAILED: Expected error $ExpectedStatus but got success" -ForegroundColor Red
            return $false
        }

        Write-Host "  Response: $($response | ConvertTo-Json -Compress)" -ForegroundColor Gray

        if ($ExpectedMessage -and $response.message -ne $ExpectedMessage) {
            Write-Host "  FAILED: Expected message '$ExpectedMessage' but got '$($response.message)'" -ForegroundColor Red
            return $false
        }

        Write-Host "  PASSED" -ForegroundColor Green
        return $true
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__

        if ($statusCode -eq $ExpectedStatus) {
            Write-Host "  Got expected error: $statusCode" -ForegroundColor Gray
            Write-Host "  PASSED" -ForegroundColor Green
            return $true
        }
        else {
            Write-Host "  FAILED: Expected $ExpectedStatus but got $statusCode" -ForegroundColor Red
            Write-Host "  Error: $($_.Exception.Message)" -ForegroundColor Red
            return $false
        }
    }
}

# ============================================
# SETUP: Login to get token
# ============================================
Write-Host "SETUP: Logging in to get token..." -ForegroundColor Magenta

try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/login" -Method POST -Body '{"email":"test@example.com","password":"password123"}' -ContentType "application/json"
    $token = $loginResponse.token
    Write-Host "  Token received successfully" -ForegroundColor Green
    Write-Host ""
}
catch {
    Write-Host "  FAILED to login! Make sure server is running and user exists." -ForegroundColor Red
    Write-Host "  Error: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Create auth header
$authHeaders = @{ Authorization = "Bearer $token" }

# ============================================
# SETUP: Create a test entry to delete
# ============================================
Write-Host "SETUP: Creating test entry to delete..." -ForegroundColor Magenta

try {
    $createResponse = Invoke-RestMethod -Uri "$baseUrl/entries" -Method POST -Headers $authHeaders -Body '{"text":"Entry to be deleted","mood":5,"category":"delete-test"}' -ContentType "application/json"
    $testEntryId = $createResponse.id
    Write-Host "  Created entry with ID: $testEntryId" -ForegroundColor Green
    Write-Host ""
}
catch {
    Write-Host "  FAILED to create test entry!" -ForegroundColor Red
    Write-Host "  Error: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# ============================================
# TEST 1: Delete own entry (should succeed)
# ============================================
Write-Host ""
if (Test-Endpoint -Name "Delete own entry" `
        -Method "DELETE" `
        -Url "$baseUrl/entries?id=$testEntryId" `
        -Headers $authHeaders `
        -ExpectedStatus 200 `
        -ExpectedMessage "Entry deleted successfully") {
    $testsPassed++
}
else {
    $testsFailed++
}

# ============================================
# TEST 2: Verify entry was actually deleted
# ============================================
Write-Host ""
Write-Host "TEST: Verify entry was actually deleted" -ForegroundColor Yellow
try {
    $entries = Invoke-RestMethod -Uri "$baseUrl/entries" -Method GET -Headers $authHeaders
    $deletedEntry = $entries.entries | Where-Object { $_.id -eq $testEntryId }

    if ($null -eq $deletedEntry) {
        Write-Host "  Entry ID $testEntryId no longer exists in database" -ForegroundColor Gray
        Write-Host "  PASSED" -ForegroundColor Green
        $testsPassed++
    }
    else {
        Write-Host "  FAILED: Entry still exists!" -ForegroundColor Red
        $testsFailed++
    }
}
catch {
    Write-Host "  FAILED: Could not verify deletion" -ForegroundColor Red
    $testsFailed++
}

# ============================================
# TEST 3: Delete already deleted entry (404)
# ============================================
Write-Host ""
if (Test-Endpoint -Name "Delete already deleted entry (should be 404)" `
        -Method "DELETE" `
        -Url "$baseUrl/entries?id=$testEntryId" `
        -Headers $authHeaders `
        -ExpectedStatus 404) {
    $testsPassed++
}
else {
    $testsFailed++
}

# ============================================
# TEST 4: Delete non-existent entry (404)
# ============================================
Write-Host ""
if (Test-Endpoint -Name "Delete non-existent entry (ID 99999)" `
        -Method "DELETE" `
        -Url "$baseUrl/entries?id=99999" `
        -Headers $authHeaders `
        -ExpectedStatus 404) {
    $testsPassed++
}
else {
    $testsFailed++
}

# ============================================
# TEST 5: Delete without auth token (401)
# ============================================
Write-Host ""
if (Test-Endpoint -Name "Delete without authentication" `
        -Method "DELETE" `
        -Url "$baseUrl/entries?id=1" `
        -ExpectedStatus 401) {
    $testsPassed++
}
else {
    $testsFailed++
}

# ============================================
# TEST 6: Delete with missing ID (400)
# ============================================
Write-Host ""
if (Test-Endpoint -Name "Delete with missing entry ID" `
        -Method "DELETE" `
        -Url "$baseUrl/entries" `
        -Headers $authHeaders `
        -ExpectedStatus 400) {
    $testsPassed++
}
else {
    $testsFailed++
}

# ============================================
# TEST 7: Delete with invalid ID format (400)
# ============================================
Write-Host ""
if (Test-Endpoint -Name "Delete with invalid ID format (id=abc)" `
        -Method "DELETE" `
        -Url "$baseUrl/entries?id=abc" `
        -Headers $authHeaders `
        -ExpectedStatus 400) {
    $testsPassed++
}
else {
    $testsFailed++
}

# ============================================
# SUMMARY
# ============================================
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  TEST SUMMARY" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Passed: $testsPassed" -ForegroundColor Green
Write-Host "  Failed: $testsFailed" -ForegroundColor $(if ($testsFailed -gt 0) { "Red" } else { "Green" })
Write-Host "  Total:  $($testsPassed + $testsFailed)" -ForegroundColor White
Write-Host ""

if ($testsFailed -eq 0) {
    Write-Host "  ALL TESTS PASSED!" -ForegroundColor Green
}
else {
    Write-Host "  SOME TESTS FAILED!" -ForegroundColor Red
}
