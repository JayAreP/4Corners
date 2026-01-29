# Build script for 4cornerscli
# Builds for both Windows and Linux

Write-Host "Building 4cornerscli for Windows and Linux..." -ForegroundColor Green
Write-Host ""

# Build for Windows
Write-Host "Building for Windows (amd64)..." -ForegroundColor Cyan
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o 4cornerscli.exe ./cmd/4cornerscli
if ($LASTEXITCODE -eq 0) {
    Write-Host "  ✓ Windows build successful: 4cornerscli.exe" -ForegroundColor Green
} else {
    Write-Host "  ✗ Windows build failed" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Build for Linux
Write-Host "Building for Linux (amd64)..." -ForegroundColor Cyan
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o 4cornerscli ./cmd/4cornerscli
if ($LASTEXITCODE -eq 0) {
    Write-Host "  ✓ Linux build successful: 4cornerscli" -ForegroundColor Green
} else {
    Write-Host "  ✗ Linux build failed" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Build complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Files created:" -ForegroundColor Yellow
Write-Host "  - 4cornerscli.exe (Windows)"
Write-Host "  - 4cornerscli (Linux)"
