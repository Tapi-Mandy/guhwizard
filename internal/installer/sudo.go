// FILE: internal/installer/sudo.go
package installer

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Session manages the sudo session in memory
type Session struct {
	Password string
	active   bool
	mu       sync.Mutex
	stopChan chan struct{}
}

var CurrentSession = &Session{}

// ValidateSudo checks if the password is correct by running `sudo -S -v`
func ValidateSudo(pwd string) error {
	cmd := exec.Command("sudo", "-S", "-v", "-k") // -k invalidates cache first
	cmd.Stdin = strings.NewReader(pwd + "\n")
	return cmd.Run()
}

// StartSudoKeepAlive starts a background routine to keep the sudo timestamp alive.
func (s *Session) StartSudoKeepAlive(pwd string) error {
	if err := ValidateSudo(pwd); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.active {
		return nil // Already running
	}

	s.Password = pwd
	s.active = true
	s.stopChan = make(chan struct{})

	go func() {
		ticker := time.NewTicker(4 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Refresh credentials
				cmd := exec.Command("sudo", "-S", "-v")
				cmd.Stdin = strings.NewReader(s.Password + "\n")
				_ = cmd.Run()
			case <-s.stopChan:
				return
			}
		}
	}()

	return nil
}

func (s *Session) StopSudo() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active {
		close(s.stopChan)
		s.active = false
		s.Password = ""                  // Clear memory
		exec.Command("sudo", "-k").Run() // Invalidate credentials
	}
}

// RunSudo executes a command with sudo privileges using the stored session password.
// It uses `sudo -S` to read the password from stdin.
func RunSudo(onLog func(string), command string, args ...string) error {
	CurrentSession.mu.Lock()
	pwd := CurrentSession.Password
	CurrentSession.mu.Unlock()

	if pwd == "" {
		return fmt.Errorf("sudo session not active")
	}

	// construct command
	cmdArgs := append([]string{"-S", command}, args...)
	cmd := exec.Command("sudo", cmdArgs...)

	// Pipe password to stdin
	cmd.Stdin = strings.NewReader(pwd + "\n")

	// Capture stdout/stderr
	// We cannot just pipe stdout because sudo -S might prompt (though -v should prevent that)
	// We'll capture combined output for logging.

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
