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

Write-Host "`n=== Test 4: Access /entries WITH invalid token (Should Fail) ===" -ForegroundColor Cyan
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

Write-Host "`n=== Test 5: Access /entries WITH wrong format (Should Fail) ===" -ForegroundColor Cyan
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
