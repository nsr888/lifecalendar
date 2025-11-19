package colors

import (
	"fmt"
	"hash/fnv"

	"github.com/charmbracelet/lipgloss"
)

func MonthDay() lipgloss.Style {
	return lipgloss.NewStyle().
		Width(2).
		Align(lipgloss.Right).
		Bold(false)
}

func Text() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#909090"))
}

func PrintText(text string) {
	textStyle := Text()
	fmt.Println(textStyle.Render(text))
}

func Header() lipgloss.Style {
	return lipgloss.NewStyle().
		MarginTop(1).
		MarginBottom(1).
		Foreground(lipgloss.Color("#4d4d4d"))
}

func PrintHeader(text string) {
	headerStyle := Header()
	fmt.Println(headerStyle.Render(text))
}

// PickBGStyle maps category color names to background lipgloss styles
func PickBGStyle(colorCode string) lipgloss.Style {
	baseStyle := lipgloss.NewStyle().
		Width(2).
		Align(lipgloss.Right).
		Bold(false)

	return baseStyle.Background(lipgloss.Color(colorCode))
}

// GenerateColor creates a dark color from a label hash
func GenerateColor(label string) string {
	h := fnv.New64a()
	h.Write([]byte(label))
	hashVal := int64(h.Sum64())

	r := (hashVal & 0xFF0000) >> 16
	g := (hashVal & 0x00FF00) >> 8
	b := hashVal & 0x0000FF

	// Darken colors
	rd := int(r / 2)
	gd := int(g / 2)
	bd := int(b / 2)

	// Clamp to max 20
	if rd > 20 {
		rd = 20
	}
	if gd > 20 {
		gd = 20
	}
	if bd > 20 {
		bd = 20
	}
	// Ensure non-negative
	if rd < 0 {
		rd = 0
	}
	if gd < 0 {
		gd = 0
	}
	if bd < 0 {
		bd = 0
	}

	return fmt.Sprintf("#%02x%02x%02x", rd, gd, bd)
}
