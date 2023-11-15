package main

import (
	"fmt"
	"os"
	"os/exec"
)

type (
	nullReader struct{}
	nullWriter struct{}
)

func (nullReader) Read(p []byte) (int, error) {
	return len(p), nil
}

func (nullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	cmd := exec.Command(command, args...)
	cmd.Stdin = nullReader{}
	cmd.Stderr = nullWriter{}
	cmd.Stdout = nullWriter{}

	if err := cmd.Run(); err != nil {
		fmt.Printf("Err: %v", err)
		os.Exit(cmd.ProcessState.ExitCode())
	}
}
