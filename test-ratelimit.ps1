

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Rate Limiting Test Script" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "⚠️  ÖNEMLİ: Load balancer'ın port 8080'de çalıştığından emin olun!" -ForegroundColor Yellow
Write-Host "   Komut: go run cmd/lb/main.go" -ForegroundColor Yellow
Write-Host ""

Write-Host "Sending 5 rapid requests to port 8080..." -ForegroundColor Green
Write-Host ""

for ($i = 1; $i -le 5; $i++) {
    Write-Host "Request $i..." -ForegroundColor White
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/" -Method GET -ErrorAction SilentlyContinue
        Write-Host "  Status: $($response.StatusCode) - OK" -ForegroundColor Green
    }
    catch {
        if ($_.Exception.Response.StatusCode -eq 429) {
            Write-Host "  Status: 429 - RATE LIMITED! ✅" -ForegroundColor Red
            $stream = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($stream)
            $body = $reader.ReadToEnd()
            Write-Host "  Response: $body" -ForegroundColor Yellow
        }
        else {
            Write-Host "  Error: $($_.Exception.Message)" -ForegroundColor Red
        }
    }
    Start-Sleep -Milliseconds 200
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Checking Dashboard Stats..." -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

try {
    $stats = Invoke-RestMethod -Uri "http://localhost:9091/stats" -Method GET
    Write-Host "Total Blocked Requests: $($stats.total_blocked)" -ForegroundColor Magenta
    Write-Host ""
    Write-Host "Backend Stats:" -ForegroundColor White
    foreach ($backend in $stats.backends) {
        $status = if ($backend.alive) { "✅ ALIVE" } else { "❌ DOWN" }
        Write-Host "  - $($backend.url): $status (Requests: $($backend.request_count))" -ForegroundColor White
    }
}
catch {
    Write-Host "Error fetching stats: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Test Complete!" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
