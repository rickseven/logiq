param(
    [string]$Version = "v1.0.1"
)

New-Item -ItemType Directory -Force -Path .\build | Out-Null

$env:GOOS = "windows"
$env:GOARCH = "amd64"

go build -ldflags "-X github.com/rickseven/logiq/internal/domain.Version=$Version" `
  -o .\build\logiq-windows-amd64.exe .\cmd\logiq

.\build\logiq-windows-amd64.exe version
