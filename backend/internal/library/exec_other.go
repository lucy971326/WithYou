//go:build !windows

package library

import "os/exec"

func hideWindow(cmd *exec.Cmd) {}
