param(
    [Parameter(Mandatory = $true)]
    [string]$Target
)
$ErrorActionPreference = 'Stop'
$source = $PSScriptRoot
if (-not (Test-Path -LiteralPath $Target -PathType Container)) { throw 'Target folder was not found.' }
$destination = (Resolve-Path -LiteralPath $Target).Path
if ($destination -eq $source) { throw 'Choose your existing project folder as Target, not this extracted update folder.' }
if (-not (Test-Path -LiteralPath (Join-Path $destination 'go.mod'))) { throw 'Target must be the existing Go project folder containing go.mod.' }
if (-not (Test-Path -LiteralPath (Join-Path $destination '.env'))) { throw 'Your existing .env is missing in Target. Copy your working Gmail-configured .env there first; do not recreate your email setup.' }
$backup = Join-Path $destination ('backup-update-' + (Get-Date -Format 'yyyyMMdd-HHmmss-fff'))
New-Item -ItemType Directory -Path $backup | Out-Null
# The user stops the server before running this script, so database files can be backed up consistently.
Get-ChildItem -LiteralPath $destination -File -Force | Where-Object { $_.Extension -eq '.go' -or $_.Name -in @('go.mod','go.sum','.env') -or $_.Name -like '*.db*' } | ForEach-Object { Copy-Item -LiteralPath $_.FullName -Destination $backup }
if (Test-Path -LiteralPath (Join-Path $destination 'web')) { Copy-Item -LiteralPath (Join-Path $destination 'web') -Destination $backup -Recurse }
Get-ChildItem -LiteralPath $source -File -Force | Where-Object { $_.Extension -in @('.go','.md') -or $_.Name -in @('go.mod','go.sum','START.ps1') } | ForEach-Object { Copy-Item -LiteralPath $_.FullName -Destination $destination -Force }
$web = Join-Path $destination 'web'
New-Item -ItemType Directory -Path $web -Force | Out-Null
Get-ChildItem -LiteralPath (Join-Path $source 'web') -Force | ForEach-Object { Copy-Item -LiteralPath $_.FullName -Destination $web -Recurse -Force }
$ignore = Join-Path $destination '.gitignore'
$existing = if (Test-Path -LiteralPath $ignore) { Get-Content -LiteralPath $ignore } else { @() }
Add-Content -LiteralPath $ignore -Value ''
foreach ($rule in @('.env','*.db','*.db-*','*.exe','backup*/')) { if ($existing -notcontains $rule) { Add-Content -LiteralPath $ignore -Value $rule } }
if (-not (Test-Path -LiteralPath (Join-Path $destination 'web\teacher-dashboard.html'))) { throw 'Teacher dashboard file was not copied.' }
Write-Host 'Update copied. Your .env and active database were preserved.'
Write-Host ('Backup: ' + $backup)
Write-Host ('Open a terminal in: ' + $destination)
Write-Host 'Run: go run .'
Write-Host 'Then refresh your browser with Ctrl+F5.'
