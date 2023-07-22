//go:build windows
// +build windows

package internal

import (
	"log"
	"os/exec"
	"strings"
	"syscall"
)

// Executes the given command
func Exec(c string, args ...string) string {
	log.Printf("Exec: %v %v", c, args)
	cmd := exec.Command(c, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	out, _ := cmd.Output()

	if len(out) == 0 {
		return ""
	}
	s := string(out)
	return strings.ReplaceAll(s, "\n", "")
}
