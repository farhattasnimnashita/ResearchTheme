param([string]$LoginPath = 'D:\project-250\ResearchTheme-GitHub\Login')
$ErrorActionPreference = 'Stop'
$mainPath = Join-Path $LoginPath 'main.go'
if (!(Test-Path $mainPath)) { throw "main.go not found. Pass your Login folder with -LoginPath." }
$source = [System.IO.File]::ReadAllText($mainPath)
$needle = 'page := "web/index.html"'
$alreadyInstalled = $source.Contains('page = "web/landing.html"')
if (!$alreadyInstalled -and ([regex]::Matches($source, [regex]::Escape($needle))).Count -ne 1) {
  throw 'The homepage route differs in this version. No files changed. Follow README_BN.txt or share your current main.go.'
}
$stamp = Get-Date -Format 'yyyyMMdd-HHmmss-fff'
$backup = Join-Path $LoginPath "backup-landing-$stamp"
New-Item -ItemType Directory -Path $backup | Out-Null
Copy-Item $mainPath (Join-Path $backup 'main.go')
$relativeFiles = @('web\landing.html', 'web\assets\landing.css', 'web\assets\landing-campus.png')
foreach ($relative in $relativeFiles) {
  $target = Join-Path $LoginPath $relative
  if (Test-Path $target) {
    $saved = Join-Path $backup $relative
    New-Item -ItemType Directory -Force -Path (Split-Path $saved) | Out-Null
    Copy-Item $target $saved
  }
  New-Item -ItemType Directory -Force -Path (Split-Path $target) | Out-Null
  Copy-Item (Join-Path (Join-Path $PSScriptRoot 'Login') $relative) $target
}
if (!$alreadyInstalled) {
  $newline = "`n"
  if ($source.Contains("`r`n")) { $newline = "`r`n" }
  $addition = $needle + $newline + "`t`t" + 'if p == "/" {' + $newline + "`t`t`t" + 'page = "web/landing.html"' + $newline + "`t`t" + '}'
  $source = $source.Replace($needle, $addition)
  [System.IO.File]::WriteAllText($mainPath, $source, (New-Object System.Text.UTF8Encoding($false)))
}
Write-Host 'Landing page installed. Restart go run . and open http://localhost:8080/'
Write-Host "Previous files saved in: $backup"
