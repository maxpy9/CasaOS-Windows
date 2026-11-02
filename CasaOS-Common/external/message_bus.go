package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/IceWhaleTech/CasaOS-Common/utils/constants"
	"github.com/IceWhaleTech/CasaOS-Common/utils/vfs"
)

const (
	MessageBusAddressFilename = "message-bus.url"
	APIMessageBus             = "/v2/message_bus"
)

type EventType struct {
	Name             string
	SourceID         string
	PropertyTypeList []PropertyType
}

type PropertyType struct {
	Name        string
	Description *string
	Example     *string
}

func PrintEventTypesAsMarkdown(sourceID, version string, eventTypes []EventType) {
	fmt.Printf("## Source ID: `%s` (v%s)\n\n", sourceID, version)
	for _, eventType := range eventTypes {
		fmt.Printf("### Event Type: `%s`\n\n", eventType.Name)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

		fmt.Fprintln(w, "| Property\t| Description\t| Example\t|")
		fmt.Fprintln(w, "| --------\t| -----------\t| -------\t|")

		for _, propertyType := range eventType.PropertyTypeList {
			fmt.Fprintf(w, "| `%s`\t|", propertyType.Name)

			if propertyType.Description != nil {
				fmt.Fprintf(w, " %s", *propertyType.Description)
			}

			fmt.Fprintf(w, "\t|")

			if propertyType.Example != nil {
				fmt.Fprintf(w, " `%s`", *propertyType.Example)
			}

			fmt.Fprintln(w, "\t|")
		}
		w.Flush()

		fmt.Println()
	}
}

func GetMessageBusAddress(runtimePath string) (string, error) {
	address, err := getAddress(filepath.Join(runtimePath, MessageBusAddressFilename))
	if err != nil {
		return "", err
	}

	return strings.TrimRight(address, "/") + APIMessageBus, nil
}

func PublishEventInSocket(ctx context.Context, SourceID string, Name string, properties map[string]string) (*http.Response, error) {
	base := "http://unix" + APIMessageBus
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", vfs.SocketPath("/tmp/message-bus.sock"))
			},
		},
	}

	// on Windows the message bus only listens on its TCP port: the file of a Unix socket stays
	// behind when the process is terminated and it may not be possible to delete it
	if runtime.GOOS == "windows" {
		address, err := GetMessageBusAddress(constants.DefaultRuntimePath)
		if err != nil {
			return nil, err
		}
		base = address
		httpClient = &http.Client{}
	}

	body, err := json.Marshal(properties)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/event/%s/%s", base, SourceID, Name),
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()
	return resp, nil
}
