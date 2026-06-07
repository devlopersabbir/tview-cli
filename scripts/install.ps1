$ErrorActionPreference = "Stop"

$Repo = "devlopersabbir/tview-cli"
$Binary = "tview.exe"
$Version = if ($env:TVIEW_VERSION) { $env:TVIEW_VERSION } else { "latest" }
$InstallDir = if ($env:TVIEW_INSTALL_DIR) { $env:TVIEW_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "Programs\tview\bin" }

if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq "Arm64") {
    $Arch = "arm64"
} else {
    $Arch = "amd64"
}

if ($Version -eq "latest") {
    $Url = "https://github.com/$Repo/releases/latest/download/tview_windows_$Arch.zip"
} else {
    if ($Version.StartsWith("v")) {
        $Tag = $Version
    } else {
        $Tag = "v$Version"
    }
    $Url = "https://github.com/$Repo/releases/download/$Tag/tview_windows_$Arch.zip"
}

$Temp = Join-Path ([System.IO.Path]::GetTempPath()) ("tview-" + [System.Guid]::NewGuid())
$Zip = Join-Path $Temp "tview.zip"

New-Item -ItemType Directory -Force -Path $Temp, $InstallDir | Out-Null

try {
    Write-Host "Downloading tview for windows/$Arch..."
    Invoke-WebRequest -Uri $Url -OutFile $Zip
    Expand-Archive -Path $Zip -DestinationPath $Temp -Force
    Copy-Item -Path (Join-Path $Temp $Binary) -Destination (Join-Path $InstallDir $Binary) -Force

    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $PathParts = @()
    if ($UserPath) {
        $PathParts = $UserPath -split ";"
    }

    if ($PathParts -notcontains $InstallDir) {
        $NewPath = if ($UserPath) { "$UserPath;$InstallDir" } else { $InstallDir }
        [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
        $env:Path = "$env:Path;$InstallDir"
        Write-Host "Added $InstallDir to your user PATH. Restart your terminal if tview is not found."
    }

    Write-Host "Installed tview to $(Join-Path $InstallDir $Binary)"
    & (Join-Path $InstallDir $Binary) version
} finally {
    Remove-Item -Recurse -Force $Temp -ErrorAction SilentlyContinue
}
