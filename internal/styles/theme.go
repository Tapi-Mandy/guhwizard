// FILE: internal/styles/theme.go
package styles

import "github.com/charmbracelet/lipgloss"

// Palette (Catppuccin Mocha inspired)
const (
    ColorRosewater = "#f5e0dc"
    ColorFlamingo  = "#f2cdcd"
    ColorMauve     = "#cba6f7"
    ColorRed       = "#f38ba8"
    ColorGreen     = "#a6e3a1"
    ColorBase      = "#1e1e2e"
    ColorText      = "#cdd6f4"
    ColorSubtext   = "#a6adc8"
    ColorSurface   = "#313244" // Darker background for selected item
)

var (
    // General Text Styles
    Subtle    = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSubtext))
    Highlight = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorMauve)).Bold(true)
    Error     = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRed))
    Success   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen))
    
    // Logo Style
    LogoStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color(ColorMauve)).
        MarginBottom(1)

    // Container
    Container = lipgloss.NewStyle().
        Padding(1, 2).
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color(ColorMauve))

    // --- NEW: List Item Styles (Matching your screenshot) ---
    
    // The "Normal" item (Unselected)
    ItemNormalTitle = lipgloss.NewStyle().
        Foreground(lipgloss.Color(ColorText)).
        PaddingLeft(2) // Space for the missing border
    
    ItemNormalDesc = lipgloss.NewStyle().
        Foreground(lipgloss.Color(ColorSubtext)).
        PaddingLeft(2)

    // The "Selected" item
    // We use a thick left border with the Mauve color
    ItemSelectedTitle = lipgloss.NewStyle().
        Foreground(lipgloss.Color(ColorMauve)).
        Bold(true).
        Border(lipgloss.ThickBorder(), false, false, false, true). // Left Border Only
        BorderForeground(lipgloss.Color(ColorMauve)).
        PaddingLeft(1)

    ItemSelectedDesc = lipgloss.NewStyle().
        Foreground(lipgloss.Color(ColorFlamingo)).
        Border(lipgloss.ThickBorder(), false, false, false, true). // Left Border Only
        BorderForeground(lipgloss.Color(ColorMauve)).
        PaddingLeft(1)
)

const Logo = `
  ________      .__             .__                         .___
 /  _____/ __ __|  |____  _  __ |__|_____________ _______ __| _/
/   \  ___|  |  \  |  \ \/ \/ / |  \___   /\__  \\_  __ \/ __ | 
\    \_\  \  |  /   Y  \     /  |  |/    /  / __ \|  | \/ /_/ | 
 \______  /____/|___|  /\/\_/   |__/_____ \(____  /__|  \____ | 
        \/           \/                    \/     \/           \/ 
`
