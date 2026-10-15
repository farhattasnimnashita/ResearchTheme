$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw 'Install Go, restart VS Code, and run this file again.'
}
if (-not (Test-Path '.env')) { Copy-Item '.env.example' '.env' }
go mod download
if ($LASTEXITCODE -ne 0) { throw 'Could not download Go dependencies.' }
go run .
