//go:build !windows
// +build !windows

package internal

import (
	"log"
	"os/exec"
	"strings"
)

// Executes the given command
func Exec(c string, args ...string) string {
	log.Printf("Exec: %v %v", c, args)
	cmd := exec.Command(c, args...)

	out, _ := cmd.Output()
	if len(out) == 0 {
		return ""
	}
	s := string(out)
	return strings.ReplaceAll(s, "\n", "")
}
