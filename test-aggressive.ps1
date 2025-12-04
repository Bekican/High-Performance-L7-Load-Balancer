

Write-Host "🔥 Aggressive Rate Limit Test" -ForegroundColor Cyan
Write-Host "Sending 10 rapid requests (no delay)..." -ForegroundColor Yellow
Write-Host ""

$successCount = 0
$blockedCount = 0

for ($i = 1; $i -le 10; $i++) {
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/" -Method GET -TimeoutSec 2 -ErrorAction Stop
        $successCount++
        Write-Host "✅ Request $i - Status: $($response.StatusCode)" -ForegroundColor Green
    }
    catch {
        if ($_.Exception.Response.StatusCode -eq 429) {
            $blockedCount++
            Write-Host "🚫 Request $i - RATE LIMITED (429)" -ForegroundColor Red
        }
        else {
            Write-Host "❌ Request $i - Error: $($_.Exception.Message)" -ForegroundColor Yellow
        }
    }
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Results:" -ForegroundColor White
Write-Host "  ✅ Successful: $successCount" -ForegroundColor Green
Write-Host "  🚫 Rate Limited: $blockedCount" -ForegroundColor Red
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "Dashboard Stats:" -ForegroundColor Cyan
try {
    $stats = Invoke-RestMethod -Uri "http://localhost:9091/stats"
    Write-Host "  Total Blocked: $($stats.total_blocked)" -ForegroundColor Magenta
}
catch {
    Write-Host "  Error fetching stats" -ForegroundColor Red
}
