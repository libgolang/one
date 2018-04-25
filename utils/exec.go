package utils

import (
	"io"
	"os"
	"os/exec"
)

// Exec executes and passes output to stdout, returns error on non zero exit
func Exec(args ...string) error {

	// cmd
	cmd := exec.Command(args[0], args[1:]...)

	// env
	cmd.Env = os.Environ()

	// Out
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	go func() {
		_, _ = io.Copy(os.Stdout, stdout)
	}()

	// Err
	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}
	go func() {
		_, _ = io.Copy(os.Stderr, stderr)
	}()

	// start
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	// wati
	return cmd.Wait()
}

// ExecSilent executes a command without passing the output to stdin and stderr
func ExecSilent(args ...string) error {

	// cmd
	cmd := exec.Command(args[0], args[1:]...)

	// env
	cmd.Env = os.Environ()

	// Out
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	go func() {
		buf := make([]byte, 256)
		for _, err := stdout.Read(buf); err != nil; _, err = stdout.Read(buf) {
		}
		//_, _ = io.Copy(os.Stdout, stdout)
	}()

	// Err
	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}
	go func() {
		buf := make([]byte, 256)
		for _, err := stderr.Read(buf); err != nil; _, err = stderr.Read(buf) {
		}
	}()

	// start
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	// wati
	return cmd.Wait()
}
