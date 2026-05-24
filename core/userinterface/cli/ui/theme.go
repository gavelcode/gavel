package ui

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	Gold    = lipgloss.Color("#D4A017")
	Ivory   = lipgloss.Color("#FFFDD0")
	Crimson = lipgloss.Color("#8B0000")
	Emerald = lipgloss.Color("#2E8B57")
	Slate   = lipgloss.Color("#708090")
)

var (
	Title   = lipgloss.NewStyle().Bold(true).Foreground(Gold)
	Phase   = lipgloss.NewStyle().Bold(true).Foreground(Gold)
	Success = lipgloss.NewStyle().Foreground(Emerald).Bold(true)
	Error   = lipgloss.NewStyle().Foreground(Crimson).Bold(true)
	Dim     = lipgloss.NewStyle().Foreground(Slate)
	Label   = lipgloss.NewStyle().Foreground(Ivory)
	GoldBar = lipgloss.NewStyle().Foreground(Gold)
)

const bannerHorizontalPadding = 2

var Banner = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(Gold).
	Padding(0, bannerHorizontalPadding)

func ThemeGavel() *huh.Theme {
	theme := huh.ThemeBase()

	theme.Focused.Base = theme.Focused.Base.BorderForeground(Gold)
	theme.Focused.Card = theme.Focused.Base
	theme.Focused.Title = theme.Focused.Title.Foreground(Gold).Bold(true)
	theme.Focused.NoteTitle = theme.Focused.NoteTitle.Foreground(Gold).Bold(true).MarginBottom(1)
	theme.Focused.Description = theme.Focused.Description.Foreground(Slate)
	theme.Focused.ErrorIndicator = theme.Focused.ErrorIndicator.Foreground(Crimson)
	theme.Focused.ErrorMessage = theme.Focused.ErrorMessage.Foreground(Crimson)
	theme.Focused.SelectSelector = theme.Focused.SelectSelector.Foreground(Gold)
	theme.Focused.NextIndicator = theme.Focused.NextIndicator.Foreground(Gold)
	theme.Focused.PrevIndicator = theme.Focused.PrevIndicator.Foreground(Gold)
	theme.Focused.Option = theme.Focused.Option.Foreground(Ivory)
	theme.Focused.MultiSelectSelector = theme.Focused.MultiSelectSelector.Foreground(Gold)
	theme.Focused.SelectedOption = theme.Focused.SelectedOption.Foreground(Emerald)
	theme.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(Emerald).SetString("✓ ")
	theme.Focused.UnselectedPrefix = lipgloss.NewStyle().Foreground(Slate).SetString("• ")
	theme.Focused.UnselectedOption = theme.Focused.UnselectedOption.Foreground(Ivory)
	theme.Focused.FocusedButton = theme.Focused.FocusedButton.Foreground(lipgloss.Color("#000")).Background(Gold)
	theme.Focused.Next = theme.Focused.FocusedButton
	theme.Focused.BlurredButton = theme.Focused.BlurredButton.Foreground(Ivory).Background(lipgloss.Color("237"))
	theme.Focused.TextInput.Cursor = theme.Focused.TextInput.Cursor.Foreground(Emerald)
	theme.Focused.TextInput.Placeholder = theme.Focused.TextInput.Placeholder.Foreground(Slate)
	theme.Focused.TextInput.Prompt = theme.Focused.TextInput.Prompt.Foreground(Gold)

	theme.Blurred = theme.Focused
	theme.Blurred.Base = theme.Focused.Base.BorderStyle(lipgloss.HiddenBorder())
	theme.Blurred.Card = theme.Blurred.Base
	theme.Blurred.NextIndicator = lipgloss.NewStyle()
	theme.Blurred.PrevIndicator = lipgloss.NewStyle()

	theme.Group.Title = theme.Focused.Title
	theme.Group.Description = theme.Focused.Description

	return theme
}
