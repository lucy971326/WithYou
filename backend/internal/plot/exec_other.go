//go:build !windows

package plot

import "os/exec"

func hideWindow(cmd *exec.Cmd) {}
