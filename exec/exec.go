package exec

import (
	"bytes"
	"os/exec"
)

const gitCmd = "git"

type GitRunnerI interface {
	ExecGitCommand(
		args []string,
		stdoutb *bytes.Buffer,
		erroutb *bytes.Buffer,
		dir string,
	) (cmd *exec.Cmd, err error)
}

type ShellGitRunner struct{}

// ExecGitCommand executes git with flags passed as `args` and can change working directory if `dir` is passed
func (repo *ShellGitRunner) ExecGitCommand(
	args []string,
	stdoutb *bytes.Buffer,
	erroutb *bytes.Buffer,
	dir string,
) (cmd *exec.Cmd, err error) {
	cmd = exec.Command(gitCmd, args...)

	if stdoutb != nil {
		cmd.Stdout = stdoutb
	}
	if erroutb != nil {
		cmd.Stderr = erroutb
	}

	if dir != "" {
		cmd.Dir = dir
	}

	err = cmd.Run()
	return cmd, err
}
