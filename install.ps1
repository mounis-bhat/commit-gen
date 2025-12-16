# commit-gen installer for Windows
# Usage: irm https://raw.githubusercontent.com/mounis-bhat/commit-gen/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$Repo = "mounis-bhat/commit-gen"
$BinaryName = "commit-gen"
$InstallDir = "$env:USERPROFILE\.local\bin"

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] " -ForegroundColor Green -NoNewline
    Write-Host $Message
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] " -ForegroundColor Yellow -NoNewline
    Write-Host $Message
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] " -ForegroundColor Red -NoNewline
    Write-Host $Message
    exit 1
}

function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { Write-Error "Unsupported architecture: $arch" }
    }
}

function Get-LatestVersion {
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing
        return $response.tag_name
    }
    catch {
        Write-Error "Could not determine latest version. Please check https://github.com/$Repo/releases"
    }
}

function Install-CommitGen {
    $arch = Get-Architecture
    $version = Get-LatestVersion
    
    Write-Info "Detected Architecture: $arch"
    Write-Info "Latest version: $version"
    
    $filename = "$BinaryName-windows-$arch.zip"
    $downloadUrl = "https://github.com/$Repo/releases/download/$version/$filename"
    
    Write-Info "Downloading $filename..."
    
    # Create temp directory
    $tempDir = New-Item -ItemType Directory -Path ([System.IO.Path]::GetTempPath()) -Name ([System.Guid]::NewGuid().ToString()) -Force
    $zipPath = Join-Path $tempDir $filename
    
    try {
        # Download
        Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath -UseBasicParsing
        
        # Extract
        Write-Info "Extracting..."
        Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force
        
        # Install
        Write-Info "Installing to $InstallDir..."
        if (-not (Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }
        
        $exePath = Join-Path $tempDir "$BinaryName.exe"
        $destPath = Join-Path $InstallDir "$BinaryName.exe"
        Move-Item -Path $exePath -Destination $destPath -Force
        
        Write-Info "Successfully installed $BinaryName to $destPath"
        
        # Check if InstallDir is in PATH
        $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if ($userPath -notlike "*$InstallDir*") {
            Write-Host ""
            Write-Warn "$InstallDir is not in your PATH."
            Write-Host ""
            $addToPath = Read-Host "Would you like to add it to your PATH? (y/n)"
            
            if ($addToPath -eq "y" -or $addToPath -eq "Y") {
                $newPath = "$userPath;$InstallDir"
                [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
                Write-Info "Added $InstallDir to your PATH."
                Write-Host ""
                Write-Warn "Please restart your terminal for the PATH change to take effect."
            }
            else {
                Write-Host ""
                Write-Host "To add it manually, run:"
                Write-Host ""
                Write-Host "    `$env:Path += `";$InstallDir`""
                Write-Host ""
                Write-Host "Or add it permanently via System Properties > Environment Variables"
            }
        }
        
        Write-Host ""
        Write-Info "Installation complete! Run '$BinaryName' to get started."
    }
    finally {
        # Cleanup
        Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

Install-CommitGen
