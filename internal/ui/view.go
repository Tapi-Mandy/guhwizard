// FILE: internal/ui/view.go
package ui

import (
	"guhwizard/internal/styles"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	var content string
	header := styles.Highlight.Render(styles.Logo)

	switch m.state {
	case StateWelcome:
		content = lipgloss.JoinVertical(lipgloss.Center,
			header,
			"\nWelcome to the GuhWizard Engine",
			styles.Subtle.Render("Press Enter to Start Configuration"),
		)

	case StateSelection:
		// Requested text: "Press space to select, Enter to skip this step"
		// Logic suggests "skip" is relevant if nothing is selected, but user asked for this as general instruction.
		// However, "Enter to skip" implies moving next without selection.
		// "Next Step" is more accurate if items ARE selected.
		// But I will stick to the requested text format or make it dynamic.

		hasSelection := false
		for _, itm := range m.list.Items() {
			if itm.(listItem).configItem.Selected {
				hasSelection = true
				break
			}
		}

		var footerText string
		if hasSelection {
			footerText = "Press [Space] to select, [Enter] for Next Step"
		} else {
			footerText = "Press [Space] to select, [Enter] to skip this step"
		}

		content = lipgloss.JoinVertical(lipgloss.Left,
			header,
			m.list.View(),
			styles.Subtle.Render("\n"+footerText),
		)

	case StateConfirmation:
		// Generate summary
		var summary string
		summary += styles.Highlight.Render("Summary of Changes:") + "\n\n"

		summary += "• " + m.cfg.Settings.AURHelper + " (AUR Helper)\n"

		for _, step := range m.cfg.Steps {
			for _, item := range step.Items {
				if item.Selected {
					summary += "• " + item.Name + "\n"
				}
			}
		}

		summary += "\n" + styles.Subtle.Render("Press [Enter] to Confirm or [Ctrl+C] to Cancel")

		content = lipgloss.JoinVertical(lipgloss.Center,
			header,
			summary,
		)

	case StatePassword:
		content = lipgloss.JoinVertical(lipgloss.Center,
			header,
			"Configuration Complete.",
			"Authentication required to proceed with installation.",
			"\n",
			m.textInput.View(),
		)

	case StateInstalling:
		var mainArea string
		if m.showLogs {
			// CRITICAL FIX: Removed the inner styles.Container wrapper
			// Now logs render cleanly without a "box in a box"
			mainArea = lipgloss.JoinVertical(lipgloss.Left,
				styles.Highlight.Render("Installation Logs ('V' to hide):"),
				m.viewport.View(),
			)
		} else {
			mainArea = lipgloss.JoinVertical(lipgloss.Center,
				m.statusMsg,
				"\n",
				m.progress.View(),
				"\n",
				styles.Subtle.Render("(Press 'V' to view verbose logs)"),
			)
		}
		content = lipgloss.JoinVertical(lipgloss.Center, header, mainArea)

	case StateDone:
		content = lipgloss.JoinVertical(lipgloss.Center,
			header,
			styles.Success.Render("Installation Complete!"),
			"Press Enter to Exit",
		)
	}

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		styles.Container.Render(content),
	)
}
