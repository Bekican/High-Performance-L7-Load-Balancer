
Write-Host "Sending 10 requests (1 req/sec to avoid rate limiting)..." -ForegroundColor Cyan
Write-Host ""

for ($i = 1; $i -le 10; $i++) {
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/" -Method GET -ErrorAction SilentlyContinue
        Write-Host "Request $i - OK" -ForegroundColor Green
    }
    catch {
        Write-Host "Request $i - Error" -ForegroundColor Red
    }
    Start-Sleep -Seconds 1
}

Write-Host ""
Write-Host "Checking stats..." -ForegroundColor Cyan

try {
    $stats = Invoke-RestMethod -Uri "http://localhost:9091/stats"
    Write-Host ""
    Write-Host "Backend Distribution:" -ForegroundColor White
    foreach ($backend in $stats.backends) {
        Write-Host "  $($backend.url): $($backend.request_count) requests" -ForegroundColor Cyan
    }
    Write-Host ""
    Write-Host "Dashboard'u yenileyin!" -ForegroundColor Yellow
}
catch {
    Write-Host "Error fetching stats" -ForegroundColor Red
}
