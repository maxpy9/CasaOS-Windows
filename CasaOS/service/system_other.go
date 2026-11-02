//go:build !windows

package service

import (
	"strings"

	"github.com/IceWhaleTech/CasaOS-Common/utils/command"
	exec2 "github.com/IceWhaleTech/CasaOS-Common/utils/exec"
	"github.com/IceWhaleTech/CasaOS/pkg/config"
)

func platformGetNet(physics bool) []string {
	t := "1"
	if physics {
		t = "2"
	}

	if output, err := command.OnlyExec("source " + config.AppInfo.ShellPath + "/helper.sh ;GetNetCard " + t); err != nil {
		return []string{}
	} else {
		return strings.Split(output, "\n")
	}
}

func platformGetNetState(name string) string {
	if output, err := command.OnlyExec("source " + config.AppInfo.ShellPath + "/helper.sh ;CatNetCardState " + name); err != nil {
		return ""
	} else {
		return output
	}
}

func platformGetTimeZone() string {
	if output, err := command.OnlyExec("source " + config.AppInfo.ShellPath + "/helper.sh ;GetTimeZone"); err != nil {
		return ""
	} else {
		return output
	}
}

func platformReboot() error {
	arg := []string{"6"}
	cmd := exec2.Command("init", arg...)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func platformShutdown() error {
	arg := []string{"0"}
	cmd := exec2.Command("init", arg...)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func platformVirtualDir(path string) ([]string, bool) { return nil, false }

func platformIsProtected(path string) bool { return false }
