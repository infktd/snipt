package common

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Catppuccin Macchiato color palette.
var (
	ColorBg          = lipgloss.Color("#24273a")
	ColorBgSurface   = lipgloss.Color("#363a4f")
	ColorBgOverlay   = lipgloss.Color("#494d64")
	ColorBgHighlight = lipgloss.Color("#363a4f")
	ColorBgSelected  = lipgloss.Color("#494d64")
	ColorBorder      = lipgloss.Color("#494d64")
	ColorBorderDim   = lipgloss.Color("#363a4f")
	ColorBorderFocus = lipgloss.Color("#c6a0f6")
	ColorText        = lipgloss.Color("#cad3f5")
	ColorTextSub     = lipgloss.Color("#b8c0e0")
	ColorTextDim     = lipgloss.Color("#8087a2")
	ColorTextMuted   = lipgloss.Color("#5b6078")
	ColorPink        = lipgloss.Color("#f5bde6")
	ColorMauve       = lipgloss.Color("#c6a0f6")
	ColorPeach       = lipgloss.Color("#f5a97f")
	ColorGreen       = lipgloss.Color("#a6da95")
	ColorTeal        = lipgloss.Color("#8bd5ca")
	ColorBlue        = lipgloss.Color("#8aadf4")
	ColorYellow      = lipgloss.Color("#eed49f")
	ColorRed         = lipgloss.Color("#ed8796")
	ColorLavender    = lipgloss.Color("#b7bdf8")
	ColorSky         = lipgloss.Color("#91d7e3")
	ColorFlamingo    = lipgloss.Color("#f0c6c6")
	ColorRosewater   = lipgloss.Color("#f4dbd6")
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
