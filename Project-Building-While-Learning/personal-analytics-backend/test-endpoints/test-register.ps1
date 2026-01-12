# Test User Registration

# Test 1: Valid registration
Write-Host "`n=== Test 1: Valid Registration ===" -ForegroundColor Cyan
$body = @{
    email = "test@example.com"
    password = "password123"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri http://localhost:8080/register -Method Post -Body $body -ContentType "application/json"
    Write-Host "Success!" -ForegroundColor Green
    $response | ConvertTo-Json
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

# Test 2: Duplicate email (should fail)
Write-Host "`n=== Test 2: Duplicate Email (Should Fail) ===" -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri http://localhost:8080/register -Method Post -Body $body -ContentType "application/json"
    Write-Host "Success!" -ForegroundColor Green
    $response | ConvertTo-Json
} catch {
    Write-Host "Expected error: $_" -ForegroundColor Yellow
}

# Test 3: Short password (should fail)
Write-Host "`n=== Test 3: Short Password (Should Fail) ===" -ForegroundColor Cyan
$body = @{
    email = "test2@example.com"
    password = "123"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri http://localhost:8080/register -Method Post -Body $body -ContentType "application/json"
    Write-Host "Success!" -ForegroundColor Green
    $response | ConvertTo-Json
} catch {
    Write-Host "Expected error: $_" -ForegroundColor Yellow
}

# Test 4: Invalid email (should fail)
Write-Host "`n=== Test 4: Invalid Email (Should Fail) ===" -ForegroundColor Cyan
$body = @{
    email = "notanemail"
    password = "password123"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri http://localhost:8080/register -Method Post -Body $body -ContentType "application/json"
    Write-Host "Success!" -ForegroundColor Green
    $response | ConvertTo-Json
} catch {
    Write-Host "Expected error: $_" -ForegroundColor Yellow
}

Write-Host "`n=== All Tests Complete ===" -ForegroundColor Cyan
