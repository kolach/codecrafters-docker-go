package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/kolach/docker/hub"
)

func parseImgAndVersion() (string, string) {
	imgAndVersion := strings.SplitN(os.Args[2], ":", 2)
	img := imgAndVersion[0]
	ver := "latest"
	if len(imgAndVersion) > 1 {
		ver = imgAndVersion[1]
	}
	return img, ver
}

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	ctx := context.Background()

	// Create root of executable command
	tempDir, err := os.MkdirTemp("", "mychroot")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	img, ver := parseImgAndVersion()

	// pull image into tempDir
	if err := hub.PullImage(ctx, img, ver, tempDir); err != nil {
		panic(err)
	}

	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	// Enter the chroot.
	if err := syscall.Chroot(tempDir); err != nil {
		panic(err)
	}

	cmd := exec.Command(command, args...)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID,
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf("Err: %v", err)
		os.Exit(cmd.ProcessState.ExitCode())
	}
}
