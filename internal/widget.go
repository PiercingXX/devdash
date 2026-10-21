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

func (w *Widget) typeID() string {
	n := strings.Split(w.Name, ".")[1]
	return strings.Split(n, "_")[0]
}

func (w *Widget) serviceID() string {
	return strings.Split(w.Name, ".")[0]
}

// truncate a text to a number of characters.
func truncate(text string, limit int) string {
	if r := []rune(text); limit >= 0 && len(r) > limit {
		return string(r[:limit])
	}

	return text
}
