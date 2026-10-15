param(
  [Parameter(Mandatory = $true)]
  [string]$DistPath,
  [Parameter(Mandatory = $true)]
  [string]$Version,
  [Parameter(Mandatory = $true)]
  [string]$Repo,
  [string]$OutFile = ""
)

$ErrorActionPreference = "Stop"

if ($Version -like "v*") {
  $Version = $Version.Substring(1)
}

if (-not $OutFile) {
  $OutFile = Join-Path $DistPath "latest.json"
}

$assets = @()
Get-ChildItem -Path $DistPath -Recurse -File | Where-Object {
  $_.Name -match '^gogui-(windows|linux|darwin)-(amd64|arm64)(\.exe)?$'
} | ForEach-Object {
  $parts = $_.BaseName -split '-'
  # File name pattern: gogui-<os>-<arch>[.exe]
  $os = $parts[1]
  $arch = $parts[2]
  $sum = (Get-FileHash -Algorithm SHA256 -LiteralPath $_.FullName).Hash.ToLowerInvariant()
  $url = "https://github.com/$Repo/releases/download/v$Version/$($_.Name)"
  $assets += [ordered]@{
    os     = $os
    arch   = $arch
    name   = $_.Name
    url    = $url
    sha256 = $sum
  }
}

$feed = [ordered]@{
  version = $Version
  notes   = "Release $Version download $Version"
  url     = "https://github.com/$Repo/releases/tag/v$Version"
  assets  = $assets
}

$feed | ConvertTo-Json -Depth 5 | Set-Content -Path $OutFile -Encoding utf8
Write-Host "wrote $OutFile ($($assets.Count) assets)"
