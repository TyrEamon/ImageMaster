//go:build !windows

package jmbridge

import "os/exec"

func hideConsoleWindow(cmd *exec.Cmd) {}
