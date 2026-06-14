package components

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Starlexxx/lazy-k8s/internal/k8s"
	"github.com/Starlexxx/lazy-k8s/internal/ui/theme"
)

type LogLineMsg struct {
	Line  string
	Error error
}

type LogViewer struct {
	styles    *theme.Styles
	lines     []string
	offset    int
	width     int
	height    int
	follow    bool
	pod       string
	namespace string
	container string
	// title overrides "Logs: <pod>" when streaming from multiple pods
	// (e.g. "Logs: Deployment/my-app (3 pods)"). Empty in single-pod mode.
	title        string
	cancel       context.CancelFunc
	maxLines     int
	searchActive bool
	searchInput  textinput.Model
	searchQuery  string
	matchLines   []int
	matchIndex   int
}

func NewLogViewer(styles *theme.Styles) *LogViewer {
	ti := textinput.New()
	ti.Placeholder = "search..."
	ti.CharLimit = 100
	ti.Width = 30
	ti.Prompt = "/ "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(styles.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(styles.Text)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(styles.MutedColor)

	return &LogViewer{
		styles:      styles,
		lines:       make([]string, 0),
		follow:      true,
		maxLines:    10000,
		searchInput: ti,
		matchLines:  make([]int, 0),
	}
}

func (l *LogViewer) Start(client *k8s.Client, namespace, pod, container string) tea.Cmd {
	l.pod = pod
	l.namespace = namespace
	l.container = container
	l.title = ""
	l.lines = make([]string, 0)
	l.offset = 0

	ctx, cancel := context.WithCancel(context.Background())
	l.cancel = cancel

	return func() tea.Msg {
		logChan, err := client.StreamPodLogs(ctx, namespace, pod, k8s.LogOptions{
			Container: container,
			Follow:    true,
			TailLines: 100,
		})
		if err != nil {
			return LogLineMsg{Error: err}
		}

		go func() {
			for logLine := range logChan {
				if logLine.Error != nil {
					// Send error but don't stop
					continue
				}
			}
		}()

		// Initial load - get snapshot
		logs, err := client.GetPodLogSnapshot(ctx, namespace, pod, k8s.LogOptions{
			Container: container,
			TailLines: 100,
		})
		if err != nil {
			return LogLineMsg{Error: err}
		}

		return LogLineMsg{Line: logs}
	}
}

// StartMulti tails logs from every pod matching selector, prefixing each line
// with the pod name so multiplexed output is readable. The title reflects the
// owning workload ("Logs: Deployment/my-app (3 pods)") rather than a pod name.
//
// If no pods match the selector, returns a status line instead of silent
// emptiness — a label typo is the common reason and empty UI hides it.
func (l *LogViewer) StartMulti(
	client *k8s.Client,
	namespace, workloadKind, workloadName, selector string,
) tea.Cmd {
	l.pod = ""
	l.namespace = namespace
	l.container = ""
	l.title = "Logs: " + workloadKind + "/" + workloadName
	l.lines = make([]string, 0)
	l.offset = 0

	ctx, cancel := context.WithCancel(context.Background())
	l.cancel = cancel

	return func() tea.Msg {
		pods, err := client.ListPodsBySelector(ctx, namespace, selector)
		if err != nil {
			return LogLineMsg{Error: err}
		}

		if len(pods) == 0 {
			return LogLineMsg{Line: "No pods match selector: " + selector + "\n"}
		}

		names := make([]string, 0, len(pods))
		for _, p := range pods {
			names = append(names, p.Name)
		}

		// Update the header count now that we know how many pods matched.
		l.title = "Logs: " + workloadKind + "/" + workloadName +
			" (" + podCountLabel(len(names)) + ")"

		logs, err := client.GetPodsLogSnapshot(ctx, namespace, names, k8s.LogOptions{
			TailLines: 100,
		})
		if err != nil {
			return LogLineMsg{Error: err}
		}

		return LogLineMsg{Line: logs}
	}
}

func podCountLabel(n int) string {
	if n == 1 {
		return "1 pod"
	}

	return strconv.Itoa(n) + " pods"
}

func (l *LogViewer) Stop() {
	if l.cancel != nil {
		l.cancel()
		l.cancel = nil
	}
}

func (l *LogViewer) Update(msg tea.Msg) (*LogViewer, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if l.searchActive {
			switch msg.String() {
			case "esc":
				l.searchActive = false
				l.searchInput.Blur()

				return l, nil
			case "enter":
				l.searchActive = false
				l.searchInput.Blur()
				l.performSearch()

				if len(l.matchLines) > 0 {
					l.matchIndex = 0
					l.offset = l.matchLines[0]
					l.follow = false
				}

				return l, nil
			default:
				var cmd tea.Cmd

				l.searchInput, cmd = l.searchInput.Update(msg)
				l.searchQuery = l.searchInput.Value()

				return l, cmd
			}
		}

		switch msg.String() {
		case "/":
			l.searchActive = true
			l.searchInput.Focus()
			l.searchInput.SetValue("")

			return l, nil
		case "n":
			if len(l.matchLines) > 0 {
				l.matchIndex = (l.matchIndex + 1) % len(l.matchLines)
				l.offset = l.matchLines[l.matchIndex]
				l.follow = false
			}
		case "N":
			if len(l.matchLines) > 0 {
				l.matchIndex--
				if l.matchIndex < 0 {
					l.matchIndex = len(l.matchLines) - 1
				}

				l.offset = l.matchLines[l.matchIndex]
				l.follow = false
			}
		case "up", "k":
			if l.offset > 0 {
				l.offset--
				l.follow = false
			}
		case "down", "j":
			maxOffset := max(len(l.lines)-l.height+6, 0)

			if l.offset < maxOffset {
				l.offset++
			}
			// Re-enable follow if at bottom
			if l.offset >= maxOffset {
				l.follow = true
			}
		case "g":
			l.offset = 0
			l.follow = false
		case "G":
			maxOffset := max(len(l.lines)-l.height+6, 0)

			l.offset = maxOffset
			l.follow = true
		case "f":
			l.follow = !l.follow
			if l.follow {
				maxOffset := max(len(l.lines)-l.height+6, 0)

				l.offset = maxOffset
			}
		case "pgup", "ctrl+u":
			l.offset -= l.height / 2
			if l.offset < 0 {
				l.offset = 0
			}

			l.follow = false
		case "pgdown", "ctrl+d":
			maxOffset := max(len(l.lines)-l.height+6, 0)

			l.offset += l.height / 2
			if l.offset > maxOffset {
				l.offset = maxOffset
			}

			if l.offset >= maxOffset {
				l.follow = true
			}
		}

	case LogLineMsg:
		if msg.Error != nil {
			l.lines = append(l.lines, "Error: "+msg.Error.Error())
		} else if msg.Line != "" {
			newLines := strings.Split(msg.Line, "\n")
			l.lines = append(l.lines, newLines...)

			if len(l.lines) > l.maxLines {
				l.lines = l.lines[len(l.lines)-l.maxLines:]
			}

			if l.follow {
				maxOffset := max(len(l.lines)-l.height+6, 0)

				l.offset = maxOffset
			}
		}

		return l, l.tickCmd()
	}

	return l, nil
}

func (l *LogViewer) tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		// This just keeps the UI responsive
		// Real log streaming would happen in goroutine
		return nil
	})
}

func (l *LogViewer) performSearch() {
	l.matchLines = make([]int, 0)

	if l.searchQuery == "" {
		return
	}

	query := strings.ToLower(l.searchQuery)

	for i, line := range l.lines {
		if strings.Contains(strings.ToLower(line), query) {
			l.matchLines = append(l.matchLines, i)
		}
	}
}

func (l *LogViewer) isCurrentMatch(lineIdx int) bool {
	if len(l.matchLines) == 0 || l.matchIndex >= len(l.matchLines) {
		return false
	}

	return l.matchLines[l.matchIndex] == lineIdx
}

func (l *LogViewer) View(width, height int) string {
	l.width = width
	l.height = height

	var b strings.Builder

	titleText := l.title
	if titleText == "" {
		titleText = "Logs: " + l.pod
	}

	title := l.styles.ModalTitle.Render(titleText)

	followIndicator := ""
	if l.follow {
		followIndicator = l.styles.StatusSuccess.Render(" [FOLLOW]")
	}

	var hint string
	if l.searchActive {
		hint = l.styles.Muted.Render("enter search • esc cancel")
	} else {
		hint = l.styles.Muted.Render("/ search • n/N next/prev • f follow • esc close")
	}

	titleBar := lipgloss.JoinHorizontal(lipgloss.Center, title, followIndicator, "  ", hint)
	b.WriteString(titleBar)
	b.WriteString("\n")

	if l.searchActive {
		l.searchInput.Width = width - 10
		b.WriteString(l.searchInput.View())
		b.WriteString("\n")
	} else if l.searchQuery != "" && len(l.matchLines) > 0 {
		matchInfo := lipgloss.NewStyle().Foreground(l.styles.Primary).Render(
			" [" + l.searchQuery + "] " +
				string(rune('0'+l.matchIndex+1)) + "/" +
				string(rune('0'+len(l.matchLines))) + " matches",
		)
		b.WriteString(matchInfo)
		b.WriteString("\n")
	}

	b.WriteString(strings.Repeat("─", width-4))
	b.WriteString("\n")

	visibleHeight := height - 7
	if l.searchActive || (l.searchQuery != "" && len(l.matchLines) > 0) {
		visibleHeight--
	}

	if visibleHeight < 1 {
		visibleHeight = 1
	}

	endIdx := min(l.offset+visibleHeight, len(l.lines))

	startIdx := max(l.offset, 0)

	for i := startIdx; i < endIdx; i++ {
		line := l.lines[i]
		if len(line) > width-6 {
			line = line[:width-9] + "..."
		}

		highlighted := l.highlightLogLine(line)

		searchMatch := l.searchQuery != "" &&
			strings.Contains(strings.ToLower(l.lines[i]), strings.ToLower(l.searchQuery))

		if searchMatch {
			if l.isCurrentMatch(i) {
				highlighted = l.styles.ListItemFocused.Render("► " + highlighted)
			} else {
				highlighted = lipgloss.NewStyle().
					Background(lipgloss.Color("#3d4966")).
					Render(highlighted)
			}
		}

		b.WriteString(highlighted)
		b.WriteString("\n")
	}

	// Fill remaining space
	for i := endIdx - startIdx; i < visibleHeight; i++ {
		b.WriteString("\n")
	}

	if width > 40 {
		separatorWidth := max(width-30, 0)

		lineInfo := l.styles.Muted.Render(
			lipgloss.NewStyle().Align(lipgloss.Right).Width(width - 8).
				Render(strings.Repeat("─", separatorWidth) + " " +
					l.styles.StatusValue.Render(
						strings.TrimSpace(strings.Repeat(" ", 10)+
							"Lines: "+string(rune('0'+len(l.lines)%10))),
					)),
		)
		b.WriteString(lineInfo)
	}

	return l.styles.Modal.
		Width(width - 4).
		Height(height - 2).
		Render(b.String())
}

func (l *LogViewer) highlightLogLine(line string) string {
	timestampStyle := l.styles.Muted
	errorStyle := lipgloss.NewStyle().Foreground(l.styles.Error)
	warnStyle := lipgloss.NewStyle().Foreground(l.styles.Warning)
	infoStyle := lipgloss.NewStyle().Foreground(l.styles.Text)

	lower := strings.ToLower(line)

	if strings.Contains(lower, "error") || strings.Contains(lower, "fatal") ||
		strings.Contains(lower, "panic") {
		return errorStyle.Render(line)
	}

	if strings.Contains(lower, "warn") {
		return warnStyle.Render(line)
	}

	// Check for timestamp at start (common formats)
	if len(line) > 20 {
		// ISO format: 2024-01-15T10:30:00
		if line[4] == '-' && line[7] == '-' && line[10] == 'T' {
			return timestampStyle.Render(line[:23]) + " " + infoStyle.Render(line[24:])
		}
	}

	return infoStyle.Render(line)
}
