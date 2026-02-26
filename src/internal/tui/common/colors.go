package common

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Catppuccin Mocha color palette.
var (
	ColorBg          = lipgloss.Color("#1e1e2e")
	ColorBgSurface   = lipgloss.Color("#242435")
	ColorBgOverlay   = lipgloss.Color("#2a2a3c")
	ColorBgHighlight = lipgloss.Color("#313147")
	ColorBgSelected  = lipgloss.Color("#3e3e5e")
	ColorBorder      = lipgloss.Color("#45456a")
	ColorBorderDim   = lipgloss.Color("#363654")
	ColorBorderFocus = lipgloss.Color("#cba6f7")
	ColorText        = lipgloss.Color("#cdd6f4")
	ColorTextSub     = lipgloss.Color("#a6adc8")
	ColorTextDim     = lipgloss.Color("#6c7086")
	ColorTextMuted   = lipgloss.Color("#45475a")
	ColorPink        = lipgloss.Color("#f5c2e7")
	ColorMauve       = lipgloss.Color("#cba6f7")
	ColorPeach       = lipgloss.Color("#fab387")
	ColorGreen       = lipgloss.Color("#a6e3a1")
	ColorTeal        = lipgloss.Color("#94e2d5")
	ColorBlue        = lipgloss.Color("#89b4fa")
	ColorYellow      = lipgloss.Color("#f9e2af")
	ColorRed         = lipgloss.Color("#f38ba8")
	ColorLavender    = lipgloss.Color("#b4befe")
	ColorSky         = lipgloss.Color("#89dceb")
	ColorFlamingo    = lipgloss.Color("#f2cdcd")
	ColorRosewater   = lipgloss.Color("#f5e0dc")
)

// LanguageColor returns the theme color for a given language.
func LanguageColor(lang string) color.Color {
	switch lang {
	case "go":
		return ColorSky
	case "nix":
		return ColorMauve
	case "sql":
		return ColorYellow
	case "bash":
		return ColorGreen
	case "python":
		return ColorBlue
	case "typescript", "javascript":
		return ColorBlue
	case "rust":
		return ColorPeach
	case "lua":
		return ColorLavender
	case "ruby":
		return ColorRed
	case "java":
		return ColorPeach
	case "c", "cpp":
		return ColorBlue
	case "markdown":
		return ColorTextSub
	case "toml", "yaml", "json":
		return ColorTextDim
	default:
		return ColorTextSub
	}
}
