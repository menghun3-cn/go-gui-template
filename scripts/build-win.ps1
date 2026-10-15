$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$version = git -C $root describe --tags --always --dirty
if (-not $version) { $version = "dev" }
$dist = Join-Path $root "dist"
New-Item -ItemType Directory -Force -Path $dist | Out-Null
$exe = Join-Path $dist "gogui-windows-amd64.exe"
& go build -tags production -trimpath -ldflags "-s -w -X main.version=$version -H windowsgui" -o $exe (Join-Path $root "cmd\gogui")
Write-Host "built $exe"
