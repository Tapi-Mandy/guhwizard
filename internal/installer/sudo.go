// FILE: internal/installer/sudo.go
package installer

import (
	"bufio"
	"os/exec"
	"sync"
)

// Session manages the sudo session (now mostly a placeholder for passwordless)
type Session struct {
	active bool
	mu     sync.Mutex
}

var CurrentSession = &Session{}

// ValidateSudo checks if sudo is working without a password using `sudo -n true`
func ValidateSudo() error {
	cmd := exec.Command("sudo", "-n", "true")
	return cmd.Run()
}

func (s *Session) StartSudoKeepAlive(pwd string) error {
	// No-op now as we use timestamp_timeout=-1
	return nil
}

func (s *Session) StopSudo() {
	// No-op
}

// RunSudo executes a command with sudo privileges.
// It assumes passwordless sudo is configured in /etc/sudoers.d/
func RunSudo(onLog func(string), command string, args ...string) error {
	cmd := exec.Command("sudo", append([]string{command}, args...)...)

	pipe, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(pipe)
	go func() {
		for scanner.Scan() {
			onLog(scanner.Text() + "\n")
		}
	}()

	return cmd.Wait()
}
