# Demo Backend Sunucularını Başlat
Write-Host "🚀 Demo Backend Sunucuları Başlatılıyor..." -ForegroundColor Cyan

# Backend 1 (Port 8001)
Start-Process -FilePath "go" -ArgumentList "run", "demo/backend/main.go", "8001" -NoNewWindow
Write-Host "✅ Backend-8001 başlatıldı" -ForegroundColor Green

# Backend 2 (Port 8002)
Start-Process -FilePath "go" -ArgumentList "run", "demo/backend/main.go", "8002" -NoNewWindow
Write-Host "✅ Backend-8002 başlatıldı" -ForegroundColor Green

Start-Sleep -Seconds 2

Write-Host ""
Write-Host "📊 Sunucular hazır! Config.yaml'ı güncelleyin:" -ForegroundColor Yellow
Write-Host "   backends:" -ForegroundColor White
Write-Host "     - localhost:8001" -ForegroundColor White
Write-Host "     - localhost:8002" -ForegroundColor White
Write-Host ""
Write-Host "Veya Dashboard'dan 'Add Server' ile ekleyin!" -ForegroundColor Cyan
