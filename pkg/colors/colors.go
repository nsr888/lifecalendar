package colors

import (
	"fmt"

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
