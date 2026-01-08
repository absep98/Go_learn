# Test User Login

# First, register a test user
Write-Host "`n=== Setup: Register Test User ===" -ForegroundColor Cyan
$registerBody = @{
    email = "login-test@example.com"
    password = "password123"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri http://localhost:8080/register -Method Post -Body $registerBody -ContentType "application/json"
    Write-Host "User registered: $($response.email)" -ForegroundColor Green
} catch {
    Write-Host "User might already exist (that's OK)" -ForegroundColor Yellow
}

# Test 1: Valid login
Write-Host "`n=== Test 1: Valid Login ===" -ForegroundColor Cyan
$loginBody = @{
    email = "login-test@example.com"
    password = "password123"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri http://localhost:8080/login -Method Post -Body $loginBody -ContentType "application/json"
    Write-Host "✅ Login successful!" -ForegroundColor Green
    Write-Host "Token received: $($response.token.Substring(0, 50))..." -ForegroundColor Cyan
    $global:token = $response.token
} catch {
    Write-Host "❌ Error: $_" -ForegroundColor Red
}

# Test 2: Wrong password
Write-Host "`n=== Test 2: Wrong Password (Should Fail) ===" -ForegroundColor Cyan
$wrongPasswordBody = @{
    email = "login-test@example.com"
    password = "wrongpassword"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri http://localhost:8080/login -Method Post -Body $wrongPasswordBody -ContentType "application/json"
    Write-Host "Success: $response" -ForegroundColor Green
} catch {
    Write-Host "✅ Expected error: Invalid credentials" -ForegroundColor Yellow
}

# Test 3: User doesn't exist
Write-Host "`n=== Test 3: User Doesn't Exist (Should Fail) ===" -ForegroundColor Cyan
$noUserBody = @{
    email = "nonexistent@example.com"
    password = "password123"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri http://localhost:8080/login -Method Post -Body $noUserBody -ContentType "application/json"
    Write-Host "Success: $response" -ForegroundColor Green
} catch {
    Write-Host "✅ Expected error: Invalid credentials" -ForegroundColor Yellow
}

# Test 4: Empty password
Write-Host "`n=== Test 4: Empty Password (Should Fail) ===" -ForegroundColor Cyan
$emptyPasswordBody = @{
    email = "login-test@example.com"
    password = ""
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri http://localhost:8080/login -Method Post -Body $emptyPasswordBody -ContentType "application/json"
    Write-Host "Success: $response" -ForegroundColor Green
} catch {
    Write-Host "✅ Expected error: Password required" -ForegroundColor Yellow
}

Write-Host "`n=== All Tests Complete ===" -ForegroundColor Cyan
if ($global:token) {
    Write-Host "`n💡 Token saved in `$token variable. You'll use this in Week 2 Day 10!" -ForegroundColor Green
}
