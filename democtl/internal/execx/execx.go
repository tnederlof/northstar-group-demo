package execx

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

// RunOpts configures how a command should be run
type RunOpts struct {
	Dir    string            // working directory
	Env    map[string]string // additional environment variables (merged with os.Environ())
	Stdin  io.Reader         // standard input
	Stdout io.Writer         // standard output (defaults to os.Stdout)
	Stderr io.Writer         // standard error (defaults to os.Stderr)
}

// Run executes a command with the given options
func Run(name string, args []string, opts RunOpts) error {
	cmd := exec.Command(name, args...)

	if opts.Dir != "" {
		cmd.Dir = opts.Dir
	}

	// Start with current environment
	cmd.Env = os.Environ()

	// Add additional environment variables
	for k, v := range opts.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Set up I/O
	cmd.Stdin = opts.Stdin

	if opts.Stdout != nil {
		cmd.Stdout = opts.Stdout
	} else {
		cmd.Stdout = os.Stdout
	}

	if opts.Stderr != nil {
		cmd.Stderr = opts.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("command failed with exit code %d", exitErr.ExitCode())
		}
		return fmt.Errorf("failed to run command: %w", err)
	}

	return nil
}

// RunScript runs a bash script with the given arguments
func RunScript(scriptPath string, args []string, opts RunOpts) error {
	bashArgs := append([]string{scriptPath}, args...)
	return Run("bash", bashArgs, opts)
}
