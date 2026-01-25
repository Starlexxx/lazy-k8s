package layout

import (
	"github.com/charmbracelet/lipgloss"
)

type LayoutType string

const (
	LayoutVertical   LayoutType = "vertical"
	LayoutHorizontal LayoutType = "horizontal"
	LayoutGrid       LayoutType = "grid"
)

type Layout struct {
	layoutType     LayoutType
	width          int
	height         int
	leftPanelRatio float64
}

func NewLayout(layoutType LayoutType) *Layout {
	return &Layout{
		layoutType:     layoutType,
		leftPanelRatio: 0.25, // 25% for left panel
	}
}

func (l *Layout) SetSize(width, height int) {
	l.width = width
	l.height = height
}

func (l *Layout) SetLeftPanelRatio(ratio float64) {
	if ratio > 0 && ratio < 1 {
		l.leftPanelRatio = ratio
	}
}

// LeftPanelWidth returns the width for the left panel area.
func (l *Layout) LeftPanelWidth() int {
	return int(float64(l.width) * l.leftPanelRatio)
}

// RightPanelWidth returns the width for the right panel area (detail view).
func (l *Layout) RightPanelWidth() int {
	return l.width - l.LeftPanelWidth() - 1 // -1 for separator
}

// PanelHeight returns the height for each panel given the count.
func (l *Layout) PanelHeight(panelCount int, availableHeight int) int {
	if panelCount == 0 {
		return availableHeight
	}

	return availableHeight / panelCount
}

// ContentHeight returns the height available for content (excluding header/footer).
func (l *Layout) ContentHeight(headerHeight, footerHeight int) int {
	return l.height - headerHeight - footerHeight
}

// JoinPanelsVertical joins multiple panel views vertically.
func JoinPanelsVertical(panels ...string) string {
	return lipgloss.JoinVertical(lipgloss.Left, panels...)
}

// JoinPanelsHorizontal joins multiple panel views horizontally.
func JoinPanelsHorizontal(panels ...string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, panels...)
}

// CenterContent centers content within the given dimensions.
func CenterContent(content string, width, height int) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

// Dimensions holds width and height.
type Dimensions struct {
	Width  int
	Height int
}

// CalculatePanelDimensions calculates dimensions for a set of panels.
func (l *Layout) CalculatePanelDimensions(
	panelCount int,
	headerHeight, footerHeight int,
) []Dimensions {
	if panelCount == 0 {
		return nil
	}

	contentHeight := l.ContentHeight(headerHeight, footerHeight)
	panelWidth := l.LeftPanelWidth()
	panelHeight := contentHeight / panelCount

	dims := make([]Dimensions, panelCount)
	for i := 0; i < panelCount; i++ {
		dims[i] = Dimensions{
			Width:  panelWidth,
			Height: panelHeight,
		}
	}

	// Give extra height to last panel if there's remainder
	remainder := contentHeight % panelCount
	if remainder > 0 {
		dims[panelCount-1].Height += remainder
	}

	return dims
}

// CalculateDetailDimensions returns dimensions for the detail panel.
func (l *Layout) CalculateDetailDimensions(headerHeight, footerHeight int) Dimensions {
	return Dimensions{
		Width:  l.RightPanelWidth(),
		Height: l.ContentHeight(headerHeight, footerHeight),
	}
}
