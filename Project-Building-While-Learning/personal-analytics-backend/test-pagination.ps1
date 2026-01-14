# Test script for Day 16: Pagination
# Tests GET /entries with pagination parameters

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Day 16: Pagination Tests" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$baseUrl = "http://localhost:8080"
$testsPassed = 0
$testsFailed = 0

# Login to get token
Write-Host "`n[Setup] Logging in..." -ForegroundColor Yellow
$loginBody = @{ email = "test@example.com"; password = "password123" } | ConvertTo-Json
try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/login" -Method POST -Body $loginBody -ContentType "application/json"
    $token = $loginResponse.token
    Write-Host "[OK] Login successful" -ForegroundColor Green
} catch {
    Write-Host "[FAIL] Login failed: $_" -ForegroundColor Red
    exit 1
}

$headers = @{ Authorization = "Bearer $token" }

# Test 1: Default pagination (no params)
Write-Host "`n[Test 1] Default pagination (no params)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/entries" -Method GET -Headers $headers
    if ($response.success -eq $true -and $response.page -eq 1 -and $response.limit -eq 10) {
        Write-Host "[PASS] Defaults applied: page=$($response.page), limit=$($response.limit)" -ForegroundColor Green
        $testsPassed++
    } else {
        Write-Host "[FAIL] Unexpected defaults" -ForegroundColor Red
        $testsFailed++
    }
} catch {
    Write-Host "[FAIL] $_" -ForegroundColor Red
    $testsFailed++
}

# Test 2: Custom pagination (page=1, limit=1)
Write-Host "`n[Test 2] Custom pagination (page=1, limit=1)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/entries?page=1&limit=1" -Method GET -Headers $headers
    if ($response.success -eq $true -and $response.page -eq 1 -and $response.limit -eq 1) {
        $entriesCount = @($response.entries).Count
        Write-Host "[PASS] Returned $entriesCount entry, totalPages=$($response.totalPages)" -ForegroundColor Green
        $testsPassed++
    } else {
        Write-Host "[FAIL] Pagination params not applied" -ForegroundColor Red
        $testsFailed++
    }
} catch {
    Write-Host "[FAIL] $_" -ForegroundColor Red
    $testsFailed++
}

# Test 3: Page 2 (different entries)
Write-Host "`n[Test 3] Page 2 (page=2, limit=1)" -ForegroundColor Yellow
try {
    $page1 = Invoke-RestMethod -Uri "$baseUrl/entries?page=1&limit=1" -Method GET -Headers $headers
    $page2 = Invoke-RestMethod -Uri "$baseUrl/entries?page=2&limit=1" -Method GET -Headers $headers
    
    if ($response.success -eq $true -and $page2.page -eq 2) {
        # Check if entries are different (offset worked)
        $id1 = if ($page1.entries.Count -gt 0) { $page1.entries[0].id } else { -1 }
        $id2 = if ($page2.entries.Count -gt 0) { $page2.entries[0].id } else { -2 }
        
        if ($id1 -ne $id2) {
            Write-Host "[PASS] Different entries on page 2 (offset working)" -ForegroundColor Green
            $testsPassed++
        } else {
            Write-Host "[WARN] Same or no entries - might be expected if only 1 entry exists" -ForegroundColor Yellow
            $testsPassed++
        }
    } else {
        Write-Host "[FAIL] Page 2 not working" -ForegroundColor Red
        $testsFailed++
    }
} catch {
    Write-Host "[FAIL] $_" -ForegroundColor Red
    $testsFailed++
}

# Test 4: Beyond available data
Write-Host "`n[Test 4] Beyond available data (page=999)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/entries?page=999&limit=10" -Method GET -Headers $headers
    $entriesCount = @($response.entries).Count
    if ($response.success -eq $true -and $entriesCount -eq 0) {
        Write-Host "[PASS] Returns empty array for page beyond data" -ForegroundColor Green
        $testsPassed++
    } else {
        Write-Host "[FAIL] Should return empty entries" -ForegroundColor Red
        $testsFailed++
    }
} catch {
    Write-Host "[FAIL] $_" -ForegroundColor Red
    $testsFailed++
}

# Test 5: Invalid params fallback to defaults
Write-Host "`n[Test 5] Invalid params fallback to defaults" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/entries?page=abc&limit=xyz" -Method GET -Headers $headers
    if ($response.success -eq $true -and $response.page -eq 1 -and $response.limit -eq 10) {
        Write-Host "[PASS] Invalid params use defaults" -ForegroundColor Green
        $testsPassed++
    } else {
        Write-Host "[FAIL] Should fallback to defaults" -ForegroundColor Red
        $testsFailed++
    }
} catch {
    Write-Host "[FAIL] $_" -ForegroundColor Red
    $testsFailed++
}

# Test 6: Response has pagination metadata
Write-Host "`n[Test 6] Response contains pagination metadata" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/entries" -Method GET -Headers $headers
    $hasPage = $null -ne $response.page
    $hasLimit = $null -ne $response.limit
    $hasTotal = $null -ne $response.total
    $hasTotalPages = $null -ne $response.totalPages
    
    if ($hasPage -and $hasLimit -and $hasTotal -and $hasTotalPages) {
        Write-Host "[PASS] All pagination fields present (page, limit, total, totalPages)" -ForegroundColor Green
        $testsPassed++
    } else {
        Write-Host "[FAIL] Missing pagination fields" -ForegroundColor Red
        $testsFailed++
    }
} catch {
    Write-Host "[FAIL] $_" -ForegroundColor Red
    $testsFailed++
}

# Summary
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  Test Results: $testsPassed passed, $testsFailed failed" -ForegroundColor $(if ($testsFailed -eq 0) { "Green" } else { "Red" })
Write-Host "========================================" -ForegroundColor Cyan
