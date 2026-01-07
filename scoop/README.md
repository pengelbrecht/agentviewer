# Scoop Bucket for agentviewer

This directory contains the Scoop manifest for agentviewer on Windows.

## Setting Up the Scoop Bucket

1. **Create a new repository** named `scoop-agentviewer` at:
   https://github.com/pengelbrecht/scoop-agentviewer

2. **Clone the repository** and copy the manifest:
   ```powershell
   git clone https://github.com/pengelbrecht/scoop-agentviewer
   cd scoop-agentviewer
   mkdir bucket
   Copy-Item /path/to/agentviewer/scoop/agentviewer.json bucket/
   ```

3. **Update SHA256 checksum** after a release:
   ```powershell
   # Download binary and compute SHA256
   $url = "https://github.com/pengelbrecht/agentviewer/releases/download/v0.1.0/agentviewer-windows-amd64.exe"
   $hash = (Get-FileHash -Algorithm SHA256 -InputStream (Invoke-WebRequest $url).RawContentStream).Hash
   Write-Host "SHA256: $hash"
   ```

   Or using curl:
   ```bash
   curl -sL https://github.com/pengelbrecht/agentviewer/releases/download/v0.1.0/agentviewer-windows-amd64.exe | sha256sum
   ```

4. **Commit and push** the manifest:
   ```powershell
   git add bucket/agentviewer.json
   git commit -m "Add agentviewer manifest"
   git push
   ```

## Installing via Scoop

Once the bucket is set up:

```powershell
scoop bucket add agentviewer https://github.com/pengelbrecht/scoop-agentviewer
scoop install agentviewer
```

Or without adding the bucket:

```powershell
scoop install https://raw.githubusercontent.com/pengelbrecht/scoop-agentviewer/main/bucket/agentviewer.json
```

## Updating the Manifest

After each release:

1. Update the `version` in the manifest
2. Recalculate SHA256 checksum
3. Replace the placeholder hash
4. Commit and push to the bucket repository

The manifest includes `autoupdate` configuration, so you can use Scoop's excavator or similar tools to automate updates.

### Automation Script

PowerShell script to update the manifest:

```powershell
param(
    [string]$Version = "0.1.0"
)

$url = "https://github.com/pengelbrecht/agentviewer/releases/download/v$Version/agentviewer-windows-amd64.exe"

Write-Host "Fetching checksum for v$Version..."

$response = Invoke-WebRequest -Uri $url -UseBasicParsing
$hash = (Get-FileHash -Algorithm SHA256 -InputStream $response.RawContentStream).Hash

Write-Host "windows-amd64: $hash"

# Update manifest
$manifest = Get-Content bucket/agentviewer.json | ConvertFrom-Json
$manifest.version = $Version
$manifest.architecture.'64bit'.url = $url
$manifest.architecture.'64bit'.hash = $hash
$manifest | ConvertTo-Json -Depth 10 | Set-Content bucket/agentviewer.json
```

## Testing the Manifest Locally

```powershell
# Install from local file
scoop install ./agentviewer.json

# Verify installation
agentviewer --version

# Uninstall
scoop uninstall agentviewer
```

## Scoop Bucket Structure

The bucket repository should have this structure:

```
scoop-agentviewer/
├── bucket/
│   └── agentviewer.json
└── README.md
```
