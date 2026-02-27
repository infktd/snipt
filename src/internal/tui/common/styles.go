package common

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// RenderBadgePill renders the SNIPT badge as a solid mauve background pill
// with dark text, matching the mockup's gradient background badge.
// In terminal: solid mauve bg since we can't do CSS gradients.
func RenderBadgePill(text string) string {
	style := lipgloss.NewStyle().
		Foreground(ColorBg).
		Background(ColorMauve).
		Bold(true).
		Padding(0, 1) // 1 char padding left and right for pill shape
	return style.Render(text)
}

// RenderLangBadge renders a language name as a colored pill with a subtle
// tinted background. Language-colored text on surface0 for contrast.
func RenderLangBadge(lang string, outerBg color.Color) string {
	langColor := LanguageColor(lang)

	badgeStyle := lipgloss.NewStyle().
		Foreground(langColor).
		Background(lipgloss.Color("#363a4f")).
		Bold(true)

	return badgeStyle.Render(" " + lang + " ")
}

// RenderTagBadge renders a tag as a flat pill with subtle tinted background,
// matching the language badge style.
func RenderTagBadge(tag string, fg color.Color, bg color.Color) string {
	badgeStyle := lipgloss.NewStyle().
		Foreground(fg).
		Background(lipgloss.Color("#363a4f"))

	return badgeStyle.Render(" #" + tag + " ")
}

// RenderFuzzyTitle renders the title with matched characters highlighted.
func RenderFuzzyTitle(title string, indices []int, selected bool) string {
	baseColor := ColorTextSub
	if selected {
		baseColor = ColorText
	}

	if len(indices) == 0 {
		style := lipgloss.NewStyle().Foreground(baseColor)
		return style.Render(title)
	}

	matchSet := make(map[int]bool, len(indices))
	for _, idx := range indices {
		matchSet[idx] = true
	}

	matchStyle := lipgloss.NewStyle().Foreground(ColorPink).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(baseColor)

	var out strings.Builder
	for i, ch := range title {
		if matchSet[i] {
			out.WriteString(matchStyle.Render(string(ch)))
		} else {
			out.WriteString(normalStyle.Render(string(ch)))
		}
	}

	return out.String()
}

// RenderFuzzyTitleWithBg renders the title with matched characters highlighted
// and a specific background color applied to every character.
func RenderFuzzyTitleWithBg(title string, indices []int, selected bool, bg color.Color) string {
	baseColor := ColorTextSub
	if selected {
		baseColor = ColorText
	}

	if len(indices) == 0 {
		style := lipgloss.NewStyle().Foreground(baseColor).Background(bg)
		return style.Render(title)
	}

	matchSet := make(map[int]bool, len(indices))
	for _, idx := range indices {
		matchSet[idx] = true
	}

	matchStyle := lipgloss.NewStyle().Foreground(ColorPink).Bold(true).Background(bg)
	normalStyle := lipgloss.NewStyle().Foreground(baseColor).Background(bg)

	var out strings.Builder
	for i, ch := range title {
		if matchSet[i] {
			out.WriteString(matchStyle.Render(string(ch)))
		} else {
			out.WriteString(normalStyle.Render(string(ch)))
		}
	}

	return out.String()
}
