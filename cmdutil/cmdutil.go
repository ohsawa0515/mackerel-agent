package cmdutil

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/Songmu/timeout"
	"github.com/mackerelio/golib/logging"
)

var logger = logging.GetLogger("cmdutil")

const (
	// defaultTimeoutDuration is the duration after which a command execution will be timeout.
	defaultTimeoutDuration = 30 * time.Second
)

// TimeoutKillAfter is option of `RunCommand()` set waiting limit to `kill -kill` after terminating the command.
var TimeoutKillAfter = 10 * time.Second

var cmdBase = []string{"sh", "-c"}

// CommandOption carries a timeout duration.
type CommandOption struct {
	TimeoutDuration time.Duration
}

func init() {
	if runtime.GOOS == "windows" {
		cmdBase = []string{"cmd", "/c"}
	}
}

// RunCommand runs command (in two string) and returns stdout, stderr strings and its exit code.
func RunCommand(command, user string, env []string, opt CommandOption) (stdout, stderr string, exitCode int, err error) {
	cmdArgs := append(cmdBase, command)
	return RunCommandArgs(cmdArgs, user, env, opt)
}

// RunCommandArgs run the command
func RunCommandArgs(cmdArgs []string, user string, env []string, opt CommandOption) (stdout, stderr string, exitCode int, err error) {
	args := append([]string{}, cmdArgs...)
	if user != "" {
		if runtime.GOOS == "windows" {
			logger.Warningf("RunCommand ignore option: user = %q", user)
		} else {
			args = append([]string{"sudo", "-Eu", user}, args...)
		}
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = append(os.Environ(), env...)
	tio := &timeout.Timeout{
		Cmd:       cmd,
		Duration:  defaultTimeoutDuration,
		KillAfter: TimeoutKillAfter,
	}
	if opt.TimeoutDuration != 0 {
		tio.Duration = opt.TimeoutDuration
	}
	exitStatus, stdout, stderr, err := tio.Run()

	if err == nil && exitStatus.IsTimedOut() {
		err = fmt.Errorf("command timed out")
	}
	if err != nil {
		logger.Errorf("RunCommand error command: %v, error: %s", cmdArgs, err.Error())
	}
	return stdout, stderr, exitStatus.GetChildExitCode(), err
}
