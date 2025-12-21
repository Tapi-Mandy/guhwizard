// FILE: internal/engine/runner.go
package engine

import (
	"fmt"
	"guhwizard/internal/config"
	"guhwizard/internal/installer"
)

type ProgressMsg struct {
	CurrentPercent float64
	CurrentStep    string
}

type Runner struct {
	Config       *config.Config
	LogChan      chan string
	ProgressChan chan ProgressMsg
}

func NewRunner(cfg *config.Config, logChan chan string, progChan chan ProgressMsg) *Runner {
	return &Runner{
		Config:       cfg,
		LogChan:      logChan,
		ProgressChan: progChan,
	}
}

func (r *Runner) Log(msg string) {
	if r.LogChan != nil {
		r.LogChan <- msg
	}
}

func (r *Runner) reportProgress(pct float64, step string) {
	if r.ProgressChan != nil {
		r.ProgressChan <- ProgressMsg{CurrentPercent: pct, CurrentStep: step}
	}
}

func (r *Runner) Install(password string) error {
	r.reportProgress(0.0, "Authenticating...")

	// 1. Start Sudo KeepAlive
	if err := installer.CurrentSession.StartSudoKeepAlive(password); err != nil {
		return fmt.Errorf("authentication failed: %v", err)
	}
	defer installer.CurrentSession.StopSudo()

	// 2. Install AUR Helper
	r.reportProgress(0.1, fmt.Sprintf("Installing AUR Helper (%s)...", r.Config.Settings.AURHelper))
	if err := installer.InstallAURHelper(r.Config, r.Log); err != nil {
		return err
	}

	// 3. Install Packages (Base + Selected)
	r.reportProgress(0.2, "Installing Packages...")
	if err := installer.InstallPackages(r.Config, r.Log); err != nil {
		return err
	}

	// 4. External Scripts
	r.reportProgress(0.5, "Running Setup Scripts...")
	if err := installer.RunExternalScripts(r.Config, r.Log); err != nil {
		return err
	}

	// 5. System Configuration (SDDM, Shell, Terminal)
	r.reportProgress(0.6, "Configuring System...")

	// Check for SDDM
	for _, step := range r.Config.Steps {
		if step.ID == "dm" {
			for _, item := range step.Items {
				if item.Name == "sddm" && item.Selected {
					r.reportProgress(0.65, "Configuring SDDM...")
					if err := installer.ConfigureSDDM(r.Config, r.Log); err != nil {
						r.Log(fmt.Sprintf("Error configuring SDDM: %v", err))
					}
				}
			}
		}
	}

	// Check for Terminal Patch
	for _, step := range r.Config.Steps {
		if step.ID == "terminals" {
			for _, item := range step.Items {
				if item.Selected {
					if err := installer.PatchTerminal(item.Name, r.Log); err != nil {
						r.Log(fmt.Sprintf("Error patching terminal: %v", err))
					}
				}
			}
		}
	}

	// Check for Shell
	for _, step := range r.Config.Steps {
		if step.ID == "shell" {
			for _, item := range step.Items {
				if item.Selected && item.Name != "bash" { // bash is default usually
					if err := installer.ChangeShell(item.Name, r.Log); err != nil {
						r.Log(fmt.Sprintf("Error changing shell: %v", err))
					}
				}
			}
		}
	}

	// 6. Dotfiles
	r.reportProgress(0.8, "Installing Dotfiles...")
	if err := installer.ProcessDotfiles(r.Config, r.Log); err != nil {
		return err
	}

	r.reportProgress(1.0, "Installation Complete!")
	return nil
}
