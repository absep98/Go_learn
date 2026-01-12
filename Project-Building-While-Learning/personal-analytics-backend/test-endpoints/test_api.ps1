# Test the POST /entries endpoint

# Test 1: Valid entry
Write-Host "`n=== Test 1: Valid Entry ===" -ForegroundColor Green
$body = @{
    user_id = 101
    text = "Feeling productive today!"
    mood = 8
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "http://localhost:8080/entries" -Method Post -Body $body -ContentType "application/json"
Write-Host "Response:" -ForegroundColor Yellow
$response | ConvertTo-Json

# Test 2: Invalid - empty text
Write-Host "`n=== Test 2: Empty Text (Should Fail) ===" -ForegroundColor Green
$body = @{
    user_id = 101
    text = ""
    mood = 5
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/entries" -Method Post -Body $body -ContentType "application/json"
    Write-Host "Response:" -ForegroundColor Yellow
    $response | ConvertTo-Json
} catch {
    Write-Host "Error (Expected):" -ForegroundColor Red
    $_.Exception.Message
}

# Test 3: Invalid - mood out of range
Write-Host "`n=== Test 3: Mood Out of Range (Should Fail) ===" -ForegroundColor Green
$body = @{
    user_id = 101
    text = "Testing"
    mood = 15
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/entries" -Method Post -Body $body -ContentType "application/json"
    Write-Host "Response:" -ForegroundColor Yellow
    $response | ConvertTo-Json
} catch {
    Write-Host "Error (Expected):" -ForegroundColor Red
    $_.Exception.Message
}

# Test 4: Multiple valid entries
Write-Host "`n=== Test 4: Multiple Valid Entries ===" -ForegroundColor Green
$entries = @(
    @{ user_id = 102; text = "Started a new project"; mood = 7 },
    @{ user_id = 102; text = "Learned Go today"; mood = 9 },
    @{ user_id = 103; text = "Had a relaxing day"; mood = 6 }
)

foreach ($entry in $entries) {
    $body = $entry | ConvertTo-Json
    $response = Invoke-RestMethod -Uri "http://localhost:8080/entries" -Method Post -Body $body -ContentType "application/json"
    Write-Host "Created entry ID: $($response.id)" -ForegroundColor Cyan
}

Write-Host "`n✅ All tests complete!" -ForegroundColor Green
