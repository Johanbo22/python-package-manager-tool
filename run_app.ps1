Stop-Process -Name "main" -ErrorAction SilentlyContinue
Stop-Process -Name "python" -ErrorAction SilentlyContinue

$env:MANAGER_API_KEY = "my-secret-key-123"

Write-Host "Starting Python Backend" -ForegroundColor Green
Start-Process powershell -ArgumentList "-NoExit", "-Command", "& {cd backend_python; python main_service.py}"

Start-Sleep -Seconds 2

Write-Host "Starting Go" -ForegroundColor Cyan
cd backend_go
go run cmd/main.go