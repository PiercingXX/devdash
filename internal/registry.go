package internal

import (
	"slices"
	"strings"
)

// The registry lists every widget and the options it accepts. It's used to check the dashboards and to generate
// their JSON schema.

// displayOptions of each type of widget.
var displayOptions = map[string][]string{
	"box":   {optionTitle, optionHeight, optionColor, optionTitleColor, optionBorderColor, optionTextColor, optionBold, optionMultiline},
	"gauge": {optionTitle, optionHeight, optionColor, optionTitleColor, optionBorderColor, optionTextColor, optionBarColor},
	"bar": {
		optionTitle, optionHeight, optionBarWidth, optionBarGap, optionColor, optionTitleColor, optionBorderColor,
		optionTextColor, optionBarColor, optionNumColor, optionEmptyNumColor,
	},
	"table": {optionTitle, optionColor, optionTitleColor, optionBorderColor, optionTextColor},
}

// stackedBarOptions replace bar_color and empty_num_color for the stacked bar charts.
var stackedBarOptions = []string{
	optionTitle, optionHeight, optionBarWidth, optionBarGap, optionColor, optionTitleColor, optionBorderColor,
	optionTextColor, optionNumColor, optionFirstColor, optionSecondColor, optionThirdColor, optionFourthColor,
	optionFifthColor,
}

var (
	gaBarOptions  = []string{optionStartDate, optionEndDate, optionGlobal, optionTimePeriod}
	gaTableOpts   = []string{optionStartDate, optionEndDate, optionGlobal, optionMetrics, optionOrder, optionFilters, optionRowLimit, optionCharLimit}
	gscTableOpts  = []string{optionStartDate, optionEndDate, optionMetrics, optionFilters, optionRowLimit, optionCharLimit}
	githubRepo    = []string{optionRepository}
	githubTable   = []string{optionRepository, optionRowLimit}
	jsonRequest   = []string{optionURL, optionPath, optionHeader}
	hostIOOptions = []string{optionUnit, optionRate}
)

// widgetOptions are the options of each widget, in addition to the display options of its type.
var widgetOptions = map[string][]string{
	boxAvailability: {optionAddress},
	boxPing:         {optionAddress},

	rhUptime:        nil,
	rhLoad:          nil,
	rhProcesses:     nil,
	rhBoxCPURate:    nil,
	rhBoxMemRate:    nil,
	rhBoxSwapRate:   nil,
	rhGaugeCPURate:  nil,
	rhGaugeMemRate:  nil,
	rhGaugeSwapRate: nil,
	rhBarRates:      nil,
	rhBoxNetIO:      hostIOOptions,
	rhBoxDiskIO:     hostIOOptions,
	rhBarMemory:     {optionUnit, optionMetrics, optionHeaders},
	rhTableDisk:     {optionUnit, optionHeaders},
	rhBox:           {optionCommand},
	rhGauge:         {optionCommand},
	rhBar:           {optionCommand, optionHeaders},
	rhTable:         {optionCommand, optionHeaders},

	githubBoxStars:          githubRepo,
	githubBoxWatchers:       githubRepo,
	githubBoxOpenIssues:     githubRepo,
	githubTableRepositories: {optionRowLimit, optionMetrics, optionOrder},
	githubTableBranches:     githubTable,
	githubTableIssues:       githubTable,
	githubTablePullRequests: githubTable,
	githubBarViews:          githubRepo,
	githubBarCommits:        {optionRepository, optionStartDate, optionEndDate, optionScope},
	githubBarStars:          {optionRepository, optionStartDate, optionEndDate},
	githubTableWorkflowRuns: {optionRepository, optionBranch, optionRowLimit},
	githubBoxLastRun:        {optionRepository, optionBranch},

	gaBoxRealtime:         nil,
	gaBoxTotal:            {optionStartDate, optionEndDate, optionGlobal, optionMetric},
	gaBar:                 append(slices.Clone(gaBarOptions), optionMetric, optionDimensions, optionFilters),
	gaBarSessions:         append(slices.Clone(gaBarOptions), optionMetric, optionDimensions, optionFilters),
	gaBarUsers:            append(slices.Clone(gaBarOptions), optionDimensions, optionFilters),
	gaBarBounces:          append(slices.Clone(gaBarOptions), optionDimensions, optionFilters),
	gaBarReturning:        gaBarOptions,
	gaBarPages:            append(slices.Clone(gaBarOptions), optionFilters),
	gaBarCountries:        {optionStartDate, optionEndDate, optionGlobal, optionFilters},
	gaBarNewReturning:     {optionStartDate, optionEndDate, optionTimePeriod, optionMetric},
	gaBarDevices:          {optionStartDate, optionEndDate, optionTimePeriod, optionMetric},
	gaTablePages:          append(slices.Clone(gaTableOpts), optionDimension),
	gaTableTrafficSources: gaTableOpts,
	gaTable:               append(slices.Clone(gaTableOpts), optionDimension),

	gscTableQueries: append(slices.Clone(gscTableOpts), optionDimension),
	gscTablePages:   gscTableOpts,
	gscTable:        append(slices.Clone(gscTableOpts), optionDimension),

	FeedlySubscribers: nil,
	gitBranches:       nil,
	displayBox:        {optionContent},

	jsonBox:   jsonRequest,
	jsonGauge: jsonRequest,
	jsonBar:   append(slices.Clone(jsonRequest), optionLabel, optionValue),
	jsonTable: append(slices.Clone(jsonRequest), optionColumns, optionHeaders, optionRowLimit),
}

// stackedBars use the stacked bar options instead of the bar options.
var stackedBars = []string{gaBarNewReturning, gaBarDevices}

// WidgetNames of every widget, sorted. The host widgets are listed for the local host (lh) and the remote host (rh).
func WidgetNames() []string {
	names := []string{}
	for n := range widgetOptions {
		names = append(names, n)
		if lh, ok := strings.CutPrefix(n, "rh."); ok {
			names = append(names, "lh."+lh)
		}
	}
	slices.Sort(names)

	return names
}

// WidgetOptions accepted by a widget: its own options and the display options of its type. False if the widget
// doesn't exist.
func WidgetOptions(name string) ([]string, bool) {
	own, ok := widgetOptions[strings.Replace(name, "lh.", "rh.", 1)]
	if !ok {
		return nil, false
	}

	display := displayOptions[(&Widget{Name: name}).typeID()]
	if slices.Contains(stackedBars, name) {
		display = stackedBarOptions
	}

	options := append(slices.Clone(own), display...)
	slices.Sort(options)

	return slices.Compact(options), true
}

// AllOptions of every widget, sorted.
func AllOptions() []string {
	all := []string{}
	for _, n := range WidgetNames() {
		o, _ := WidgetOptions(n)
		all = append(all, o...)
	}
	slices.Sort(all)

	return slices.Compact(all)
}
