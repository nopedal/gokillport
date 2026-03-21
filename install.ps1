# gokillport installer for Windows
# Usage: irm https://raw.githubusercontent.com/nopedal/gokillport/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$Repo = "nopedal/gokillport"
$Binary = "gokillport"
$Alias = "killport"

function Get-InstallDir {
    $dir = Join-Path $env:LOCALAPPDATA "gokillport" "bin"
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
    }
    return $dir
}

function Get-LatestVersion {
    try {
        $release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
        return $release.tag_name
    } catch {
        Write-Host "error: could not determine latest version" -ForegroundColor Red
        Write-Host "  check https://github.com/$Repo/releases" -ForegroundColor DarkGray
        exit 1
    }
}

function Add-ToPath($dir) {
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentPath -notlike "*$dir*") {
        [Environment]::SetEnvironmentVariable("Path", "$currentPath;$dir", "User")
        $env:Path = "$env:Path;$dir"
        Write-Host "  added $dir to PATH" -ForegroundColor DarkGray
    }
}

function Install-Gokillport {
    $version = Get-LatestVersion
    $installDir = Get-InstallDir
    $arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
    $filename = "$Binary-windows-$arch.exe"
    $url = "https://github.com/$Repo/releases/download/$version/$filename"

    Write-Host "> downloading $Binary $version for windows/$arch" -ForegroundColor Green

    $tmpFile = Join-Path $env:TEMP "$Binary.exe"

    try {
        Invoke-WebRequest -Uri $url -OutFile $tmpFile -UseBasicParsing
    } catch {
        Write-Host "error: download failed" -ForegroundColor Red
        Write-Host "  check $url" -ForegroundColor DarkGray
        exit 1
    }

    # Install binary
    $dest = Join-Path $installDir "$Binary.exe"
    Move-Item -Path $tmpFile -Destination $dest -Force

    # Create killport alias (copy of the binary)
    $aliasDest = Join-Path $installDir "$Alias.exe"
    Copy-Item -Path $dest -Destination $aliasDest -Force

    # Add to PATH
    Add-ToPath $installDir

    Write-Host "> installed $Binary $version to $installDir" -ForegroundColor Green
    Write-Host "> also available as: $Alias" -ForegroundColor Green
    Write-Host ""
    Write-Host "  usage: killport 3000" -ForegroundColor DarkGray
    Write-Host ""
    Write-Host "  restart your terminal for PATH changes to take effect" -ForegroundColor DarkGray
}

Install-Gokillport
