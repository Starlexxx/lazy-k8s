package panels

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/lazyk8s/lazy-k8s/internal/k8s"
	"github.com/lazyk8s/lazy-k8s/internal/ui/theme"
	"github.com/lazyk8s/lazy-k8s/internal/utils"
)

type EventsPanel struct {
	BasePanel
	client   *k8s.Client
	styles   *theme.Styles
	events   []corev1.Event
	filtered []corev1.Event
}

func NewEventsPanel(client *k8s.Client, styles *theme.Styles) *EventsPanel {
	return &EventsPanel{
		BasePanel: BasePanel{
			title:       "Events",
			shortcutKey: "8",
		},
		client: client,
		styles: styles,
	}
}

func (p *EventsPanel) Init() tea.Cmd {
	return p.Refresh()
}

func (p *EventsPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
			p.MoveUp()
		case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
			p.MoveDown(len(p.filtered))
		case key.Matches(msg, key.NewBinding(key.WithKeys("g"))):
			p.MoveToTop()
		case key.Matches(msg, key.NewBinding(key.WithKeys("G"))):
			p.MoveToBottom(len(p.filtered))
		}

	case eventsLoadedMsg:
		p.events = msg.events
		// Sort by last timestamp (most recent first)
		sort.Slice(p.events, func(i, j int) bool {
			ti := p.events[i].LastTimestamp.Time

			tj := p.events[j].LastTimestamp.Time
			if ti.IsZero() {
				ti = p.events[i].EventTime.Time
			}

			if tj.IsZero() {
				tj = p.events[j].EventTime.Time
			}

			return ti.After(tj)
		})
		p.applyFilter()

		return p, nil

	case RefreshMsg:
		if msg.PanelName == p.Title() {
			return p, p.Refresh()
		}
	}

	return p, nil
}

func (p *EventsPanel) View() string {
	var b strings.Builder

	title := fmt.Sprintf("%s [%s]", p.title, p.shortcutKey)
	if p.focused {
		b.WriteString(p.styles.PanelTitleActive.Render(title))
	} else {
		b.WriteString(p.styles.PanelTitle.Render(title))
	}

	b.WriteString("\n")

	visibleHeight := p.height - 3
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	startIdx := 0
	if p.cursor >= visibleHeight {
		startIdx = p.cursor - visibleHeight + 1
	}

	endIdx := startIdx + visibleHeight
	if endIdx > len(p.filtered) {
		endIdx = len(p.filtered)
	}

	for i := startIdx; i < endIdx; i++ {
		event := p.filtered[i]
		line := p.renderEventLine(event, i == p.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	style := p.styles.Panel
	if p.focused {
		style = p.styles.PanelFocused
	}

	return style.Width(p.width).Height(p.height).Render(b.String())
}

func (p *EventsPanel) renderEventLine(event corev1.Event, selected bool) string {
	reason := utils.Truncate(event.Reason, p.width-15)
	eventType := event.Type

	var line string
	if selected {
		line = "> " + reason
	} else {
		line = "  " + reason
	}

	line = utils.PadRight(line, p.width-10)

	typeStyle := p.styles.StatusRunning
	if eventType == "Warning" {
		typeStyle = p.styles.StatusWarning
	}

	line += " " + typeStyle.Render(utils.Truncate(eventType, 8))

	if selected && p.focused {
		return p.styles.ListItemFocused.Render(line)
	} else if selected {
		return p.styles.ListItemSelected.Render(line)
	}

	return p.styles.ListItem.Render(line)
}

func (p *EventsPanel) DetailView(width, height int) string {
	if p.cursor >= len(p.filtered) {
		return "No event selected"
	}

	event := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(p.styles.DetailTitle.Render("Event"))
	b.WriteString("\n\n")

	typeStyle := p.styles.StatusRunning
	if event.Type == "Warning" {
		typeStyle = p.styles.StatusWarning
	}

	b.WriteString(p.styles.DetailLabel.Render("Type:"))
	b.WriteString(typeStyle.Render(event.Type))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Reason:"))
	b.WriteString(p.styles.DetailValue.Render(event.Reason))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Object:"))
	b.WriteString(
		p.styles.DetailValue.Render(
			fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name),
		),
	)
	b.WriteString("\n")

	lastSeen := event.LastTimestamp.Time
	if lastSeen.IsZero() {
		lastSeen = event.EventTime.Time
	}

	b.WriteString(p.styles.DetailLabel.Render("Last Seen:"))
	b.WriteString(p.styles.DetailValue.Render(utils.FormatAge(lastSeen)))
	b.WriteString("\n")

	firstSeen := event.FirstTimestamp.Time
	if !firstSeen.IsZero() {
		b.WriteString(p.styles.DetailLabel.Render("First Seen:"))
		b.WriteString(p.styles.DetailValue.Render(utils.FormatAge(firstSeen)))
		b.WriteString("\n")
	}

	b.WriteString(p.styles.DetailLabel.Render("Count:"))
	b.WriteString(p.styles.DetailValue.Render(fmt.Sprintf("%d", event.Count)))
	b.WriteString("\n")

	if p.allNs {
		b.WriteString(p.styles.DetailLabel.Render("Namespace:"))
		b.WriteString(p.styles.DetailValue.Render(event.Namespace))
		b.WriteString("\n")
	}

	if event.Source.Component != "" {
		b.WriteString(p.styles.DetailLabel.Render("Source:"))

		source := event.Source.Component
		if event.Source.Host != "" {
			source += " on " + event.Source.Host
		}

		b.WriteString(p.styles.DetailValue.Render(source))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(p.styles.DetailTitle.Render("Message:"))
	b.WriteString("\n")

	wrapped := utils.WrapText(event.Message, width-4)
	for _, line := range wrapped {
		b.WriteString("  " + line + "\n")
	}

	return b.String()
}

func (p *EventsPanel) Refresh() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var (
			events []corev1.Event
			err    error
		)

		if p.allNs {
			events, err = p.client.ListEventsAllNamespaces(ctx)
		} else {
			events, err = p.client.ListEvents(ctx, "")
		}

		if err != nil {
			return ErrorMsg{Error: err}
		}

		return eventsLoadedMsg{events: events}
	}
}

func (p *EventsPanel) Delete() tea.Cmd {
	return func() tea.Msg {
		return StatusMsg{Message: "Events cannot be deleted"}
	}
}

func (p *EventsPanel) SelectedItem() interface{} {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	return &p.filtered[p.cursor]
}

func (p *EventsPanel) SelectedName() string {
	if p.cursor >= len(p.filtered) {
		return ""
	}

	return p.filtered[p.cursor].Name
}

func (p *EventsPanel) GetSelectedYAML() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	event := p.filtered[p.cursor]

	data, err := yaml.Marshal(event)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (p *EventsPanel) GetSelectedDescribe() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	event := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Name:          %s\n", event.Name))
	b.WriteString(fmt.Sprintf("Namespace:     %s\n", event.Namespace))
	b.WriteString(fmt.Sprintf("Type:          %s\n", event.Type))
	b.WriteString(fmt.Sprintf("Reason:        %s\n", event.Reason))
	b.WriteString(
		fmt.Sprintf("Object:        %s/%s\n", event.InvolvedObject.Kind, event.InvolvedObject.Name),
	)
	b.WriteString(fmt.Sprintf("Count:         %d\n", event.Count))

	if !event.FirstTimestamp.IsZero() {
		b.WriteString(
			fmt.Sprintf("First Seen:    %s\n", utils.FormatTimestampFromMeta(event.FirstTimestamp)),
		)
	}

	if !event.LastTimestamp.IsZero() {
		b.WriteString(
			fmt.Sprintf("Last Seen:     %s\n", utils.FormatTimestampFromMeta(event.LastTimestamp)),
		)
	}

	if event.Source.Component != "" {
		b.WriteString(fmt.Sprintf("Source:        %s\n", event.Source.Component))
	}

	b.WriteString(fmt.Sprintf("\nMessage:\n%s\n", event.Message))

	return b.String(), nil
}

func (p *EventsPanel) applyFilter() {
	if p.filter == "" {
		p.filtered = p.events

		return
	}

	filter := strings.ToLower(p.filter)

	p.filtered = make([]corev1.Event, 0)
	for _, event := range p.events {
		if strings.Contains(strings.ToLower(event.Reason), filter) ||
			strings.Contains(strings.ToLower(event.Message), filter) ||
			strings.Contains(strings.ToLower(event.InvolvedObject.Name), filter) {
			p.filtered = append(p.filtered, event)
		}
	}

	if p.cursor >= len(p.filtered) {
		p.cursor = len(p.filtered) - 1
		if p.cursor < 0 {
			p.cursor = 0
		}
	}
}

func (p *EventsPanel) SetFilter(query string) {
	p.BasePanel.SetFilter(query)
	p.applyFilter()
}

type eventsLoadedMsg struct {
	events []corev1.Event
}
