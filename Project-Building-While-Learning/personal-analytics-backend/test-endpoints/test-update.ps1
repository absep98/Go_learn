# Test UPDATE Endpoint (PATCH /entries?id=X)
# Day 14 - Week 3
# Tests: Update own entry, non-existent entry, without token, invalid data

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  UPDATE ENDPOINT TESTS (PATCH /entries)" -ForegroundColor Cyan
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
            Uri = $Url
            Method = $Method
            ContentType = "application/json"
            Headers = $Headers
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
# SETUP: Create a test entry to update
# ============================================
Write-Host "SETUP: Creating test entry to update..." -ForegroundColor Magenta

try {
    $createResponse = Invoke-RestMethod -Uri "$baseUrl/entries" -Method POST -Headers $authHeaders -Body '{"text":"Original text for update test","mood":5,"category":"test"}' -ContentType "application/json"
    $testEntryId = $createResponse.id
    Write-Host "  Created entry with ID: $testEntryId" -ForegroundColor Green
    Write-Host ""
}
catch {
    # Entry might already exist, get the first one
    $entries = Invoke-RestMethod -Uri "$baseUrl/entries" -Method GET -Headers $authHeaders
    if ($entries.entries.Count -gt 0) {
        $testEntryId = $entries.entries[0].id
        Write-Host "  Using existing entry with ID: $testEntryId" -ForegroundColor Yellow
        Write-Host ""
    }
    else {
        Write-Host "  FAILED to create or find test entry!" -ForegroundColor Red
        exit 1
    }
}

# ============================================
# TEST 1: Update own entry (should succeed)
# ============================================
Write-Host ""
if (Test-Endpoint -Name "Update own entry with valid data" `
    -Method "PATCH" `
    -Url "$baseUrl/entries?id=$testEntryId" `
    -Headers $authHeaders `
    -Body '{"text":"Updated text!","mood":8,"category":"updated"}' `
    -ExpectedStatus 200 `
    -ExpectedMessage "Entry updated successfully") {
    $testsPassed++
} else {
    $testsFailed++
}

# ============================================
# TEST 2: Verify entry was actually updated
# ============================================
Write-Host ""
Write-Host "TEST: Verify entry was actually updated" -ForegroundColor Yellow
try {
    $entries = Invoke-RestMethod -Uri "$baseUrl/entries" -Method GET -Headers $authHeaders
    $updatedEntry = $entries.entries | Where-Object { $_.id -eq $testEntryId }
    
    if ($updatedEntry.text -eq "Updated text!" -and $updatedEntry.mood -eq 8 -and $updatedEntry.category -eq "updated") {
        Write-Host "  Entry values verified: text='$($updatedEntry.text)', mood=$($updatedEntry.mood), category='$($updatedEntry.category)'" -ForegroundColor Gray
        Write-Host "  PASSED" -ForegroundColor Green
        $testsPassed++
    }
    else {
        Write-Host "  FAILED: Entry values not updated correctly" -ForegroundColor Red
        Write-Host "  Got: text='$($updatedEntry.text)', mood=$($updatedEntry.mood), category='$($updatedEntry.category)'" -ForegroundColor Red
        $testsFailed++
    }
}
catch {
    Write-Host "  FAILED: Could not verify entry" -ForegroundColor Red
    $testsFailed++
}

# ============================================
# TEST 3: Update non-existent entry (404)
# ============================================
Write-Host ""
if (Test-Endpoint -Name "Update non-existent entry (ID 99999)" `
    -Method "PATCH" `
    -Url "$baseUrl/entries?id=99999" `
    -Headers $authHeaders `
    -Body '{"text":"Test","mood":5,"category":"test"}' `
    -ExpectedStatus 404) {
    $testsPassed++
} else {
    $testsFailed++
}

# ============================================
# TEST 4: Update without auth token (401)
# ============================================
Write-Host ""
if (Test-Endpoint -Name "Update without authentication" `
    -Method "PATCH" `
    -Url "$baseUrl/entries?id=$testEntryId" `
    -Body '{"text":"No auth","mood":5,"category":"test"}' `
    -ExpectedStatus 401) {
    $testsPassed++
} else {
    $testsFailed++
}

# ============================================
# TEST 5: Update with invalid mood (400)
# ============================================
Write-Host ""
if (Test-Endpoint -Name "Update with invalid mood (mood=15)" `
    -Method "PATCH" `
    -Url "$baseUrl/entries?id=$testEntryId" `
    -Headers $authHeaders `
    -Body '{"text":"Test","mood":15,"category":"test"}' `
    -ExpectedStatus 400) {
    $testsPassed++
} else {
    $testsFailed++
}

# ============================================
# TEST 6: Update with empty text (400)
# ============================================
Write-Host ""
if (Test-Endpoint -Name "Update with empty text" `
    -Method "PATCH" `
    -Url "$baseUrl/entries?id=$testEntryId" `
    -Headers $authHeaders `
    -Body '{"text":"","mood":5,"category":"test"}' `
    -ExpectedStatus 400) {
    $testsPassed++
} else {
    $testsFailed++
}

# ============================================
# TEST 7: Update with missing ID (400)
# ============================================
Write-Host ""
if (Test-Endpoint -Name "Update with missing entry ID" `
    -Method "PATCH" `
    -Url "$baseUrl/entries" `
    -Headers $authHeaders `
    -Body '{"text":"Test","mood":5,"category":"test"}' `
    -ExpectedStatus 400) {
    $testsPassed++
} else {
    $testsFailed++
}

# ============================================
# TEST 8: Update with invalid ID format (400)
# ============================================
Write-Host ""
if (Test-Endpoint -Name "Update with invalid ID format (id=abc)" `
    -Method "PATCH" `
    -Url "$baseUrl/entries?id=abc" `
    -Headers $authHeaders `
    -Body '{"text":"Test","mood":5,"category":"test"}' `
    -ExpectedStatus 400) {
    $testsPassed++
} else {
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
} else {
    Write-Host "  SOME TESTS FAILED!" -ForegroundColor Red
}
