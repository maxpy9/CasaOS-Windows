# Builds the installer of CasaOS for Windows: dist\CasaOS-Setup-<version>.exe
# Run setup-tools.ps1 once before it.
param(
    [string]$Version = "0.4.15",
    [string]$Tools = "",
    [string]$Cache = "$env:LOCALAPPDATA\casaos-build"
)

$ErrorActionPreference = 'Stop'
$here = Split-Path -Parent $MyInvocation.MyCommand.Path
$root = Split-Path -Parent $here
if (-not $Tools) { $Tools = "$root\_tools" }
if (-not (Test-Path "$Tools\go\bin\go.exe")) { throw "Tools not found in ${Tools}: run setup-tools.ps1 first" }
$sysroot = "$Tools\ui\build\sysroot\var\lib\casaos"
$app = "$Cache\app"

. "$here\env.ps1"
$env:GOFLAGS = "-mod=mod"

New-Item -ItemType Directory -Force "$app\bin", "$here\dist" | Out-Null

if (-not (Test-Path "$root\CasaOS\codegen\casaos_api.go")) {
    $env:GOFLAGS = ""
    & "$here\generate.ps1" -Tools $Tools -Cache $Cache
    $env:GOFLAGS = "-mod=mod"
}

function Go-Build($dir, $out, $ldflags) {
    Push-Location $dir
    try {
        # go writes its progress to stderr, which is not an error here
        $ErrorActionPreference = 'Continue'
        $o = & go build -trimpath -buildvcs=false -ldflags $ldflags -o $out . 2>&1 | ForEach-Object { "$_" } | Where-Object { $_ -notmatch '^go: (downloading|finding|found)' }
        $code = $LASTEXITCODE
        $ErrorActionPreference = 'Stop'
        if ($code -ne 0) { $o; throw "compilation of $dir failed" }
    } finally { Pop-Location }
    "OK {0,-28} {1,6:N1} MB" -f (Split-Path -Leaf $out), ((Get-Item $out).Length / 1MB)
}

# services: the launcher runs them with a hidden console
$services = [ordered]@{
    'CasaOS-Gateway'       = 'casaos-gateway.exe'
    'CasaOS-MessageBus'    = 'casaos-message-bus.exe'
    'CasaOS-UserService'   = 'casaos-user-service.exe'
    'CasaOS'               = 'casaos.exe'
    'CasaOS-LocalStorage'  = 'casaos-local-storage.exe'
    'CasaOS-AppManagement' = 'casaos-app-management.exe'
}
foreach ($repo in $services.Keys) {
    Go-Build "$root\$repo" "$app\bin\$($services[$repo])" "-s -w -X main.commit=windows -X main.date=$(Get-Date -Format yyyy-MM-dd)"
}

# launcher, with its icon and its manifest
$launcher = "$here\launcher"
Push-Location $launcher
$ErrorActionPreference = 'Continue'
if (-not (Test-Path "$Cache\gopath\bin\rsrc.exe")) { $env:GOFLAGS = ""; & go install github.com/akavel/rsrc@v0.10.2 2>&1 | Out-Null; $env:GOFLAGS = "-mod=mod" }
& "$Cache\gopath\bin\rsrc.exe" -manifest launcher.manifest -ico "assets\icon.ico" -arch amd64 -o rsrc_windows_amd64.syso
$ErrorActionPreference = 'Stop'
Pop-Location
Go-Build $launcher "$app\CasaOS.exe" "-s -w -H windowsgui"

# dashboard
New-Item -ItemType Directory -Force "$app\www" | Out-Null
Copy-Item "$sysroot\www\*" "$app\www\" -Recurse -Force
Copy-Item "$sysroot\ui-message-bus.json" "$app\" -Force

# installer
& "$Tools\inno\ISCC.exe" /Q "/DAppVersion=$Version" "/DSourceDir=$app" "/DOutputDir=$here\dist" "$here\installer\casaos.iss"
if ($LASTEXITCODE -ne 0) { throw "the installer failed" }
$setup = Get-Item "$here\dist\CasaOS-Setup-$Version.exe"
"OK {0}  {1:N1} MB" -f $setup.FullName, ($setup.Length / 1MB)
