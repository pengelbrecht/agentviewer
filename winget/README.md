# Winget Manifest for agentviewer

This directory contains the Windows Package Manager (winget) manifest for agentviewer.

## Overview

Winget requires submitting a Pull Request to the official winget-pkgs repository:
https://github.com/microsoft/winget-pkgs

## Manifest Structure

Winget uses a multi-file manifest format:

- `pengelbrecht.agentviewer.yaml` - Version manifest (required)
- `pengelbrecht.agentviewer.locale.en-US.yaml` - Locale/description manifest (required)
- `pengelbrecht.agentviewer.installer.yaml` - Installer manifest (required)

## Submitting to Winget

### Prerequisites

1. **GitHub Account**: Required to submit PRs
2. **Winget-Create Tool** (optional but recommended):
   ```powershell
   winget install wingetcreate
   ```

### Option 1: Using Winget-Create (Recommended)

The easiest way to submit a new package:

```powershell
wingetcreate submit https://github.com/pengelbrecht/agentviewer/releases/download/v0.1.0/agentviewer-windows-amd64.exe
```

This will:
1. Download the installer
2. Generate manifests automatically
3. Create a PR to winget-pkgs

### Option 2: Manual Submission

1. **Fork** the winget-pkgs repository:
   https://github.com/microsoft/winget-pkgs

2. **Clone** your fork:
   ```powershell
   git clone https://github.com/YOUR_USERNAME/winget-pkgs
   cd winget-pkgs
   ```

3. **Create the package directory**:
   ```powershell
   mkdir -p manifests/p/pengelbrecht/agentviewer/0.1.0
   ```

4. **Copy manifests**:
   ```powershell
   Copy-Item /path/to/agentviewer/winget/*.yaml manifests/p/pengelbrecht/agentviewer/0.1.0/
   ```

5. **Update SHA256 hash** in the installer manifest:
   ```powershell
   $url = "https://github.com/pengelbrecht/agentviewer/releases/download/v0.1.0/agentviewer-windows-amd64.exe"
   $response = Invoke-WebRequest -Uri $url -UseBasicParsing
   $hash = (Get-FileHash -Algorithm SHA256 -InputStream $response.RawContentStream).Hash
   Write-Host "SHA256: $hash"
   ```

6. **Validate manifests**:
   ```powershell
   winget validate manifests/p/pengelbrecht/agentviewer/0.1.0/
   ```

7. **Commit and push**:
   ```powershell
   git add manifests/p/pengelbrecht/agentviewer/
   git commit -m "Add pengelbrecht.agentviewer version 0.1.0"
   git push origin main
   ```

8. **Create Pull Request** to microsoft/winget-pkgs

## Updating the Package

For new releases:

1. Create a new version directory:
   ```powershell
   mkdir manifests/p/pengelbrecht/agentviewer/0.2.0
   ```

2. Copy and update the manifests with new version and SHA256

3. Submit a new PR

### Using Winget-Create for Updates

```powershell
wingetcreate update pengelbrecht.agentviewer --version 0.2.0 --urls https://github.com/pengelbrecht/agentviewer/releases/download/v0.2.0/agentviewer-windows-amd64.exe --submit
```

## Installation (After Publishing)

Once the package is published to the winget repository:

```powershell
winget install agentviewer

# Or with full package identifier
winget install pengelbrecht.agentviewer
```

## Local Testing

Before submitting, test the manifest locally:

```powershell
# Validate manifest syntax
winget validate ./

# Test installation from manifest
winget install --manifest ./
```

## CI/CD Integration

You can automate winget submissions using GitHub Actions:

```yaml
- name: Submit to winget
  uses: vedantmgoyal2009/winget-releaser@v2
  with:
    identifier: pengelbrecht.agentviewer
    installers-regex: 'agentviewer-windows-amd64\.exe$'
    token: ${{ secrets.WINGET_TOKEN }}
```

## Notes

- Winget review process typically takes 1-3 days
- Portable installers (standalone .exe) are supported
- The package identifier format is: `Publisher.PackageName`
- Version numbers should follow semver (e.g., 0.1.0)
