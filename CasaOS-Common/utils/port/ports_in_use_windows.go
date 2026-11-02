//go:build windows

package port

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/samber/lo"
	"golang.org/x/sys/windows"
)

const (
	afInet  = 2
	afInet6 = 23

	tcpTableOwnerPidAll = 5
	udpTableOwnerPid    = 1

	// sizes of MIB_TCPROW_OWNER_PID, MIB_TCP6ROW_OWNER_PID, MIB_UDPROW_OWNER_PID and MIB_UDP6ROW_OWNER_PID
	tcpRowSize  = 24
	tcp6RowSize = 56
	udpRowSize  = 12
	udp6RowSize = 28

	// offset of the local port in the rows above
	tcpPortOffset  = 8
	tcp6PortOffset = 20
	udpPortOffset  = 4
	udp6PortOffset = 20
)

var (
	iphlpapi            = windows.NewLazySystemDLL("iphlpapi.dll")
	getExtendedTCPTable = iphlpapi.NewProc("GetExtendedTcpTable")
	getExtendedUDPTable = iphlpapi.NewProc("GetExtendedUdpTable")
)

func readPorts(proc *windows.LazyProc, family, class uintptr, rowSize, portOffset int) ([]int, error) {
	var size uint32

	// the first call returns the size of the table in `size`
	proc.Call(0, uintptr(unsafe.Pointer(&size)), 0, family, class, 0) // nolint: errcheck
	if size == 0 {
		return nil, nil
	}

	buffer := make([]byte, size+1024)
	size = uint32(len(buffer))

	result, _, _ := proc.Call(uintptr(unsafe.Pointer(&buffer[0])), uintptr(unsafe.Pointer(&size)), 0, family, class, 0)
	if result != 0 {
		return nil, fmt.Errorf("%s failed with code %d", proc.Name, result)
	}

	count := int(binary.LittleEndian.Uint32(buffer))
	ports := make([]int, 0, count)

	for i := 0; i < count; i++ {
		start := 4 + i*rowSize + portOffset
		if start+2 > len(buffer) {
			break
		}
		// the port is in network byte order
		ports = append(ports, int(binary.BigEndian.Uint16(buffer[start:])))
	}

	return ports, nil
}

func ListPortsInUse() ([]int, []int, error) {
	tcp, err := readPorts(getExtendedTCPTable, afInet, tcpTableOwnerPidAll, tcpRowSize, tcpPortOffset)
	if err != nil {
		return nil, nil, err
	}

	tcp6, err := readPorts(getExtendedTCPTable, afInet6, tcpTableOwnerPidAll, tcp6RowSize, tcp6PortOffset)
	if err != nil {
		return nil, nil, err
	}

	udp, err := readPorts(getExtendedUDPTable, afInet, udpTableOwnerPid, udpRowSize, udpPortOffset)
	if err != nil {
		return nil, nil, err
	}

	udp6, err := readPorts(getExtendedUDPTable, afInet6, udpTableOwnerPid, udp6RowSize, udp6PortOffset)
	if err != nil {
		return nil, nil, err
	}

	return lo.Union(tcp, tcp6), lo.Union(udp, udp6), nil
}
