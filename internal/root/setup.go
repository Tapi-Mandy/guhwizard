// FILE: internal/root/setup.go
package root

import (
	"fmt"
	"os"
	"os/exec"
)

const SudoersFile = "/etc/sudoers.d/99-no-password-until-reboot"
const SudoersContent = "Defaults timestamp_timeout=-1\n"

// ConfigureSudoTimestamp creates a temporary sudoers file to allow passwordless sudo until reboot.
// It validates the file using visudo before claiming success.
func ConfigureSudoTimestamp() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("this mode must be run as root")
	}

	fmt.Println("Configuring passwordless sudo until reboot...")

	// 1. Write the content
	if err := os.WriteFile(SudoersFile, []byte(SudoersContent), 0440); err != nil {
		return fmt.Errorf("failed to write sudoers file: %w", err)
	}

	// 2. Set permissions (must be 0440)
	if err := os.Chmod(SudoersFile, 0440); err != nil {
		os.Remove(SudoersFile) // Cleanup
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// 3. Validate with visudo
	// -c: check-only, -f: file path
	cmd := exec.Command("visudo", "-cf", SudoersFile)
	if out, err := cmd.CombinedOutput(); err != nil {
		os.Remove(SudoersFile) // CRITICAL: Cleanup invalid file
		return fmt.Errorf("visudo validation failed: %s", string(out))
	}

	fmt.Println("Success: Sudo timestamp timeout set to -1.")
	return nil
}
