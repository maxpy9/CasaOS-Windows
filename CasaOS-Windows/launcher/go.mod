module github.com/casaos-windows/launcher

go 1.21

require (
	github.com/IceWhaleTech/CasaOS-Common v0.4.11-alpha4
	golang.org/x/sys v0.20.0
)

require github.com/Microsoft/go-winio v0.6.2

replace github.com/IceWhaleTech/CasaOS-Common => ../../CasaOS-Common
