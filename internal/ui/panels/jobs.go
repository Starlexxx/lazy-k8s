package panels

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/lazyk8s/lazy-k8s/internal/k8s"
	"github.com/lazyk8s/lazy-k8s/internal/ui/theme"
	"github.com/lazyk8s/lazy-k8s/internal/utils"
)

type JobsPanel struct {
	BasePanel
	client   *k8s.Client
	styles   *theme.Styles
	jobs     []batchv1.Job
	filtered []batchv1.Job
}

func NewJobsPanel(client *k8s.Client, styles *theme.Styles) *JobsPanel {
	return &JobsPanel{
		BasePanel: BasePanel{
			title:       "Jobs",
			shortcutKey: "9",
		},
		client: client,
		styles: styles,
	}
}

func (p *JobsPanel) Init() tea.Cmd {
	return p.Refresh()
}

func (p *JobsPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
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

	case jobsLoadedMsg:
		p.jobs = msg.jobs
		p.applyFilter()

		return p, nil

	case RefreshMsg:
		if msg.PanelName == p.Title() {
			return p, p.Refresh()
		}
	}

	return p, nil
}

func (p *JobsPanel) View() string {
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
		job := p.filtered[i]
		line := p.renderJobLine(job, i == p.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	style := p.styles.Panel
	if p.focused {
		style = p.styles.PanelFocused
	}

	return style.Width(p.width).Height(p.height).Render(b.String())
}

func (p *JobsPanel) renderJobLine(job batchv1.Job, selected bool) string {
	name := utils.Truncate(job.Name, p.width-15)
	status := p.getJobStatus(&job)

	var line string
	if selected {
		line = "> " + name
	} else {
		line = "  " + name
	}

	line = utils.PadRight(line, p.width-12)
	statusStyle := p.styles.GetStatusStyle(status)
	line += " " + statusStyle.Render(utils.Truncate(status, 10))

	if selected && p.focused {
		return p.styles.ListItemFocused.Render(line)
	} else if selected {
		return p.styles.ListItemSelected.Render(line)
	}

	return p.styles.ListItem.Render(line)
}

func (p *JobsPanel) getJobStatus(job *batchv1.Job) string {
	if job.Status.Succeeded > 0 {
		return "Completed"
	}

	if job.Status.Failed > 0 {
		return "Failed"
	}

	if job.Status.Active > 0 {
		return "Running"
	}

	return "Pending"
}

func (p *JobsPanel) DetailView(width, height int) string {
	if p.cursor >= len(p.filtered) {
		return "No job selected"
	}

	job := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(p.styles.DetailTitle.Render("Job: " + job.Name))
	b.WriteString("\n\n")

	status := p.getJobStatus(&job)
	b.WriteString(p.styles.DetailLabel.Render("Status:"))
	b.WriteString(p.styles.GetStatusStyle(status).Render(status))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Completions:"))

	completions := int32(1)
	if job.Spec.Completions != nil {
		completions = *job.Spec.Completions
	}

	b.WriteString(
		p.styles.DetailValue.Render(fmt.Sprintf("%d/%d", job.Status.Succeeded, completions)),
	)
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Active:"))
	b.WriteString(p.styles.DetailValue.Render(fmt.Sprintf("%d", job.Status.Active)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Failed:"))
	b.WriteString(p.styles.DetailValue.Render(fmt.Sprintf("%d", job.Status.Failed)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Age:"))
	b.WriteString(p.styles.DetailValue.Render(utils.FormatAgeFromMeta(job.CreationTimestamp)))
	b.WriteString("\n")

	if p.allNs {
		b.WriteString(p.styles.DetailLabel.Render("Namespace:"))
		b.WriteString(p.styles.DetailValue.Render(job.Namespace))
		b.WriteString("\n")
	}

	// Timestamps
	if job.Status.StartTime != nil {
		b.WriteString(p.styles.DetailLabel.Render("Start Time:"))
		b.WriteString(
			p.styles.DetailValue.Render(utils.FormatTimestampFromMeta(*job.Status.StartTime)),
		)
		b.WriteString("\n")
	}

	if job.Status.CompletionTime != nil {
		b.WriteString(p.styles.DetailLabel.Render("Completion:"))
		b.WriteString(
			p.styles.DetailValue.Render(utils.FormatTimestampFromMeta(*job.Status.CompletionTime)),
		)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(p.styles.Muted.Render("[d]escribe [y]aml [l]ogs [D]elete"))

	return b.String()
}

func (p *JobsPanel) Refresh() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var (
			jobs *batchv1.JobList
			err  error
		)

		if p.allNs {
			jobs, err = p.client.Clientset().BatchV1().Jobs("").List(ctx, metav1.ListOptions{})
		} else {
			jobs, err = p.client.Clientset().
				BatchV1().
				Jobs(p.client.CurrentNamespace()).
				List(ctx, metav1.ListOptions{})
		}

		if err != nil {
			return ErrorMsg{Error: err}
		}

		return jobsLoadedMsg{jobs: jobs.Items}
	}
}

func (p *JobsPanel) Delete() tea.Cmd {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	job := p.filtered[p.cursor]

	return func() tea.Msg {
		ctx := context.Background()
		propagation := metav1.DeletePropagationBackground

		err := p.client.Clientset().
			BatchV1().
			Jobs(job.Namespace).
			Delete(ctx, job.Name, metav1.DeleteOptions{
				PropagationPolicy: &propagation,
			})
		if err != nil {
			return ErrorMsg{Error: err}
		}

		return StatusMsg{Message: fmt.Sprintf("Deleted job: %s", job.Name)}
	}
}

func (p *JobsPanel) SelectedItem() interface{} {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	return &p.filtered[p.cursor]
}

func (p *JobsPanel) SelectedName() string {
	if p.cursor >= len(p.filtered) {
		return ""
	}

	return p.filtered[p.cursor].Name
}

func (p *JobsPanel) GetSelectedYAML() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	job := p.filtered[p.cursor]

	data, err := yaml.Marshal(job)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (p *JobsPanel) GetSelectedDescribe() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	job := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Name:           %s\n", job.Name))
	b.WriteString(fmt.Sprintf("Namespace:      %s\n", job.Namespace))
	b.WriteString(fmt.Sprintf("Status:         %s\n", p.getJobStatus(&job)))

	completions := int32(1)
	if job.Spec.Completions != nil {
		completions = *job.Spec.Completions
	}

	b.WriteString(fmt.Sprintf("Completions:    %d/%d\n", job.Status.Succeeded, completions))
	b.WriteString(fmt.Sprintf("Active:         %d\n", job.Status.Active))
	b.WriteString(fmt.Sprintf("Failed:         %d\n", job.Status.Failed))

	if len(job.Labels) > 0 {
		b.WriteString("\nLabels:\n")

		for k, v := range job.Labels {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	return b.String(), nil
}

func (p *JobsPanel) applyFilter() {
	if p.filter == "" {
		p.filtered = p.jobs

		return
	}

	p.filtered = make([]batchv1.Job, 0)
	for _, job := range p.jobs {
		if strings.Contains(strings.ToLower(job.Name), strings.ToLower(p.filter)) {
			p.filtered = append(p.filtered, job)
		}
	}

	if p.cursor >= len(p.filtered) {
		p.cursor = len(p.filtered) - 1
		if p.cursor < 0 {
			p.cursor = 0
		}
	}
}

func (p *JobsPanel) SetFilter(query string) {
	p.BasePanel.SetFilter(query)
	p.applyFilter()
}

type jobsLoadedMsg struct {
	jobs []batchv1.Job
}
