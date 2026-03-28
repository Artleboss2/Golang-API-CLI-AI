package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	menuSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorNvidiaGreen)).
				Bold(true)

	menuCheckedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorCyan))

	menuNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorLightGray))

	menuTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorCyan)).
			Bold(true).
			Underline(true)

	menuHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorMidGray)).
			Italic(true)
)

type MenuItem struct {
	ID    string
	Label string
	Desc  string
}

type MultiSelectModel struct {
	title    string
	items    []MenuItem
	cursor   int
	selected map[int]bool
	done     bool
	escaped  bool
}

func NewMultiSelect(title string, items []MenuItem) MultiSelectModel {
	return MultiSelectModel{
		title:    title,
		items:    items,
		selected: make(map[int]bool),
	}
}

func (m MultiSelectModel) Init() tea.Cmd {
	return nil
}

func (m MultiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.escaped = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case " ":
			m.selected[m.cursor] = !m.selected[m.cursor]
		case "enter":
			m.done = true
			return m, tea.Quit
		case "a":
			if len(m.selected) == len(m.items) {
				m.selected = make(map[int]bool)
			} else {
				for i := range m.items {
					m.selected[i] = true
				}
			}
		}
	}
	return m, nil
}

func (m MultiSelectModel) View() string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(menuTitleStyle.Render("  " + m.title))
	sb.WriteString("\n")
	sb.WriteString(menuHintStyle.Render("  ↑↓ naviguer • espace sélectionner • a tout • entrée confirmer"))
	sb.WriteString("\n\n")

	for i, item := range m.items {
		checked := "○"
		nameStyle := menuNormalStyle
		if m.selected[i] {
			checked = "●"
			nameStyle = menuCheckedStyle
		}

		cursor := "  "
		if m.cursor == i {
			cursor = menuSelectedStyle.Render("▶ ")
			nameStyle = menuSelectedStyle
		}

		line := fmt.Sprintf("%s%s %s", cursor, checked, nameStyle.Render(item.Label))
		if item.Desc != "" {
			line += "  " + menuHintStyle.Render(item.Desc)
		}
		sb.WriteString(line + "\n")
	}

	sb.WriteString("\n")
	return sb.String()
}

func (m MultiSelectModel) Selected() []string {
	var result []string
	for i, item := range m.items {
		if m.selected[i] {
			result = append(result, item.ID)
		}
	}
	return result
}

func (m MultiSelectModel) Escaped() bool {
	return m.escaped
}

type CommandMenuItem struct {
	Key   string
	Label string
	Desc  string
}

type CommandMenuModel struct {
	items   []CommandMenuItem
	cursor  int
	chosen  string
	escaped bool
}

func NewCommandMenu(items []CommandMenuItem) CommandMenuModel {
	return CommandMenuModel{items: items}
}

func (m CommandMenuModel) Init() tea.Cmd {
	return nil
}

func (m CommandMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.escaped = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			m.chosen = m.items[m.cursor].Key
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m CommandMenuModel) View() string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(menuTitleStyle.Render("  Commandes disponibles"))
	sb.WriteString("\n")
	sb.WriteString(menuHintStyle.Render("  ↑↓ naviguer • entrée exécuter • esc annuler"))
	sb.WriteString("\n\n")

	for i, item := range m.items {
		cursor := "  "
		keyStyle := menuNormalStyle
		descStyle := menuHintStyle
		if m.cursor == i {
			cursor = menuSelectedStyle.Render("▶ ")
			keyStyle = menuSelectedStyle
			descStyle = menuNormalStyle
		}
		sb.WriteString(fmt.Sprintf("%s%s  %s\n",
			cursor,
			keyStyle.Render(fmt.Sprintf("%-20s", item.Label)),
			descStyle.Render(item.Desc),
		))
	}
	sb.WriteString("\n")
	return sb.String()
}

func (m CommandMenuModel) Chosen() string { return m.chosen }
func (m CommandMenuModel) Escaped() bool  { return m.escaped }

func RunMultiSelect(title string, items []MenuItem) ([]string, bool) {
	m := NewMultiSelect(title, items)
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return nil, true
	}
	final := result.(MultiSelectModel)
	if final.Escaped() {
		return nil, true
	}
	return final.Selected(), false
}

func RunCommandMenu(items []CommandMenuItem) (string, bool) {
	m := NewCommandMenu(items)
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return "", true
	}
	final := result.(CommandMenuModel)
	if final.Escaped() {
		return "", true
	}
	return final.Chosen(), false
}
