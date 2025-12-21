// FILE: internal/installer/system_config.go
package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"guhwizard/internal/config"
	"guhwizard/internal/fs"
)

func ConfigureSDDM(cfg *config.Config, log func(string)) error {
	log("Installing SDDM Theme dependencies...\n")
	deps := []string{"qt6-svg", "qt6-virtualkeyboard", "qt6-multimedia-ffmpeg"}

	// Install deps
	// We append the deps to the args list safely
	pacArgs := []string{"-S", "--needed", "--noconfirm"}
	pacArgs = append(pacArgs, deps...)
	if err := RunSudo(log, "pacman", pacArgs...); err != nil {
		return fmt.Errorf("failed to install sddm deps: %w", err)
	}

	home, _ := os.UserHomeDir()
	tempDir := filepath.Join(home, "Downloads", "SilentSDDM_Setup")
	os.RemoveAll(tempDir)

	log("Cloning SilentSDDM theme...\n")
	if err := exec.Command("git", "clone", "https://github.com/uiriansan/SilentSDDM", tempDir).Run(); err != nil {
		return fmt.Errorf("failed to clone theme repo: %w", err)
	}

	log("Installing Theme Files...\n")
	RunSudo(log, "mkdir", "-p", "/usr/share/sddm/themes/silent")
	RunSudo(log, "sh", "-c", fmt.Sprintf("cp -rf %s/. /usr/share/sddm/themes/silent/", tempDir))

	log("Installing Fonts...\n")
	RunSudo(log, "mkdir", "-p", "/usr/share/fonts")
	RunSudo(log, "sh", "-c", fmt.Sprintf("cp -r %s/fonts/* /usr/share/fonts/", tempDir))

	log("Patching /etc/sddm.conf...\n")
	// Safe Backup manually via sudo since it's root owned
	RunSudo(log, "cp", "-n", "/etc/sddm.conf", "/etc/sddm.conf.bkp")

	configBlock := `[Theme]
Current=silent

[General]
InputMethod=qtvirtualkeyboard
GreeterEnvironment=QML2_IMPORT_PATH=/usr/share/sddm/themes/silent/components/,QT_IM_MODULE=qtvirtualkeyboard
`
	// Write temp file then sudo move it
	tmpConfig := filepath.Join(os.TempDir(), "sddm_patch.conf")
	if err := os.WriteFile(tmpConfig, []byte(configBlock), 0644); err != nil {
		return err
	}

	log("Writing SDDM config...\n")
	if err := RunSudo(log, "sh", "-c", fmt.Sprintf("cat %s | tee /etc/sddm.conf", tmpConfig)); err != nil {
		return fmt.Errorf("failed to write sddm config: %w", err)
	}

	log("Enabling SDDM service...\n")
	return RunSudo(log, "systemctl", "enable", "sddm")
}

func PatchTerminal(selectedTerminal string, log func(string)) error {
	log(fmt.Sprintf("Patching default terminal to %s...\n", selectedTerminal))

	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".config/mangowc/config.conf")

	// Expand path just in case
	configPath, _ = fs.ExpandHome(configPath)

	input, err := os.ReadFile(configPath)
	if err != nil {
		log(fmt.Sprintf("Warning: Config file %s not found, skipping patch.\n", configPath))
		return nil // Not fatal
	}

	// Naive replace
	target := "bind=ALT, Return, spawn, foot"
	replacement := fmt.Sprintf("bind=ALT, Return, spawn, %s", selectedTerminal)
	output := strings.Replace(string(input), target, replacement, 1)

	// Use SafeFS for atomic write
	return fs.AtomicWrite(configPath, []byte(output), 0644)
}

func ChangeShell(shellName string, log func(string)) error {
	log(fmt.Sprintf("Changing shell to %s...\n", shellName))

	// Get path
	out, err := exec.Command("which", shellName).Output()
	if err != nil {
		return fmt.Errorf("shell %s not found", shellName)
	}
	shellPath := strings.TrimSpace(string(out))

	user := os.Getenv("USER")
	return RunSudo(log, "chsh", "-s", shellPath, user)
}
