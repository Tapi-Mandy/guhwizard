// FILE: internal/installer/dotfiles.go
package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"guhwizard/internal/config"
	"guhwizard/internal/fs"
)

func ProcessDotfiles(cfg *config.Config, log func(string)) error {
	log("Processing dotfiles...\n")

	repo := cfg.Settings.Dotfiles.Repo
	if repo == "" {
		log("No dotfiles repo configured. Skipping.\n")
		return nil
	}

	home, _ := os.UserHomeDir()
	tempDir := filepath.Join(home, "guhwm-temp")

	// Clean previous run
	os.RemoveAll(tempDir)

	log(fmt.Sprintf("Cloning %s...\n", repo))
	if err := exec.Command("git", "clone", repo, tempDir).Run(); err != nil {
		return fmt.Errorf("failed to clone dotfiles: %w", err)
	}
	defer os.RemoveAll(tempDir) // Cleanup

	// Process items
	for _, item := range cfg.Settings.Dotfiles.Items {
		// Fix: Don't expand home for the temp dir part, only verify structure
		fullSrc := filepath.Join(tempDir, item.Src)

		destPath, err := fs.ExpandHome(item.Dest)
		if err != nil {
			return err
		}

		log(fmt.Sprintf("Installing configs to %s...\n", destPath))

		// We need to copy contents of src to dest
		// If src is a dir, we copy recursively?
		// The original logic was `cp -r src/. dest/`
		// We'll walk the source directory structure

		err = filepath.Walk(fullSrc, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Get relative path from source root
			relPath, err := filepath.Rel(fullSrc, path)
			if err != nil {
				return err
			}

			if relPath == "." {
				return nil
			}

			targetPath := filepath.Join(destPath, relPath)

			if info.IsDir() {
				return os.MkdirAll(targetPath, 0755)
			}

			log(fmt.Sprintf("  -> %s\n", relPath))
			return fs.BackupAndCopy(path, targetPath)
		})

		if err != nil {
			return fmt.Errorf("failed to copy configs: %w", err)
		}
	}

	return nil
}

func RunExternalScripts(cfg *config.Config, log func(string)) error {
	for _, script := range cfg.Settings.ExternalScripts {
		log(fmt.Sprintf("Running script: %s\n", script.Name))

		cmd := exec.Command("bash", "-c", script.Command)
		out, err := cmd.CombinedOutput()
		log(string(out))
		if err != nil {
			return fmt.Errorf("script %s failed: %w", script.Name, err)
		}
	}
	return nil
}
