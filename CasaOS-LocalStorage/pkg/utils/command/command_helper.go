package command

import (
	"errors"
	"os/exec"

	exec2 "github.com/IceWhaleTech/CasaOS-Common/utils/exec"
)

func ExecuteCommand(name string, arg ...string) ([]byte, error) {
	cmd := exec2.Command(name, arg...)
	println(cmd.String())

	out, err := cmd.Output()
	println(string(out))
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			message := string(exitError.Stderr)
			println(message)
			return nil, errors.New(message)
		}
		return nil, err
	}

	return out, nil
}
