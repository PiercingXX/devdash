package internal

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/PiercingXX/devdash/internal/platform"
)

const (
	jsonBox   = "json.box"
	jsonGauge = "json.gauge"
	jsonBar   = "json.bar"
	jsonTable = "json.table"

	optionURL     = "url"
	optionPath    = "path"
	optionHeader  = "header"
	optionColumns = "columns"
	optionLabel   = "label"
	optionValue   = "value"
)

// jsonWidget displays values of the JSON returned by any URL. It doesn't need any service configuration.
type jsonWidget struct {
	tui *Tui
}

func NewJSONWidget() *jsonWidget {
	return &jsonWidget{}
}

func (j jsonWidget) CreateWidgets(widget Widget, tui *Tui) (f func() error, err error) {
	j.tui = tui

	switch widget.Name {
	case jsonBox, jsonGauge, jsonBar, jsonTable:
	default:
		return nil, fmt.Errorf("can't find the widget %s for the service json", widget.Name)
	}

	url := widget.Options[optionURL]
	if url == "" {
		return nil, fmt.Errorf("the widget %s needs an url", widget.Name)
	}

	root, err := platform.FetchJSON(url, widget.Options[optionHeader])
	if err != nil {
		return nil, err
	}

	title := " " + url + " "
	if t, ok := widget.Options[optionTitle]; ok {
		title = t
	}
	path := widget.Options[optionPath]

	switch widget.Name {
	case jsonBox:
		v, err := platform.JSONValue(root, path)
		if err != nil {
			return nil, err
		}
		return func() error { return tui.AddTextBox(v, title, widget.Options) }, nil

	case jsonGauge:
		v, err := platform.JSONValue(root, path)
		if err != nil {
			return nil, err
		}
		percent, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("the value %q at the path %q isn't a number", v, path)
		}
		return func() error { return tui.AddGauge(percent, title, widget.Options) }, nil

	case jsonBar:
		labels, values, err := platform.JSONBars(root, path, widget.Options[optionLabel], widget.Options[optionValue])
		if err != nil {
			return nil, err
		}
		return func() error { return tui.AddBarChart(values, labels, title, widget.Options) }, nil

	case jsonTable:
		columns := strings.Split(widget.Options[optionColumns], ",")
		if widget.Options[optionColumns] == "" {
			return nil, errors.New("the widget json.table needs the fields to display in columns")
		}

		limit := 5
		if l, ok := widget.Options[optionRowLimit]; ok {
			if limit, err = strconv.Atoi(l); err != nil {
				return nil, fmt.Errorf("%s must be a number: %w", l, err)
			}
		}

		rows, err := platform.JSONTable(root, path, columns, limit)
		if err != nil {
			return nil, err
		}

		headers := columns
		if h := widget.Options[optionHeaders]; h != "" {
			headers = strings.Split(h, ",")
		}
		return func() error { return tui.AddTable(append([][]string{headers}, rows...), title, widget.Options) }, nil
	}

	return nil, nil
}
