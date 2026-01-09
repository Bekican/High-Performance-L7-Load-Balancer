# Traffic Generator - Load Balancer'a istek gönderir
param(
    [int]$RequestCount = 100,
    [int]$DelayMs = 100
)

Write-Host "Traffic Generator Baslatildi" -ForegroundColor Cyan
Write-Host "   Hedef: http://localhost:8080" -ForegroundColor White
Write-Host "   Istek Sayisi: $RequestCount" -ForegroundColor White
Write-Host "   Gecikme: ${DelayMs}ms" -ForegroundColor White
Write-Host ""

$successCount = 0
$errorCount = 0

for ($i = 1; $i -le $RequestCount; $i++) {
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080" -TimeoutSec 5 -UseBasicParsing -ErrorAction Stop
        $successCount++
        
        if ($i % 10 -eq 0) {
            $percent = [math]::Round(($i / $RequestCount) * 100)
            Write-Host "Progress: $i/$RequestCount ($percent%) - Success: $successCount, Errors: $errorCount" -ForegroundColor Green
        }
    }
    catch {
        $errorCount++
        Write-Host "Request $i failed" -ForegroundColor Red
    }
    
    Start-Sleep -Milliseconds $DelayMs
}

Write-Host ""
Write-Host "Sonuclar:" -ForegroundColor Cyan
Write-Host "   Toplam Istek: $RequestCount" -ForegroundColor White
Write-Host "   Basarili: $successCount" -ForegroundColor Green
Write-Host "   Hata: $errorCount" -ForegroundColor Red
Write-Host ""
Write-Host "Dashboard'u kontrol edin: http://localhost:5173" -ForegroundColor Yellow
