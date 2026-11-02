//go:build !windows

package port

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/samber/lo"
)

func ListPortsInUse() ([]int, []int, error) {
	usedPorts := map[string]map[int]struct{}{
		"tcp":  {},
		"udp":  {},
		"tcp6": {},
		"udp6": {},
	}

	for _, protocol := range lo.Keys(usedPorts) {
		filename := fmt.Sprintf("/proc/net/%s", protocol)

		file, err := os.Open(filename)
		if err != nil {
			fmt.Println("Warning: ", err.Error())
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}

			localAddress := fields[1]
			addressParts := strings.Split(localAddress, ":")
			if len(addressParts) < 2 {
				continue
			}

			portHex := addressParts[1]
			port, err := strconv.ParseInt(portHex, 16, 0)
			if err != nil {
				continue
			}

			usedPorts[protocol][int(port)] = struct{}{}
		}

		if err := scanner.Err(); err != nil {
			return nil, nil, errors.New("Error reading from " + filename)
		}
	}

	return lo.Union(
			lo.Keys(usedPorts["tcp"]), lo.Keys(usedPorts["tcp6"]),
		),
		lo.Union(
			lo.Keys(usedPorts["udp"]), lo.Keys(usedPorts["udp6"]),
		),
		nil
}
