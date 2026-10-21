package internal

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PiercingXX/devdash/internal/platform"
	"github.com/pkg/errors"
)

const (
	// widget config names
	gscTablePages   = "gsc.table_pages"
	gscTableQueries = "gsc.table_queries"
	gscTable        = "gsc.table"

	// format for every start date / end date
	gscTimeFormat = "2006-01-02"
)

type gscWidget struct {
	tui     *Tui
	client  *platform.SearchConsole
	address string
}

var mappingGscHeader = map[string]string{
	"page":  "Page",
	"query": "Query",
}

// NewGscWidget including everything to connect to the Google Search Console API.
func NewGscWidget(keyfile string, address string) (*gscWidget, error) {
	sc, err := platform.NewSearchConsoleClient(keyfile)
	if err != nil {
		return nil, err
	}

	return &gscWidget{
		client:  sc,
		address: address,
	}, nil
}

// CreateWidgets for the Google Search Console API.
func (s *gscWidget) CreateWidgets(widget Widget, tui *Tui) (f func() error, err error) {
	s.tui = tui
	switch widget.Name {
	case gscTablePages:
		f, err = s.pages(widget)
	case gscTableQueries:
		f, err = s.table(widget)
	case gscTable:
		f, err = s.table(widget)
	default:
		return nil, errors.Errorf("can't find the widget %s", widget.Name)
	}

	return
}

func (s *gscWidget) pages(widget Widget) (f func() error, err error) {
	if widget.Options == nil {
		widget.Options = map[string]string{}
	}

	widget.Options[optionDimension] = "page"

	return s.table(widget)
}

// table of the result of a Google Search Console query.
// If no metric provided, the default is "query" with no filters.
func (s *gscWidget) table(widget Widget) (f func() error, err error) {
	sd := "7_days_ago"
	if _, ok := widget.Options[optionStartDate]; ok {
		sd = widget.Options[optionStartDate]
	}

	ed := "today"
	if _, ok := widget.Options[optionEndDate]; ok {
		ed = widget.Options[optionEndDate]
	}

	startDate, endDate, err := platform.ConvertDates(time.Now(), sd, ed)
	if err != nil {
		return nil, err
	}

	var rowLimit int64 = 5
	if _, ok := widget.Options[optionRowLimit]; ok {
		rowLimit, err = strconv.ParseInt(widget.Options[optionRowLimit], 0, 0)
		if err != nil {
			return nil, errors.Wrapf(err, "%s must be a number", widget.Options[optionRowLimit])
		}
	}

	charLimit := 1000
	if _, ok := widget.Options[optionCharLimit]; ok {
		c, err := strconv.ParseInt(widget.Options[optionCharLimit], 0, 0)
		if err != nil {
			return nil, errors.Wrapf(err, "%s must be a number", widget.Options[optionCharLimit])
		}
		charLimit = int(c)
	}

	dimension := "query"
	if _, ok := widget.Options[optionDimension]; ok {
		dimension = widget.Options[optionDimension]
	}

	filters := ""
	if _, ok := widget.Options[optionFilters]; ok {
		filters = widget.Options[optionFilters]
	}

	metrics := []string{"clicks", "impressions", "ctr", "position"}
	if _, ok := widget.Options[optionMetrics]; ok {
		if len(widget.Options[optionMetrics]) > 0 {
			metrics = strings.Split(widget.Options[optionMetrics], ",")
		}
	}

	title := fmt.Sprintf(
		" Search %s from %s to %s ",
		dimension,
		startDate.Format(gscTimeFormat),
		endDate.Format(gscTimeFormat),
	)
	if _, ok := widget.Options[optionTitle]; ok {
		title = widget.Options[optionTitle]
	}

	results, err := s.client.Table(
		startDate.Format(gscTimeFormat),
		endDate.Format(gscTimeFormat),
		rowLimit,
		s.address,
		dimension,
		filters,
	)
	if err != nil {
		return nil, err
	}

	table := formatNumerics(results, dimension, metrics)
	table = formatText(table, charLimit, s.address)

	f = func() error {
		return s.tui.AddTable(table, title, widget.Options)
	}

	return
}

func formatNumerics(results []platform.SearchConsoleResponse, dimension string, metrics []string) [][]string {
	header, ok := mappingGscHeader[dimension]
	if !ok {
		header = titleize(dimension)
	}

	table := make([][]string, len(results)+1)
	table[0] = []string{header}
	table[0] = append(table[0], metrics...)

	for k, v := range results {
		table[k+1] = append(table[k+1], v.Dimension)
		for _, m := range metrics {
			if m == "clicks" {
				table[k+1] = append(table[k+1], fmt.Sprintf("%g", v.Clicks))
			}
			if m == "impressions" {
				table[k+1] = append(table[k+1], fmt.Sprintf("%g", v.Impressions))
			}
			if m == "ctr" {
				table[k+1] = append(table[k+1], fmt.Sprintf("%.2f%%", v.Ctr*100))
			}
			if m == "position" {
				table[k+1] = append(table[k+1], fmt.Sprintf("%.2f", v.Position))
			}
		}
	}

	return table
}

// formatText removes the address of the website from the pages, and truncates the first column.
func formatText(table [][]string, charLimit int, address string) [][]string {
	address = strings.TrimSuffix(address, "/")
	// Begins the loop to 1 not to shorten the headers.
	for i := 1; i < len(table); i++ {
		text := table[i][0]
		if address != "" && strings.HasPrefix(text, address) {
			text = strings.TrimPrefix(text, address)
			if text == "" {
				text = "/"
			}
		}

		table[i][0] = truncate(text, charLimit)
	}

	return table
}
