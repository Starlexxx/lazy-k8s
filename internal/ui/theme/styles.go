package theme

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/lazyk8s/lazy-k8s/internal/config"
)

type Styles struct {
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Error      lipgloss.Color
	Warning    lipgloss.Color
	Background lipgloss.Color
	Text       lipgloss.Color
	Border     lipgloss.Color
	MutedColor lipgloss.Color
	Success    lipgloss.Color

	Muted lipgloss.Style

	App lipgloss.Style

	Header          lipgloss.Style
	HeaderTitle     lipgloss.Style
	HeaderContext   lipgloss.Style
	HeaderNamespace lipgloss.Style
	HeaderHelp      lipgloss.Style

	Panel            lipgloss.Style
	PanelFocused     lipgloss.Style
	PanelTitle       lipgloss.Style
	PanelTitleActive lipgloss.Style

	ListItem         lipgloss.Style
	ListItemSelected lipgloss.Style
	ListItemFocused  lipgloss.Style

	StatusBar     lipgloss.Style
	StatusKey     lipgloss.Style
	StatusValue   lipgloss.Style
	StatusError   lipgloss.Style
	StatusSuccess lipgloss.Style
	StatusWarning lipgloss.Style

	TableHeader lipgloss.Style
	TableRow    lipgloss.Style
	TableCell   lipgloss.Style

	DetailTitle lipgloss.Style
	DetailLabel lipgloss.Style
	DetailValue lipgloss.Style

	Modal       lipgloss.Style
	ModalTitle  lipgloss.Style
	ModalButton lipgloss.Style

	Input       lipgloss.Style
	InputPrompt lipgloss.Style

	StatusRunning     lipgloss.Style
	StatusPending     lipgloss.Style
	StatusFailed      lipgloss.Style
	StatusSucceeded   lipgloss.Style
	StatusUnknown     lipgloss.Style
	StatusTerminating lipgloss.Style
}

func NewStyles(cfg *config.ThemeConfig) *Styles {
	s := &Styles{
		Primary:    lipgloss.Color(cfg.PrimaryColor),
		Secondary:  lipgloss.Color(cfg.SecondaryColor),
		Error:      lipgloss.Color(cfg.ErrorColor),
		Warning:    lipgloss.Color(cfg.WarningColor),
		Background: lipgloss.Color(cfg.BackgroundColor),
		Text:       lipgloss.Color(cfg.TextColor),
		Border:     lipgloss.Color(cfg.BorderColor),
		MutedColor: lipgloss.Color("#565f89"),
		Success:    lipgloss.Color(cfg.SecondaryColor),
	}

	s.Muted = lipgloss.NewStyle().
		Foreground(s.MutedColor)

	s.App = lipgloss.NewStyle().
		Background(s.Background)

	s.Header = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(s.Border).
		Padding(0, 1)

	s.HeaderTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(s.Primary)

	s.HeaderContext = lipgloss.NewStyle().
		Foreground(s.Secondary)

	s.HeaderNamespace = lipgloss.NewStyle().
		Foreground(s.Warning)

	s.HeaderHelp = lipgloss.NewStyle().
		Foreground(s.MutedColor)

	s.Panel = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(s.Border).
		Padding(0, 1)

	s.PanelFocused = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(s.Primary).
		Padding(0, 1)

	s.PanelTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(s.MutedColor).
		Padding(0, 1)

	s.PanelTitleActive = lipgloss.NewStyle().
		Bold(true).
		Foreground(s.Primary).
		Padding(0, 1)

	s.ListItem = lipgloss.NewStyle().
		Foreground(s.Text)

	s.ListItemSelected = lipgloss.NewStyle().
		Foreground(s.Primary).
		Bold(true)

	s.ListItemFocused = lipgloss.NewStyle().
		Foreground(s.Background).
		Background(s.Primary)

	s.StatusBar = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(s.Border).
		Padding(0, 1)

	s.StatusKey = lipgloss.NewStyle().
		Foreground(s.Primary).
		Bold(true)

	s.StatusValue = lipgloss.NewStyle().
		Foreground(s.Text)

	s.StatusError = lipgloss.NewStyle().
		Foreground(s.Error)

	s.StatusSuccess = lipgloss.NewStyle().
		Foreground(s.Success)

	s.StatusWarning = lipgloss.NewStyle().
		Foreground(s.Warning)

	s.TableHeader = lipgloss.NewStyle().
		Bold(true).
		Foreground(s.Primary).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(s.Border)

	s.TableRow = lipgloss.NewStyle().
		Foreground(s.Text)

	s.TableCell = lipgloss.NewStyle().
		Padding(0, 1)

	s.DetailTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(s.Primary).
		MarginBottom(1)

	s.DetailLabel = lipgloss.NewStyle().
		Foreground(s.MutedColor).
		Width(15)

	s.DetailValue = lipgloss.NewStyle().
		Foreground(s.Text)

	s.Modal = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(s.Primary).
		Padding(1, 2)

	s.ModalTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(s.Primary).
		MarginBottom(1)

	s.ModalButton = lipgloss.NewStyle().
		Padding(0, 2).
		MarginRight(1)

	s.Input = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(s.Border).
		Padding(0, 1)

	s.InputPrompt = lipgloss.NewStyle().
		Foreground(s.Primary)

	s.StatusRunning = lipgloss.NewStyle().
		Foreground(s.Success)

	s.StatusPending = lipgloss.NewStyle().
		Foreground(s.Warning)

	s.StatusFailed = lipgloss.NewStyle().
		Foreground(s.Error)

	s.StatusSucceeded = lipgloss.NewStyle().
		Foreground(s.Success)

	s.StatusUnknown = lipgloss.NewStyle().
		Foreground(s.MutedColor)

	s.StatusTerminating = lipgloss.NewStyle().
		Foreground(s.Warning)

	return s
}

func (s *Styles) GetStatusStyle(status string) lipgloss.Style {
	switch status {
	case "Running", "Active", "Ready", "Bound":
		return s.StatusRunning
	case "Pending", "ContainerCreating", "PodInitializing":
		return s.StatusPending
	case "Failed", "Error", "CrashLoopBackOff", "ImagePullBackOff", "ErrImagePull":
		return s.StatusFailed
	case "Succeeded", "Completed":
		return s.StatusSucceeded
	case "Terminating":
		return s.StatusTerminating
	default:
		return s.StatusUnknown
	}
}
