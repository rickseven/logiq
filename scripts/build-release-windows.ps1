param(
  [string]$Version = ""
)

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
Push-Location $repoRoot

if ([string]::IsNullOrWhiteSpace($Version)) {
  $versionFile = Join-Path $repoRoot "internal\domain\version.go"
  $match = Select-String -Path $versionFile -Pattern 'var\s+Version\s*=\s*"([^"]+)"'
  if (-not $match) {
    throw "Failed to read version from internal/domain/version.go"
  }

  $rawVersion = $match.Matches[0].Groups[1].Value
  if ($rawVersion.StartsWith("v")) {
    $Version = $rawVersion
  } else {
    $Version = "v$rawVersion"
  }
}

New-Item -ItemType Directory -Force -Path .\build | Out-Null

$env:GOOS = "windows"
$env:GOARCH = "amd64"

go build -ldflags "-X github.com/rickseven/logiq/internal/domain.Version=$Version" `
  -o .\build\logiq-windows-amd64.exe .\cmd\logiq

.\build\logiq-windows-amd64.exe version

Pop-Location
