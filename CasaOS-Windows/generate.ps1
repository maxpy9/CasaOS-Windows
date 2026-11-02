# Generates the API code of the services from their OpenAPI files, what `go generate` does
# on Linux through bash. The output goes in the codegen folders, which are not versioned.
param([string]$Tools = "", [string]$Cache = "$env:LOCALAPPDATA\casaos-build")

$ErrorActionPreference = 'Stop'
$here = Split-Path -Parent $MyInvocation.MyCommand.Path
$root = Split-Path -Parent $here
if (-not $Tools) { $Tools = "$root\_tools" }
. "$here\env.ps1"
New-Item -ItemType Directory -Force $Cache | Out-Null

if (-not (Test-Path "$Cache\gopath\bin\oapi-codegen.exe")) {
    Push-Location $Cache
    $ErrorActionPreference = 'Continue'
    & go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.12.4 2>&1 | Out-Null
    $code = $LASTEXITCODE
    $ErrorActionPreference = 'Stop'
    Pop-Location
    if ($code -ne 0) { throw "oapi-codegen could not be installed" }
}

$messageBus = "$root\CasaOS-MessageBus\api\message_bus\openapi.yaml"

function Generate($repo, $spec, $package, $what, $out) {
    $target = "$root\$repo\$out"
    New-Item -ItemType Directory -Force (Split-Path $target) | Out-Null
    $arguments = @()
    if ($what) { $arguments += "-generate"; $arguments += $what }
    $arguments += "-package"; $arguments += $package; $arguments += $spec
    $text = & oapi-codegen @arguments
    if ($LASTEXITCODE -ne 0) { throw "code generation failed for $repo\$out" }
    [IO.File]::WriteAllText($target, ($text -join "`n") + "`n", (New-Object Text.UTF8Encoding($false)))
}

Generate "CasaOS" "$root\CasaOS\api\casaos\openapi.yaml" "codegen" "types,server,spec" "codegen\casaos_api.go"
Generate "CasaOS" $messageBus "message_bus" "types,client" "codegen\message_bus\api.go"
Generate "CasaOS-MessageBus" $messageBus "codegen" "types,server,spec" "codegen\message_bus_api.go"
Generate "CasaOS-UserService" "$root\CasaOS-UserService\api\user-service\openapi.yaml" "codegen" "types,server,spec" "codegen\user_service\user_service_api.go"
Generate "CasaOS-UserService" $messageBus "message_bus" "" "codegen\message_bus\api.go"
Generate "CasaOS-LocalStorage" "$root\CasaOS-LocalStorage\api\local_storage\openapi.yaml" "codegen" "types,server,spec" "codegen\local_storage_api.go"
Generate "CasaOS-LocalStorage" $messageBus "message_bus" "types,client" "codegen\message_bus\api.go"
Generate "CasaOS-AppManagement" "$root\CasaOS-AppManagement\api\app_management\openapi.yaml" "codegen" "types,server,spec" "codegen\app_management_api.go"
Generate "CasaOS-AppManagement" $messageBus "message_bus" "types,client" "codegen\message_bus\api.go"

"API code generated"
