package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const (
	ColorNvidiaGreen = "#76B900"
	ColorNeonGreen   = "#39FF14"
	ColorDeepBlack   = "#0A0A0A"
	ColorDarkGray    = "#1A1A2E"
	ColorMidGray     = "#2D2D44"
	ColorLightGray   = "#8892B0"
	ColorWhite       = "#E6F1FF"
	ColorCyan        = "#64FFDA"
	ColorYellow      = "#FFD700"
	ColorRed         = "#FF4444"
	ColorPurple      = "#BD93F9"
	ColorOrange      = "#FFB86C"
)

var (
	BannerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorNvidiaGreen)).
			Bold(true)

	AppNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorNeonGreen)).
			Bold(true)

	ChatBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorMidGray)).
			Padding(0, 1)

	UserMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorWhite))

	UserPrefixStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorCyan)).
			Bold(true)

	AIMessageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWhite))

	AIPrefixStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorNvidiaGreen)).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorRed)).
			Bold(true).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color(ColorRed)).
			PaddingLeft(1)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorNvidiaGreen)).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorYellow)).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorLightGray)).
			Italic(true)

	ModelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorOrange)).
			Bold(true)

	TimestampStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPurple)).
			Italic(true)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorLightGray)).
			Background(lipgloss.Color(ColorDarkGray)).
			Padding(0, 1)

	CommandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorCyan)).
			Bold(true)

	DividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorMidGray))

	PromptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorNeonGreen)).
			Bold(true)

	SectionTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorCyan)).
				Bold(true).
				Underline(true)
)

func Banner() string {
	logo := `
 ███╗   ██╗██╗   ██╗██╗██████╗ ██╗ █████╗     ███╗   ██╗██╗███╗   ███╗
 ████╗  ██║██║   ██║██║██╔══██╗██║██╔══██╗    ████╗  ██║██║████╗ ████║
 ██╔██╗ ██║██║   ██║██║██║  ██║██║███████║    ██╔██╗ ██║██║██╔████╔██║
 ██║╚██╗██║╚██╗ ██╔╝██║██║  ██║██║██╔══██║    ██║╚██╗██║██║██║╚██╔╝██║
 ██║ ╚████║ ╚████╔╝ ██║██████╔╝██║██║  ██║    ██║ ╚████║██║██║ ╚═╝ ██║
 ╚═╝  ╚═══╝  ╚═══╝  ╚═╝╚═════╝ ╚═╝╚═╝  ╚═╝    ╚═╝  ╚═══╝╚═╝╚═╝     ╚═╝`

	subtitle := "Interface CLI pour NVIDIA AI Foundation Models (NIM)"
	version := "  v1.0.0 │ Alimenté par Llama-3, Mistral & plus"

	return BannerStyle.Render(logo) + "\n" +
		AppNameStyle.Render(subtitle) + "\n" +
		InfoStyle.Render(version) + "\n"
}

func SmallBanner(model string) string {
	left := AIPrefixStyle.Render("NVIDIA NIM")
	mid := InfoStyle.Render(" │ ")
	right := ModelStyle.Render(model)
	sep := DividerStyle.Render(strings.Repeat("─", 60))
	return left + mid + right + "\n" + sep
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Spinner struct {
	message  string
	frame    int
	running  bool
	stopChan chan bool
}

func NewSpinner(message string) *Spinner {
	return &Spinner{
		message:  message,
		stopChan: make(chan bool),
	}
}

func (s *Spinner) Start() {
	s.running = true
	go func() {
		for {
			select {
			case <-s.stopChan:
				fmt.Printf("\r%s\r", strings.Repeat(" ", len(s.message)+10))
				return
			default:
				spinnerStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color(ColorNvidiaGreen)).
					Bold(true)
				frame := spinnerStyle.Render(spinnerFrames[s.frame%len(spinnerFrames)])
				msgStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color(ColorLightGray))
				fmt.Printf("\r%s %s", frame, msgStyle.Render(s.message))
				s.frame++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

func (s *Spinner) Stop(successMsg string) {
	if s.running {
		s.stopChan <- true
		s.running = false
	}
	if successMsg != "" {
		fmt.Println(SuccessStyle.Render("✓ " + successMsg))
	}
}

func (s *Spinner) StopWithError(errMsg string) {
	if s.running {
		s.stopChan <- true
		s.running = false
	}
	if errMsg != "" {
		fmt.Println(ErrorStyle.Render("✗ " + errMsg))
	}
}

func PrintError(msg string) {
	fmt.Println(ErrorStyle.Render("✗ Erreur : " + msg))
}

func PrintSuccess(msg string) {
	fmt.Println(SuccessStyle.Render("✓ " + msg))
}

func PrintWarning(msg string) {
	fmt.Println(WarningStyle.Render("⚠ " + msg))
}

func PrintInfo(msg string) {
	fmt.Println(InfoStyle.Render("ℹ " + msg))
}

func Divider(width int) string {
	if width <= 0 {
		width = 60
	}
	return DividerStyle.Render(strings.Repeat("─", width))
}

func FormatUserMessage(content string) string {
	return UserPrefixStyle.Render("Vous ▶ ") + UserMessageStyle.Render(content)
}

func FormatAIMessage(modelName, content string) string {
	return AIPrefixStyle.Render(""+modelName+" ▶ ") + "\n" + AIMessageStyle.Render(content)
}

func FormatTimestamp(t time.Time) string {
	return TimestampStyle.Render(t.Format("15:04:05"))
}

func BoxMessage(title, content string) string {
	return ChatBoxStyle.Render(SectionTitleStyle.Render(title) + "\n\n" + content)
}
