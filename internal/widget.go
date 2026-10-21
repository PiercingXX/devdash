package internal

import "strings"

const (
	// Data
	optionCommand = "command"

	// Widget config options
	optionTitle      = "title"
	optionTitleColor = "title_color"

	// Monitor
	optionAddress = "address"

	// Time
	optionStartDate  = "start_date"
	optionEndDate    = "end_date"
	optionTimePeriod = "time_period"
	optionGlobal     = "global"

	// Tables
	optionRowLimit  = "row_limit"
	optionCharLimit = "character_limit"

	// Metrics
	optionDimension  = "dimension"
	optionDimensions = "dimensions"
	optionMetrics    = "metrics"
	optionMetric     = "metric"
	optionUnit       = "unit"

	optionHeaders = "headers"

	// Ordering
	optionOrder = "order"

	// Filtering
	optionFilters = "filters"

	// Display
	optionContent = "content"

	// Repository
	optionRepository = "repository"

	// Commits of the owner or of everybody
	optionScope = "scope"
	ownerScope  = "owner"
)

// A widget is a representation of a set of data in the screen.
type Widget struct {
	Name    string            `mapstructure:"name"`
	Options map[string]string `mapstructure:"options"`
	Theme   string            `mapstructure:"theme"`
}

// typeID of the widget: "bar" for "ga.bar_pages".
func (w *Widget) typeID() string {
	_, n, _ := strings.Cut(w.Name, ".")
	t, _, _ := strings.Cut(n, "_")
	return t
}

// serviceID of the widget: "ga" for "ga.bar_pages".
func (w *Widget) serviceID() string {
	s, _, _ := strings.Cut(w.Name, ".")
	return s
}

// truncate a text to a number of characters.
func truncate(text string, limit int) string {
	if r := []rune(text); limit >= 0 && len(r) > limit {
		return string(r[:limit])
	}

	return text
}
