# LogIQ Release Guide

This document outlines how to build and distribute pre-compiled binaries of LogIQ so that other developers or AI agents can use the tool without needing to install Go.

There are two primary ways to create a release:

## 1. Automated Releases via GitHub Actions (Recommended)

LogIQ is configured with a GitHub Actions workflow (`.github/workflows/release.yml`) that automatically builds and attaches cross-platform binaries whenever a new Release is created on GitHub.

**Steps to create an automated release:**

1. Ensure all your latest changes are pushed to the `main` branch.
2. Go to your repository on GitHub: `https://github.com/rickseven/logiq`
3. Click on the **Releases** section on the right sidebar.
4. Click the **Draft a new release** button.
5. In the **Tag** field, enter a semantic version number (e.g., `v1.0.0` or `v1.1.2`). _Important: The tag must start with a `v` for the automation to recognize it properly depending on your GitHub settings._
6. Select the target branch (usually `main`).
7. Enter a Release title (e.g., "LogIQ v1.0.0 - Clean Architecture Redux").
8. Describe the changes, new features, or bug fixes in the description box.
9. Click **Publish release**.

**Result:** Within a few minutes, the GitHub Action will automatically compile LogIQ for Windows (amd64), macOS (amd64 & arm64 for M1/M2), and Linux (amd64). These `.exe` and binary files will be attached directly to the release page for anyone to download.

---

## 2. Manual Cross-Compilation (Local Builds)

If you need to quickly build a binary for a specific operating system manually (e.g., to share with a coworker via Slack before a formal release), you can use Go's built-in cross-compilation capabilities.

You execute these commands in your local terminal at the root of the `logiq` project.

**Build for Windows (Produces an `.exe` file):**

```powershell
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o logiq-windows-amd64.exe ./cmd/logiq
```

**Build for Linux (Ubuntu, Debian, etc.):**

```powershell
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o logiq-linux-amd64 ./cmd/logiq
```

**Build for macOS (Intel Processors):**

```powershell
$env:GOOS="darwin"; $env:GOARCH="amd64"; go build -o logiq-macos-intel ./cmd/logiq
```

**Build for macOS (Apple Silicon - M1/M2/M3 Processors):**

```powershell
$env:GOOS="darwin"; $env:GOARCH="arm64"; go build -o logiq-macos-arm64 ./cmd/logiq
```

_Note: The commands above are formatted for PowerShell on Windows. If you are using Git Bash or a Linux/Mac terminal, use the format: `GOOS=linux GOARCH=amd64 go build -o logiq-linux-amd64 ./cmd/logiq`._

### Distributing Manual Builds

Once the binary is built (e.g., `logiq-windows-amd64.exe`), you can send this single file directly to users. They can drop it into any folder, open their terminal, and immediately run `.\logiq-windows-amd64.exe run npm run build` without installing any dependencies.
