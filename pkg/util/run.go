package util

import (
	"bytes"
	"os/exec"
)

func Run(combinedCommand string) (stdout string, stderr string, exitCode int) {
	cmd := exec.Command("sh", "-c", combinedCommand)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	exitCode = 0

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// The command has exited with a non-zero exit code
			exitCode = exitErr.ExitCode()
		} else {
			// There was an error executing the command, setting exit code to -1 to indicate failure
			exitCode = -1
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}
