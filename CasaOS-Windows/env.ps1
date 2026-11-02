# Environment shared by the build scripts. Dot-source it after setting $Tools and $Cache.
#
# The cache of Go stays out of the folders protected by the antivirus, where renaming a
# file that was just downloaded can fail.
$env:GOROOT = "$Tools\go"
$env:GOPATH = "$Cache\gopath"
$env:GOMODCACHE = "$Cache\gopath\pkg\mod"
$env:GOCACHE = "$Cache\gocache"
$env:GOBIN = "$Cache\gopath\bin"
$env:GOTOOLCHAIN = "local"
$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:Path = "$Tools\go\bin;$Cache\gopath\bin;$env:Path"
