# Downloads what the build needs into ..\_tools, without installing anything on the computer:
# Go, Inno Setup in portable mode and the published dashboard of CasaOS.
param([string]$Tools = "")

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

$here = Split-Path -Parent $MyInvocation.MyCommand.Path
if (-not $Tools) { $Tools = Join-Path (Split-Path -Parent $here) "_tools" }
$downloads = "$Tools\dl"
New-Item -ItemType Directory -Force $downloads, "$Tools\ui" | Out-Null

$goVersion = "go1.27.3"
$goSha256 = "f26081f438319757f8a1b46d7d0b2c5ec7583d4038e485e0aa725cfde7885fff"
$uiVersion = "v0.4.25"
$innoTag = "is-7_1_1"
$innoFile = "innosetup-7.1.1-x64.exe"

if (-not (Test-Path "$Tools\go\bin\go.exe")) {
    "Go $goVersion..."
    Invoke-WebRequest "https://go.dev/dl/$goVersion.windows-amd64.zip" -OutFile "$downloads\go.zip"
    $hash = (Get-FileHash "$downloads\go.zip" -Algorithm SHA256).Hash.ToLower()
    if ($hash -ne $goSha256) { throw "The SHA256 of Go does not match: $hash" }
    tar -xf "$downloads\go.zip" -C $Tools
}

if (-not (Test-Path "$Tools\ui\build\sysroot\var\lib\casaos\www\index.html")) {
    "CasaOS-UI $uiVersion..."
    Invoke-WebRequest "https://github.com/IceWhaleTech/CasaOS-UI/releases/download/$uiVersion/linux-all-casaos-$uiVersion.tar.gz" -OutFile "$downloads\ui.tar.gz"
    tar -xzf "$downloads\ui.tar.gz" -C "$Tools\ui"
}

if (-not (Test-Path "$Tools\inno\ISCC.exe")) {
    "Inno Setup..."
    Invoke-WebRequest "https://github.com/jrsoftware/issrc/releases/download/$innoTag/$innoFile" -OutFile "$downloads\$innoFile"
    if ((Get-AuthenticodeSignature "$downloads\$innoFile").Status -ne 'Valid') { throw "The signature of Inno Setup is not valid" }
    & "$downloads\$innoFile" /VERYSILENT /SUPPRESSMSGBOXES /NORESTART /CURRENTUSER /PORTABLE=1 /NOICONS "/DIR=$Tools\inno" | Out-Null
    $waited = 0
    while (-not (Test-Path "$Tools\inno\ISCC.exe") -and $waited -lt 120) { Start-Sleep 2; $waited += 2 }
    if (-not (Test-Path "$Tools\inno\ISCC.exe")) { throw "Inno Setup was not extracted" }
}

"Tools ready in $Tools"
