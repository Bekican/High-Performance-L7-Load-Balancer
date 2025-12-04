
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Traffic Distribution Test" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "Sending 20 requests to generate traffic distribution..." -ForegroundColor Green
Write-Host ""


for ($i = 1; $i -le 20; $i++) {
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/" -Method GET -ErrorAction SilentlyContinue
        Write-Host "Request $i - Status: $($response.StatusCode)" -ForegroundColor Green
    }
    catch {
        if ($_.Exception.Response.StatusCode -eq 429) {
            Write-Host "Request $i - Rate Limited (waiting 1 sec...)" -ForegroundColor Yellow
        }
        else {
            Write-Host "Request $i - Error: $($_.Exception.Message)" -ForegroundColor Red
        }
    }
    
    
    Start-Sleep -Seconds 1
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Checking Backend Distribution..." -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

try {
    $stats = Invoke-RestMethod -Uri "http://localhost:9091/stats"
    
    Write-Host "Backend Request Distribution:" -ForegroundColor White
    foreach ($backend in $stats.backends) {
        $status = if ($backend.alive) { "ONLINE" } else { "OFFLINE" }
        Write-Host "  $($backend.url): $($backend.request_count) requests [$status]" -ForegroundColor Cyan
    }
    
    Write-Host ""
    Write-Host "Total Blocked: $($stats.total_blocked)" -ForegroundColor Magenta
    Write-Host ""
    Write-Host "Dashboard'u yenileyin - grafik artik dolu olmali!" -ForegroundColor Yellow
}
catch {
    Write-Host "Error fetching stats: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Test Complete!" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
