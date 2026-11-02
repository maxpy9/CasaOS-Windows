//go:build windows

package command

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/IceWhaleTech/CasaOS-Common/utils/vfs"
	"golang.org/x/sys/windows"
)

// Control codes of the Windows storage stack. All of them work on a handle opened
// without any access right, so they do not need administrator privileges.
const (
	ioctlVolumeGetVolumeDiskExtents = 0x00560000
	ioctlStorageQueryProperty       = 0x002D1400
	ioctlDiskGetDriveGeometryEx     = 0x000700A0

	storageDeviceProperty            = 0
	storageDeviceSeekPenaltyProperty = 7
	propertyStandardQuery            = 0
)

// blockDevice has the fields of the JSON output of `lsblk -O -J -b` that CasaOS reads.
type blockDevice struct {
	Name       string        `json:"name"`
	Path       string        `json:"path"`
	Type       string        `json:"type"`
	Size       uint64        `json:"size"`
	Model      string        `json:"model"`
	Vendor     string        `json:"vendor"`
	Serial     string        `json:"serial"`
	Tran       string        `json:"tran"`
	SubSystems string        `json:"subsystems"`
	Rota       bool          `json:"rota"`
	RM         bool          `json:"rm"`
	RO         bool          `json:"ro"`
	Health     string        `json:"health"`
	FsType     string        `json:"fstype"`
	Label      string        `json:"label"`
	UUID       string        `json:"uuid"`
	MountPoint string        `json:"mountpoint"`
	FSSize     *uint64       `json:"fssize"`
	FSAvail    *uint64       `json:"fsavail"`
	FSUsed     *uint64       `json:"fsused"`
	Children   []blockDevice `json:"children"`
}

// S.M.A.R.T. needs smartctl, which is not part of the Windows build.
func ExecSmartCTLByPath(path string) []byte { return nil }

func ExecEnabledSMART(path string) ([]byte, error) { return nil, nil }

func openDevice(name string) (windows.Handle, error) {
	p, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return windows.InvalidHandle, err
	}
	return windows.CreateFile(p, 0, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, 0, 0)
}

func ioctl(h windows.Handle, code uint32, in, out []byte) (uint32, error) {
	var inPtr, outPtr *byte
	if len(in) > 0 {
		inPtr = &in[0]
	}
	if len(out) > 0 {
		outPtr = &out[0]
	}
	var returned uint32
	err := windows.DeviceIoControl(h, code, inPtr, uint32(len(in)), outPtr, uint32(len(out)), &returned, nil)
	return returned, err
}

func storageQuery(h windows.Handle, property uint32, out []byte) (uint32, error) {
	in := make([]byte, 12) // STORAGE_PROPERTY_QUERY
	binary.LittleEndian.PutUint32(in[0:], property)
	binary.LittleEndian.PutUint32(in[4:], propertyStandardQuery)
	return ioctl(h, ioctlStorageQueryProperty, in, out)
}

// cString reads the zero-terminated string at the given offset of a descriptor.
func cString(data []byte, offset uint32) string {
	if offset == 0 || int(offset) >= len(data) {
		return ""
	}
	end := int(offset)
	for end < len(data) && data[end] != 0 {
		end++
	}
	return strings.TrimSpace(string(data[offset:end]))
}

func busName(bus uint32) string {
	switch bus {
	case 7:
		return "usb"
	case 10:
		return "sas"
	case 17:
		return "nvme"
	default:
		return "sata"
	}
}

func queryDisk(number uint32) (blockDevice, bool) {
	name := fmt.Sprintf("PhysicalDrive%d", number)
	disk := blockDevice{
		Name:   name,
		Path:   "/dev/" + name,
		Type:   "disk",
		Health: "OK",
	}

	h, err := openDevice(`\\.\` + name)
	if err != nil {
		return disk, false
	}
	defer windows.CloseHandle(h) // nolint: errcheck

	descriptor := make([]byte, 4096) // STORAGE_DEVICE_DESCRIPTOR
	if n, err := storageQuery(h, storageDeviceProperty, descriptor); err == nil && n >= 32 {
		d := descriptor[:n]
		disk.RM = d[10] != 0
		disk.Vendor = cString(d, binary.LittleEndian.Uint32(d[12:]))
		disk.Model = cString(d, binary.LittleEndian.Uint32(d[16:]))
		disk.Serial = cString(d, binary.LittleEndian.Uint32(d[24:]))
		disk.Tran = busName(binary.LittleEndian.Uint32(d[28:]))
		disk.SubSystems = "block:" + disk.Tran
	}
	if disk.Model == "" {
		disk.Model = name
	}
	if disk.Serial == "" {
		disk.Serial = name
	}

	seekPenalty := make([]byte, 12) // DEVICE_SEEK_PENALTY_DESCRIPTOR
	if n, err := storageQuery(h, storageDeviceSeekPenaltyProperty, seekPenalty); err == nil && n >= 9 {
		disk.Rota = seekPenalty[8] != 0
	}

	geometry := make([]byte, 256) // DISK_GEOMETRY_EX
	if n, err := ioctl(h, ioctlDiskGetDriveGeometryEx, nil, geometry); err == nil && n >= 32 {
		disk.Size = binary.LittleEndian.Uint64(geometry[24:])
	}

	return disk, true
}

func fsName(name string) string {
	switch strings.ToUpper(name) {
	case "NTFS":
		return "ntfs-3g"
	case "FAT", "FAT32":
		return "vfat"
	default:
		return strings.ToLower(name)
	}
}

// queryVolume returns the partition of a drive letter and the number of its disk.
func queryVolume(letter string) (blockDevice, uint32, bool) {
	part := blockDevice{
		Name:   letter + ":",
		Path:   "/dev/" + letter,
		Type:   "part",
		Health: "OK",
	}

	h, err := openDevice(`\\.\` + letter + `:`)
	if err != nil {
		return part, 0, false
	}
	defer windows.CloseHandle(h) // nolint: errcheck

	extents := make([]byte, 1024) // VOLUME_DISK_EXTENTS
	n, err := ioctl(h, ioctlVolumeGetVolumeDiskExtents, nil, extents)
	if err != nil || n < 32 {
		return part, 0, false
	}
	number := binary.LittleEndian.Uint32(extents[8:])
	part.Size = binary.LittleEndian.Uint64(extents[24:])

	// a drive without any space on a disk is virtual, as the one of Google Drive is:
	// it reports the size of the disk that holds its cache, which would be counted twice
	if part.Size == 0 {
		return part, 0, false
	}

	root := letter + `:\`
	rootPtr, err := windows.UTF16PtrFromString(root)
	if err != nil {
		return part, 0, false
	}

	label := make([]uint16, windows.MAX_PATH+1)
	system := make([]uint16, windows.MAX_PATH+1)
	if err := windows.GetVolumeInformation(rootPtr, &label[0], uint32(len(label)), nil, nil, nil, &system[0], uint32(len(system))); err != nil {
		// no media in the drive
		return part, 0, false
	}
	part.Label = windows.UTF16ToString(label)
	part.FsType = fsName(windows.UTF16ToString(system))

	guid := make([]uint16, 64)
	if err := windows.GetVolumeNameForVolumeMountPoint(rootPtr, &guid[0], uint32(len(guid))); err == nil {
		part.UUID = strings.Trim(strings.TrimSuffix(strings.TrimPrefix(windows.UTF16ToString(guid), `\\?\Volume`), `\`), "{}")
	}

	var available, total, free uint64
	if err := windows.GetDiskFreeSpaceEx(rootPtr, &available, &total, &free); err == nil {
		used := total - free
		part.FSSize = &total
		part.FSAvail = &available
		part.FSUsed = &used
	}

	if strings.EqualFold(letter+":", os.Getenv("SystemDrive")) {
		part.MountPoint = "/"
	} else {
		part.MountPoint = "/media/" + letter
	}

	return part, number, true
}

// listBlockDevices returns the disks of the computer that have at least one drive
// letter, with their volumes as children.
func listBlockDevices() []blockDevice {
	byNumber := map[uint32]*blockDevice{}

	for _, letter := range vfs.Drives() {
		rootPtr, err := windows.UTF16PtrFromString(letter + `:\`)
		if err != nil || windows.GetDriveType(rootPtr) == windows.DRIVE_REMOTE {
			continue
		}

		part, number, ok := queryVolume(letter)
		if !ok {
			continue
		}

		disk, found := byNumber[number]
		if !found {
			d, ok := queryDisk(number)
			if !ok {
				continue
			}
			disk = &d
			byNumber[number] = disk
		}

		part.Tran = disk.Tran
		part.SubSystems = disk.SubSystems
		part.Model = disk.Model
		part.RM = disk.RM
		disk.Children = append(disk.Children, part)
	}

	numbers := make([]uint32, 0, len(byNumber))
	for number := range byNumber {
		numbers = append(numbers, number)
	}
	sort.Slice(numbers, func(i, j int) bool { return numbers[i] < numbers[j] })

	devices := make([]blockDevice, 0, len(numbers))
	for _, number := range numbers {
		devices = append(devices, *byNumber[number])
	}
	return devices
}

func marshalDevices(devices []blockDevice) []byte {
	output, err := json.Marshal(map[string][]blockDevice{"blockdevices": devices})
	if err != nil {
		fmt.Println("lsblk", err)
		return nil
	}
	return output
}

// ExecLSBLK returns the same JSON as lsblk does on Linux.
func ExecLSBLK() []byte {
	return marshalDevices(listBlockDevices())
}

// ExecLSBLKByPath returns the same JSON as lsblk does on Linux for one device.
func ExecLSBLKByPath(path string) []byte {
	for _, disk := range listBlockDevices() {
		if disk.Path == path {
			return marshalDevices([]blockDevice{disk})
		}
		for _, part := range disk.Children {
			if part.Path == path {
				return marshalDevices([]blockDevice{part})
			}
		}
	}
	return marshalDevices(nil)
}
