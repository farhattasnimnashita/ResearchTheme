$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw 'Install Go, restart VS Code, and run this file again.'
}
if (-not (Test-Path '.env')) {
    throw 'No .env in this project folder. Copy your existing Gmail-configured .env here. For first-time setup, follow QUICK_START_BN.md. This script will not create console email settings.'
}
go mod download
if ($LASTEXITCODE -ne 0) { throw 'Could not download Go dependencies.' }
go run .
