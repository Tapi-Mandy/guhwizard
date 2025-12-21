// FILE: internal/installer/packages.go
package installer

import (
	"bufio"
	"fmt"
	"guhwizard/internal/config"
	"os"
	"os/exec"
	"path/filepath"
)

// InstallAURHelper installs the configured AUR helper (yay/paru).
func InstallAURHelper(cfg *config.Config, log func(string)) error {
	helper := cfg.Settings.AURHelper
	log(fmt.Sprintf("Checking for %s...", helper))

	if _, err := exec.LookPath(helper); err == nil {
		log("Already installed.\n")
		return nil
	}

	log("Installing git and base-devel...\n")
	if err := RunSudo(log, "pacman", "-S", "--needed", "--noconfirm", "git", "base-devel"); err != nil {
		return fmt.Errorf("failed to install base-devel: %v", err)
	}

	home, _ := os.UserHomeDir()
	buildDir := filepath.Join(home, "Downloads", helper)
	os.RemoveAll(buildDir)

	log(fmt.Sprintf("Cloning %s...", helper))
	if err := exec.Command("git", "clone", fmt.Sprintf("https://aur.archlinux.org/%s.git", helper), buildDir).Run(); err != nil {
		return fmt.Errorf("failed to clone %s: %v", helper, err)
	}

	log("Building package...")
	// makepkg -si without sudo prompt is tricky.
	// We build with makepkg, then install with pacman -U using our RunSudo
	buildCmd := exec.Command("makepkg", "-sfc", "--noconfirm")
	buildCmd.Dir = buildDir
	// makepkg might ask for sudo for deps if they aren't installed.
	// The safest way is to ensure all deps are met or run makepkg such that it doesn't need root immediately?
	// Actually, makepkg -si ASKS for sudo.
	// We will trust the user to have base-devel installed (done above).
	// The issue is if makepkg needs to install deps.
	// We can try to rely on current session being cached, OR we accept that makepkg might fail if it needs root and can't get it.
	// Enhanced approach: Use 'makepkg' (no install), then find the .pkg.tar.zst and install with RunSudo.

	if out, err := buildCmd.CombinedOutput(); err != nil {
		log(string(out))
		return fmt.Errorf("build failed: %v", err)
	}

	matches, _ := filepath.Glob(filepath.Join(buildDir, "*.pkg.tar.zst"))
	if len(matches) == 0 {
		return fmt.Errorf("no package found in %s", buildDir)
	}

	log("Installing built package...\n")
	return RunSudo(log, "pacman", "-U", "--noconfirm", matches[0])
}

// InstallPackages installs all selected packages from the config
func InstallPackages(cfg *config.Config, log func(string)) error {
	helper := cfg.Settings.AURHelper

	// Collect all deps
	deps := make([]string, 0)
	deps = append(deps, cfg.Settings.BasePackages...)

	// Add selected items from config
	// Note: The UI Model should have populated a list of selected packages.
	// However, the current architecture passed `cfg` around.
	// The `UserConfig` struct in tasks.go was doing some transformations.
	// We need to adhere to the config.Config structure.

	// In the new architecture, we should iterate over Steps in Config, check .Selected on Items.
	for _, step := range cfg.Steps {
		for _, item := range step.Items {
			if item.Selected {
				deps = append(deps, item.Name)
			}
		}
	}

	if len(deps) == 0 {
		log("No packages to install.\n")
		return nil
	}

	log(fmt.Sprintf("Installing %d packages using %s...\n", len(deps), helper))

	// Batch install
	args := append([]string{"-S", "--noconfirm", "--needed"}, deps...)

	// Aur helpers usually don't need sudo for the fetch/build part, but need it for install.
	// yay/paru handle sudo internally.
	// Since we don't have a PTY to pass to yay, we rely on the sudo cache we refreshed?
	// NO. yay/paru will prompt if they need absolute root and can't find it.
	// BUT we have a problem: RunSudo runs `sudo cmd`. We cannot run `sudo yay`.
	// We should treat yay as a user command that MIGHT ask for sudo.
	// Since we are running `sudo -v` in the background (KeepAlive), sudo should remain cached.

	cmd := exec.Command(helper, args...)
	pipe, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		log(scanner.Text() + "\n")
	}

	return cmd.Wait()
}
