$ErrorActionPreference = "Stop"

$testDir = "tests/mtls_data"
if (Test-Path $testDir) { Remove-Item -Recurse -Force $testDir }
New-Item -ItemType Directory -Path $testDir

Write-Host "--- 1. Generating mTLS Certificates ---" -ForegroundColor Cyan
go run scripts/gen_certs/main.go $testDir

# Set environment variables for the session (inherited by child processes)
$env:GS_INGEST_CERT = "$testDir/server.crt"
$env:GS_INGEST_KEY = "$testDir/server.key"
$env:GS_INGEST_CA = "$testDir/ca.crt"
$env:GS_MONITOR_MAX_RAM = "104857600" # 100MB

Write-Host "--- 2. Starting GopherShip with mTLS ---" -ForegroundColor Cyan

# Ensure any previous instances are cleaned up
Get-Process go -ErrorAction SilentlyContinue | Where-Object { $_.CommandLine -like "*cmd/gophership*" } | Stop-Process -Force -ErrorAction SilentlyContinue

# Start GopherShip in background with unique logging
$ts = Get-Date -Format "yyyyMMdd_HHmmss"
$outFile = "gophership_out_$ts.log"
$errFile = "gophership_err_$ts.log"

$gsProcess = Start-Process -FilePath "go" -ArgumentList "run ./cmd/gophership" -NoNewWindow -PassThru -RedirectStandardOutput $outFile -RedirectStandardError $errFile

try {
    Write-Host "Waiting for server to start (10s)..."
    Start-Sleep -Seconds 10
    if (Test-Path $errFile) { Get-Content $errFile -Tail 10 }
    if (Test-Path $outFile) { Get-Content $outFile -Tail 10 }

    Write-Host "--- 3. Testing gs-ctl with Authorized Client ---" -ForegroundColor Green
    go run ./cmd/gs-ctl --tls --cert "$testDir/client.crt" --key "$testDir/client.key" --ca "$testDir/ca.crt" --addr "localhost:9092" status
    if ($LASTEXITCODE -ne 0) { throw "Authorized connection failed!" }

    Write-Host "--- 4. Testing gs-ctl without Credentials (Should Fail) ---" -ForegroundColor Yellow
    $ErrorActionPreference = "Continue"
    go run ./cmd/gs-ctl --tls --addr "localhost:9092" status
    if ($LASTEXITCODE -eq 0) { throw "Unauthorized connection (No Cert) incorrectly allowed!" }
    $ErrorActionPreference = "Stop"
    Write-Host "Rejected as expected."

    Write-Host "--- 5. Testing gs-ctl with Wrong CA (Should Fail) ---" -ForegroundColor Yellow
    # Create a rogue CA
    $rogueDir = "$testDir/rogue"
    New-Item -ItemType Directory -Path $rogueDir
    go run scripts/gen_certs/main.go $rogueDir

    $ErrorActionPreference = "Continue"
    go run ./cmd/gs-ctl --cert "$rogueDir/client.crt" --key "$rogueDir/client.key" --ca "$rogueDir/ca.crt" --addr "localhost:9092" status
    if ($LASTEXITCODE -eq 0) { throw "Unauthorized connection (Rogue CA) incorrectly allowed!" }
    $ErrorActionPreference = "Stop"
    Write-Host "Rejected as expected."

} finally {
    Write-Host "--- Cleanup ---" -ForegroundColor Cyan
    Stop-Process -Id $gsProcess.Id -Force -ErrorAction SilentlyContinue
    Remove-Item -Recurse -Force $testDir
}

Write-Host "mTLS Validation Passed!" -ForegroundColor Green
