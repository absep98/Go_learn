# Test GET /entries endpoint

Write-Host "`n=== Testing GET /entries ===" -ForegroundColor Green

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/entries" -Method Get
    
    Write-Host "`nSuccess!" -ForegroundColor Green
    Write-Host "Total entries: $($response.count)" -ForegroundColor Cyan
    
    if ($response.count -eq 0) {
        Write-Host "`nNo entries found (database is empty)" -ForegroundColor Yellow
    } else {
        Write-Host "`nEntries:" -ForegroundColor Yellow
        foreach ($entry in $response.entries) {
            Write-Host "---" -ForegroundColor Gray
            Write-Host "  ID: $($entry.id)"
            Write-Host "  User: $($entry.user_id)"
            Write-Host "  Text: $($entry.text)"
            Write-Host "  Mood: $($entry.mood)"
            Write-Host "  Created: $($entry.created_at)"
        }
    }
    
    Write-Host "`n✅ GET request successful!" -ForegroundColor Green
    
} catch {
    Write-Host "`n❌ Error:" -ForegroundColor Red
    Write-Host $_.Exception.Message
}
