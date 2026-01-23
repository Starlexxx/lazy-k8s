package components

import (
	"context"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lazyk8s/lazy-k8s/internal/k8s"
	"github.com/lazyk8s/lazy-k8s/internal/ui/theme"
)

type LogLineMsg struct {
	Line  string
	Error error
}

type LogViewer struct {
	styles     *theme.Styles
	lines      []string
	offset     int
	width      int
	height     int
	follow     bool
	pod        string
	namespace  string
	container  string
	cancel     context.CancelFunc
	maxLines   int
}

func NewLogViewer(styles *theme.Styles) *LogViewer {
	return &LogViewer{
		styles:   styles,
		lines:    make([]string, 0),
		follow:   true,
		maxLines: 10000,
	}
}

func (l *LogViewer) Start(client *k8s.Client, namespace, pod, container string) tea.Cmd {
	l.pod = pod
	l.namespace = namespace
	l.container = container
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
				// Use tea.Program.Send would be ideal, but we return a cmd instead
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

func (l *LogViewer) Stop() {
	if l.cancel != nil {
		l.cancel()
		l.cancel = nil
	}
}

func (l *LogViewer) Update(msg tea.Msg) (*LogViewer, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if l.offset > 0 {
				l.offset--
				l.follow = false
			}
		case "down", "j":
			maxOffset := len(l.lines) - l.height + 6
			if maxOffset < 0 {
				maxOffset = 0
			}
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
			maxOffset := len(l.lines) - l.height + 6
			if maxOffset < 0 {
				maxOffset = 0
			}
			l.offset = maxOffset
			l.follow = true
		case "f":
			l.follow = !l.follow
			if l.follow {
				maxOffset := len(l.lines) - l.height + 6
				if maxOffset < 0 {
					maxOffset = 0
				}
				l.offset = maxOffset
			}
		case "pgup", "ctrl+u":
			l.offset -= l.height / 2
			if l.offset < 0 {
				l.offset = 0
			}
			l.follow = false
		case "pgdown", "ctrl+d":
			maxOffset := len(l.lines) - l.height + 6
			if maxOffset < 0 {
				maxOffset = 0
			}
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

			// Trim if too many lines
			if len(l.lines) > l.maxLines {
				l.lines = l.lines[len(l.lines)-l.maxLines:]
			}

			// Auto-scroll if following
			if l.follow {
				maxOffset := len(l.lines) - l.height + 6
				if maxOffset < 0 {
					maxOffset = 0
				}
				l.offset = maxOffset
			}
		}

		// Continue reading logs
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

func (l *LogViewer) View(width, height int) string {
	l.width = width
	l.height = height

	var b strings.Builder

	// Title bar
	title := l.styles.ModalTitle.Render("Logs: " + l.pod)

	followIndicator := ""
	if l.follow {
		followIndicator = l.styles.StatusSuccess.Render(" [FOLLOW]")
	}

	hint := l.styles.Muted.Render("↑/↓ scroll • f follow • g/G top/bottom • esc close")
	titleBar := lipgloss.JoinHorizontal(lipgloss.Center, title, followIndicator, "  ", hint)
	b.WriteString(titleBar)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", width-4))
	b.WriteString("\n")

	// Content area
	visibleHeight := height - 6
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	endIdx := l.offset + visibleHeight
	if endIdx > len(l.lines) {
		endIdx = len(l.lines)
	}

	startIdx := l.offset
	if startIdx < 0 {
		startIdx = 0
	}

	for i := startIdx; i < endIdx; i++ {
		line := l.lines[i]
		if len(line) > width-6 {
			line = line[:width-9] + "..."
		}
		// Basic log highlighting
		highlighted := l.highlightLogLine(line)
		b.WriteString(highlighted)
		b.WriteString("\n")
	}

	// Fill remaining space
	for i := endIdx - startIdx; i < visibleHeight; i++ {
		b.WriteString("\n")
	}

	// Line count
	lineInfo := l.styles.Muted.Render(
		lipgloss.NewStyle().Align(lipgloss.Right).Width(width - 8).
			Render(strings.Repeat("─", width-30) + " " +
				l.styles.StatusValue.Render(
					strings.TrimSpace(strings.Repeat(" ", 10)+
						"Lines: "+string(rune('0'+len(l.lines)%10))),
				)),
	)
	b.WriteString(lineInfo)

	return l.styles.Modal.
		Width(width - 4).
		Height(height - 2).
		Render(b.String())
}

func (l *LogViewer) highlightLogLine(line string) string {
	// Highlight timestamps
	timestampStyle := l.styles.Muted
	errorStyle := lipgloss.NewStyle().Foreground(l.styles.Error)
	warnStyle := lipgloss.NewStyle().Foreground(l.styles.Warning)
	infoStyle := lipgloss.NewStyle().Foreground(l.styles.Text)

	lower := strings.ToLower(line)

	// Check for log levels
	if strings.Contains(lower, "error") || strings.Contains(lower, "fatal") || strings.Contains(lower, "panic") {
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
