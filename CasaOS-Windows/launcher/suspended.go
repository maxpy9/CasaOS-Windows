package main

import (
	"encoding/binary"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	systemProcessInformation = 5

	// layout of SYSTEM_PROCESS_INFORMATION and SYSTEM_THREAD_INFORMATION on 64 bit
	processThreadCountOffset = 4
	processIDOffset          = 80
	processHeaderSize        = 256
	threadInfoSize           = 80
	threadStateOffset        = 68
	threadWaitReasonOffset   = 72

	threadStateWaiting   = 5
	waitReasonSuspended  = 5
	statusInfoLenMismach = 0xC0000004
)

// anyProcessSuspended tells if one of the processes has all of its threads suspended.
func anyProcessSuspended(pids []uint32) bool {
	if len(pids) == 0 {
		return false
	}

	wanted := make(map[uint64]struct{}, len(pids))
	for _, pid := range pids {
		wanted[uint64(pid)] = struct{}{}
	}

	size := uint32(1 << 20)
	var buffer []byte
	for {
		buffer = make([]byte, size)
		var needed uint32
		err := windows.NtQuerySystemInformation(systemProcessInformation, unsafe.Pointer(&buffer[0]), size, &needed)
		if err == nil {
			break
		}
		if err != windows.NTStatus(statusInfoLenMismach) || size > 1<<27 {
			return false
		}
		size *= 2
	}

	for offset := uint32(0); int(offset)+processHeaderSize <= len(buffer); {
		entry := buffer[offset:]
		next := binary.LittleEndian.Uint32(entry)
		threads := binary.LittleEndian.Uint32(entry[processThreadCountOffset:])
		pid := binary.LittleEndian.Uint64(entry[processIDOffset:])

		if _, ok := wanted[pid]; ok && threads > 0 && processHeaderSize+int(threads)*threadInfoSize <= len(entry) {
			suspended := uint32(0)
			for i := uint32(0); i < threads; i++ {
				thread := entry[processHeaderSize+i*threadInfoSize:]
				state := binary.LittleEndian.Uint32(thread[threadStateOffset:])
				reason := binary.LittleEndian.Uint32(thread[threadWaitReasonOffset:])
				if state == threadStateWaiting && reason == waitReasonSuspended {
					suspended++
				}
			}
			if suspended == threads {
				return true
			}
		}

		if next == 0 {
			break
		}
		offset += next
	}

	return false
}
