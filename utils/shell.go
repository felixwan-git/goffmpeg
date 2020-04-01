package utils

import (
	"bytes"
	"os/exec"
)

func ExecCommand(command string, arguments ...string) (string, error, string) {
	cmd := exec.Command(command, arguments...)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()

	return out.String(), err, stderr.String()
}
