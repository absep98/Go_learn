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


# Test Authentication Middleware

Write-Host "`n=== Test 1: Access /entries WITHOUT token (Should Fail) ===" -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri http://localhost:8080/entries -Method Get
    Write-Host "❌ UNEXPECTED: Request succeeded without token!" -ForegroundColor Red
    Write-Host $response
}
catch {
    Write-Host "✅ EXPECTED: Request blocked without token" -ForegroundColor Green
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Yellow
}

Write-Host "`n=== Test 2: Login to get valid token ===" -ForegroundColor Cyan
$loginBody = @{
    email    = "login-test@example.com"
    password = "password123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri http://localhost:8080/login -Method Post -Body $loginBody -ContentType "application/json"
    Write-Host "✅ Login successful!" -ForegroundColor Green
    $token = $loginResponse.token
    Write-Host "Token: $($token.Substring(0, 50))..." -ForegroundColor Cyan
}
catch {
    Write-Host "❌ Login failed. Make sure user exists (run test-register.ps1 first)" -ForegroundColor Red
    Write-Host "Error: $_" -ForegroundColor Yellow
    exit 1
}

Write-Host "`n=== Test 3: Access /entries WITH valid token (Should Work) ===" -ForegroundColor Cyan
try {
    $headers = @{
        "Authorization" = "Bearer $token"
    }
    $response = Invoke-RestMethod -Uri http://localhost:8080/entries -Method Get -Headers $headers
    Write-Host "✅ Request successful with token!" -ForegroundColor Green
    Write-Host "Response: $($response | ConvertTo-Json -Depth 2)" -ForegroundColor Cyan
}
catch {
    Write-Host "❌ Request failed even with valid token" -ForegroundColor Red
    Write-Host "Error: $_" -ForegroundColor Yellow
}

Write-Host "`n=== Test 4: Create Entry WITH valid token and data (Should Work) ===" -ForegroundColor Cyan
try {
    $headers = @{
        "Authorization" = "Bearer $token"
    }
    $entryBody = @{
        text = "Test entry from comprehensive test script"
        mood = 8
        category = "testing"
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri http://localhost:8080/entries -Method Post -Headers $headers -Body $entryBody -ContentType "application/json"
    Write-Host "✅ Entry created successfully!" -ForegroundColor Green
    Write-Host "Response: $($response | ConvertTo-Json)" -ForegroundColor Cyan
}
catch {
    Write-Host "❌ Failed to create entry" -ForegroundColor Red
    Write-Host "Error: $_" -ForegroundColor Yellow
}

Write-Host "`n=== Test 5: Create Entry WITH invalid mood (Should Fail) ===" -ForegroundColor Cyan
try {
    $headers = @{
        "Authorization" = "Bearer $token"
    }
    $invalidMoodBody = @{
        text = "This should fail"
        mood = 11
        category = "testing"
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri http://localhost:8080/entries -Method Post -Headers $headers -Body $invalidMoodBody -ContentType "application/json"
    Write-Host "❌ UNEXPECTED: Entry created with invalid mood!" -ForegroundColor Red
}
catch {
    Write-Host "✅ EXPECTED: Entry rejected with invalid mood" -ForegroundColor Green
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Yellow
}

Write-Host "`n=== Test 6: Create Entry WITH empty text (Should Fail) ===" -ForegroundColor Cyan
try {
    $headers = @{
        "Authorization" = "Bearer $token"
    }
    $emptyTextBody = @{
        text = ""
        mood = 5
        category = "testing"
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri http://localhost:8080/entries -Method Post -Headers $headers -Body $emptyTextBody -ContentType "application/json"
    Write-Host "❌ UNEXPECTED: Entry created with empty text!" -ForegroundColor Red
}
catch {
    Write-Host "✅ EXPECTED: Entry rejected with empty text" -ForegroundColor Green
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Yellow
}

Write-Host "`n=== Test 7: Verify GET /entries returns created entries ===" -ForegroundColor Cyan
try {
    $headers = @{
        "Authorization" = "Bearer $token"
    }
    $response = Invoke-RestMethod -Uri http://localhost:8080/entries -Method Get -Headers $headers
    Write-Host "✅ Retrieved entries successfully!" -ForegroundColor Green
    Write-Host "Number of entries: $($response.Count)" -ForegroundColor Cyan
    if ($response.Count -gt 0) {
        Write-Host "Sample entry: $($response[0] | ConvertTo-Json)" -ForegroundColor Cyan
    }
}
catch {
    Write-Host "❌ Failed to retrieve entries" -ForegroundColor Red
    Write-Host "Error: $_" -ForegroundColor Yellow
}

Write-Host "`n=== Test 8: Access /entries WITH invalid token (Should Fail) ===" -ForegroundColor Cyan
try {
    $headers = @{
        "Authorization" = "Bearer invalid.token.here"
    }
    $response = Invoke-RestMethod -Uri http://localhost:8080/entries -Method Get -Headers $headers
    Write-Host "❌ UNEXPECTED: Request succeeded with invalid token!" -ForegroundColor Red
}
catch {
    Write-Host "✅ EXPECTED: Request blocked with invalid token" -ForegroundColor Green
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Yellow
}

Write-Host "`n=== Test 9: Access /entries WITH wrong format (Should Fail) ===" -ForegroundColor Cyan
try {
    $headers = @{
        "Authorization" = "NotBearer $token"
    }
    $response = Invoke-RestMethod -Uri http://localhost:8080/entries -Method Get -Headers $headers
    Write-Host "❌ UNEXPECTED: Request succeeded with wrong format!" -ForegroundColor Red
}
catch {
    Write-Host "✅ EXPECTED: Request blocked with wrong Authorization format" -ForegroundColor Green
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Yellow
}

Write-Host "`n=== All Middleware Tests Complete ===" -ForegroundColor Cyan
Write-Host "🎉 If all tests showed expected results, your middleware is working!" -ForegroundColor Green
