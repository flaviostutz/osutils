package osutils

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-cmd/cmd"
	"github.com/sirupsen/logrus"
)

//ShellContext container to transport a Cmd reference
type ShellContext struct {
	//CmdRef cmd.Cmd pointer that can be used to set command references that should be killed when a backup deletion of a running job is detected
	CmdRef *cmd.Cmd
}

//ExecShellTimeout execute a shell command (like bash -c 'your command') with a timeout. After that time, the process will be cancelled
func ExecShellTimeout(command string, timeout time.Duration, ctx *ShellContext) (string, error) {
	logrus.Debugf("shell command: %s", command)
	acmd := cmd.NewCmd("bash", "-c", command)
	statusChan := acmd.Start() // non-blocking
	running := true
	if ctx != nil {
		ctx.CmdRef = acmd
	}

	//kill if taking too long
	if timeout > 0 {
		logrus.Debugf("Enforcing timeout %s", timeout)
		go func() {
			startTime := time.Now()
			for running {
				if time.Since(startTime) >= timeout {
					logrus.Warnf("Stopping command execution because it is taking too long (%d seconds)", time.Since(startTime))
					acmd.Stop()
				}
				time.Sleep(1 * time.Second)
			}
		}()
	}

	// logrus.Debugf("Waiting for command to finish...")
	<-statusChan
	// logrus.Debugf("Command finished")
	running = false

	out := GetCmdOutput(acmd)
	status := acmd.Status()
	logrus.Debugf("shell output (%d): %s", status.Exit, out)
	if status.Exit != 0 {
		return out, fmt.Errorf("Failed to run command: '%s'; exit=%d; out=%s", command, status.Exit, out)
	}
	return out, nil
}

//ExecShell execute a shell command (like bash -c 'your command')
func ExecShell(command string) (string, error) {
	return ExecShellTimeout(command, 0, nil)
}

//ExecShellf execute a shell command (like bash -c 'your command') but with format replacements
func ExecShellf(command string, args ...interface{}) (string, error) {
	cmd := fmt.Sprintf(command, args...)
	return ExecShellTimeout(cmd, 0, nil)
}

//GetCmdOutput join stdout and stderr in a single string from Cmd
func GetCmdOutput(cmd *cmd.Cmd) string {
	status := cmd.Status()
	out := strings.Join(status.Stdout, "\n")
	out = out + "\n" + strings.Join(status.Stderr, "\n")
	return out
}

//Trunc truncates string
func Trunc(str string, num int) string {
	bnoden := str
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		bnoden = str[0:num]
	}
	return bnoden
}
