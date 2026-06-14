package panels

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	batchv1 "k8s.io/api/batch/v1"

	"github.com/Starlexxx/lazy-k8s/internal/k8s"
	"github.com/Starlexxx/lazy-k8s/internal/ui/theme"
	"github.com/Starlexxx/lazy-k8s/internal/utils"
)

type CronJobsPanel struct {
	BasePanel
	client   *k8s.Client
	styles   *theme.Styles
	cronjobs []batchv1.CronJob
	filtered []batchv1.CronJob
}

func NewCronJobsPanel(client *k8s.Client, styles *theme.Styles) *CronJobsPanel {
	return &CronJobsPanel{
		BasePanel: BasePanel{
			title:       "CronJobs",
			shortcutKey: "",
		},
		client: client,
		styles: styles,
	}
}

func (p *CronJobsPanel) Init() tea.Cmd {
	return p.Refresh()
}

func (p *CronJobsPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
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
		case key.Matches(msg, key.NewBinding(key.WithKeys("t"))):
			if p.cursor >= len(p.filtered) {
				return p, nil
			}

			cj := p.filtered[p.cursor]

			return p, func() tea.Msg {
				return TriggerCronJobRequestMsg{
					CronJobName: cj.Name,
					Namespace:   cj.Namespace,
				}
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("S"))):
			if p.cursor >= len(p.filtered) {
				return p, nil
			}

			cj := p.filtered[p.cursor]

			currentSuspend := false
			if cj.Spec.Suspend != nil {
				currentSuspend = *cj.Spec.Suspend
			}

			return p, func() tea.Msg {
				return ToggleSuspendCronJobRequestMsg{
					CronJobName:    cj.Name,
					Namespace:      cj.Namespace,
					CurrentSuspend: currentSuspend,
				}
			}
		}

	case cronJobsLoadedMsg:
		p.cronjobs = msg.cronjobs
		p.applyFilter()

		return p, nil

	case RefreshMsg:
		if msg.PanelName == p.Title() {
			return p, p.Refresh()
		}
	}

	return p, nil
}

func (p *CronJobsPanel) View() string {
	var b strings.Builder

	title := p.renderTitle()
	if p.focused {
		b.WriteString(p.styles.PanelTitleActive.Render(title))
	} else {
		b.WriteString(p.styles.PanelTitle.Render(title))
	}

	b.WriteString("\n")

	startIdx, endIdx := p.visibleWindow(len(p.filtered), 0)

	for i := startIdx; i < endIdx; i++ {
		cj := p.filtered[i]
		line := p.renderCronJobLine(cj, i == p.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	style := p.styles.Panel
	if p.focused {
		style = p.styles.PanelFocused
	}

	return style.Width(p.width).Height(p.height).Render(b.String())
}

func (p *CronJobsPanel) renderCronJobLine(cj batchv1.CronJob, selected bool) string {
	status := k8s.GetCronJobStatus(&cj)

	statusStyle := p.styles.StatusRunning
	if status == "Suspended" {
		statusStyle = p.styles.StatusPending
	}

	var line string
	if selected {
		line = "> "
	} else {
		line = "  "
	}

	if p.width > 80 {
		reserved := 38
		if p.width > 120 && p.allNs {
			reserved += 16
		}

		nameW := max(p.width-reserved, 10)

		line += utils.PadRight(
			utils.Truncate(cj.Name, nameW), nameW,
		)
		line += " " + statusStyle.Render(
			utils.PadRight(utils.Truncate(status, 10), 10),
		)

		schedule := utils.Truncate(cj.Spec.Schedule, 15)
		line += " " + utils.PadRight(schedule, 15)

		age := utils.FormatAgeFromMeta(cj.CreationTimestamp)
		line += " " + utils.PadRight(age, 8)

		if p.width > 120 && p.allNs {
			line += " " + utils.Truncate(cj.Namespace, 15)
		}
	} else {
		name := utils.Truncate(cj.Name, p.width-15)
		line += name
		line = utils.PadRight(line, p.width-12)
		line += " " + statusStyle.Render(status)
	}

	if selected && p.focused {
		return p.styles.ListItemFocused.Render(line)
	} else if selected {
		return p.styles.ListItemSelected.Render(line)
	}

	return p.styles.ListItem.Render(line)
}

func (p *CronJobsPanel) DetailView(width, height int) string {
	if p.cursor >= len(p.filtered) {
		return "No cronjob selected"
	}

	cj := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(p.styles.DetailTitle.Render("CronJob: " + cj.Name))
	b.WriteString("\n\n")

	b.WriteString(p.styles.DetailLabel.Render("Schedule:"))
	b.WriteString(p.styles.DetailValue.Render(cj.Spec.Schedule))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Status:"))
	b.WriteString(p.styles.DetailValue.Render(k8s.GetCronJobStatus(&cj)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Last Schedule:"))
	b.WriteString(p.styles.DetailValue.Render(k8s.GetCronJobLastSchedule(&cj)))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Active Jobs:"))
	b.WriteString(p.styles.DetailValue.Render(fmt.Sprintf("%d", k8s.GetCronJobActiveJobs(&cj))))
	b.WriteString("\n")

	b.WriteString(p.styles.DetailLabel.Render("Age:"))
	b.WriteString(p.styles.DetailValue.Render(utils.FormatAgeFromMeta(cj.CreationTimestamp)))
	b.WriteString("\n")

	if p.allNs {
		b.WriteString(p.styles.DetailLabel.Render("Namespace:"))
		b.WriteString(p.styles.DetailValue.Render(cj.Namespace))
		b.WriteString("\n")
	}

	if cj.Spec.ConcurrencyPolicy != "" {
		b.WriteString(p.styles.DetailLabel.Render("Concurrency:"))
		b.WriteString(p.styles.DetailValue.Render(string(cj.Spec.ConcurrencyPolicy)))
		b.WriteString("\n")
	}

	if cj.Spec.SuccessfulJobsHistoryLimit != nil {
		b.WriteString(p.styles.DetailLabel.Render("Success History:"))
		b.WriteString(
			p.styles.DetailValue.Render(fmt.Sprintf("%d", *cj.Spec.SuccessfulJobsHistoryLimit)),
		)
		b.WriteString("\n")
	}

	if cj.Spec.FailedJobsHistoryLimit != nil {
		b.WriteString(p.styles.DetailLabel.Render("Failed History:"))
		b.WriteString(
			p.styles.DetailValue.Render(fmt.Sprintf("%d", *cj.Spec.FailedJobsHistoryLimit)),
		)
		b.WriteString("\n")
	}

	if len(cj.Status.Active) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Active Jobs:"))
		b.WriteString("\n")

		for _, ref := range cj.Status.Active {
			b.WriteString(fmt.Sprintf("  %s\n", ref.Name))
		}
	}

	images := getCronJobImages(&cj)
	if len(images) > 0 {
		b.WriteString("\n")
		b.WriteString(p.styles.DetailTitle.Render("Images:"))
		b.WriteString("\n")

		for _, img := range images {
			b.WriteString("  " + img + "\n")
		}
	}

	b.WriteString("\n")

	suspendAction := "[S]uspend"
	if cj.Spec.Suspend != nil && *cj.Spec.Suspend {
		suspendAction = "[S] Resume"
	}

	b.WriteString(
		p.styles.Muted.Render(
			fmt.Sprintf("[t]rigger %s [d]escribe [y]aml [D]elete", suspendAction),
		),
	)

	return b.String()
}

func getCronJobImages(cj *batchv1.CronJob) []string {
	images := make([]string, 0)
	for _, container := range cj.Spec.JobTemplate.Spec.Template.Spec.Containers {
		images = append(images, container.Image)
	}

	return images
}

func (p *CronJobsPanel) Refresh() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var (
			cronjobs []batchv1.CronJob
			err      error
		)

		if p.allNs {
			cronjobs, err = p.client.ListCronJobsAllNamespaces(ctx)
		} else {
			cronjobs, err = p.client.ListCronJobs(ctx, "")
		}

		if err != nil {
			return ErrorMsg{Error: err}
		}

		return cronJobsLoadedMsg{cronjobs: cronjobs}
	}
}

func (p *CronJobsPanel) Delete() tea.Cmd {
	if p.cursor >= len(p.filtered) {
		return nil
	}

	cj := p.filtered[p.cursor]

	return func() tea.Msg {
		ctx := context.Background()

		err := p.client.DeleteCronJob(ctx, cj.Namespace, cj.Name)
		if err != nil {
			return ErrorMsg{Error: err}
		}

		return StatusMsg{Message: fmt.Sprintf("Deleted cronjob: %s", cj.Name)}
	}
}

func (p *CronJobsPanel) SelectedItem() any {
	item := selectedItem(p.filtered, p.cursor)
	if item == nil {
		return nil
	}

	return item
}

func (p *CronJobsPanel) SelectedName() string {
	return selectedName(p.filtered, p.cursor, func(cj batchv1.CronJob) string { return cj.Name })
}

func (p *CronJobsPanel) GetSelectedYAML() (string, error) {
	return marshalSelectedYAML(p.filtered, p.cursor)
}

func (p *CronJobsPanel) GetSelectedDescribe() (string, error) {
	if p.cursor >= len(p.filtered) {
		return "", ErrNoSelection
	}

	cj := p.filtered[p.cursor]

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Name:               %s\n", cj.Name))
	b.WriteString(fmt.Sprintf("Namespace:          %s\n", cj.Namespace))
	b.WriteString(
		fmt.Sprintf(
			"CreationTimestamp:  %s\n",
			utils.FormatTimestampFromMeta(cj.CreationTimestamp),
		),
	)
	b.WriteString(fmt.Sprintf("Schedule:           %s\n", cj.Spec.Schedule))
	b.WriteString(fmt.Sprintf("Concurrency Policy: %s\n", cj.Spec.ConcurrencyPolicy))

	suspend := "False"
	if cj.Spec.Suspend != nil && *cj.Spec.Suspend {
		suspend = "True"
	}

	b.WriteString(fmt.Sprintf("Suspend:            %s\n", suspend))
	b.WriteString(fmt.Sprintf("Last Schedule Time: %s\n", k8s.GetCronJobLastSchedule(&cj)))
	b.WriteString(fmt.Sprintf("Active Jobs:        %d\n", len(cj.Status.Active)))

	if len(cj.Labels) > 0 {
		b.WriteString("\nLabels:\n")

		for k, v := range cj.Labels {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	b.WriteString("\nJob Template:\n")

	for _, container := range cj.Spec.JobTemplate.Spec.Template.Spec.Containers {
		b.WriteString(fmt.Sprintf("  Container: %s\n", container.Name))
		b.WriteString(fmt.Sprintf("    Image:   %s\n", container.Image))
	}

	return b.String(), nil
}

func (p *CronJobsPanel) applyFilter() {
	p.filtered = filterByName(
		p.cronjobs, p.filter, func(c batchv1.CronJob) string { return c.Name }, &p.cursor,
	)
}

func (p *CronJobsPanel) SetFilter(query string) {
	p.BasePanel.SetFilter(query)
	p.applyFilter()
}

func (p *CronJobsPanel) SearchItems(query string) []SearchResult {
	return searchByName(
		p.cronjobs,
		query,
		p.title,
		func(cj batchv1.CronJob) string { return cj.Name },
		func(cj batchv1.CronJob) string { return cj.Namespace },
		func(cj batchv1.CronJob) string { return k8s.GetCronJobStatus(&cj) },
	)
}

func (p *CronJobsPanel) NavigateTo(name, namespace string) bool {
	return navigateTo(
		p.filtered,
		&p.cursor,
		func(cj batchv1.CronJob) string { return cj.Name },
		func(cj batchv1.CronJob) string { return cj.Namespace },
		name,
		namespace,
	)
}

type cronJobsLoadedMsg struct {
	cronjobs []batchv1.CronJob
}
