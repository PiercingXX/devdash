// Google Search Console
package platform

import (
	"context"
	"fmt"
	"strings"

	sc "google.golang.org/api/searchconsole/v1"
)

// SearchConsole connect to the Google Search Console API.
type SearchConsole struct {
	service *sc.Service
}

// SearchConsoleResponse returned after requesting the API.
type SearchConsoleResponse struct {
	Dimension   string
	Clicks      float64
	Impressions float64
	Ctr         float64
	Position    float64
}

// NewSearchConsoleClient create a SearchConsole.
func NewSearchConsoleClient(keyfile string) (*SearchConsole, error) {
	opts, err := googleClientOptions(keyfile, sc.WebmastersReadonlyScope)
	if err != nil {
		return nil, err
	}

	service, err := sc.NewService(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("can't get search console service: %w", err)
	}

	return &SearchConsole{service: service}, nil
}

// Table of Google Search Console with a dimension and its values.
func (w *SearchConsole) Table(
	startDate string,
	endDate string,
	limit int64,
	address string,
	dimension string,
	filters string,
) ([]SearchConsoleResponse, error) {
	req := &sc.SearchAnalyticsQueryRequest{
		StartDate:  startDate,
		EndDate:    endDate,
		Dimensions: []string{dimension},
		DimensionFilterGroups: []*sc.ApiDimensionFilterGroup{
			{
				Filters: filtersFromString(filters, dimension),
			},
		},
		RowLimit: limit,
	}

	resp, err := w.service.Searchanalytics.Query(address, req).Do()
	if err != nil {
		return nil, fmt.Errorf("can't get data from Google Search Console for %s: %w", address, googleError(err))
	}

	return formatSearchTable(resp.Rows), nil
}

// filtersFromString declared in the config.
func filtersFromString(filters string, dimension string) []*sc.ApiDimensionFilter {
	fg := []*sc.ApiDimensionFilter{}
	if filters != "" {
		f := strings.SplitSeq(filters, ",")
		for v := range f {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			dim := dimension
			if strings.Contains(v, " ") {
				t := strings.Split(v, " ")
				if strings.Contains(t[0], "*") {
					dim = strings.Trim(t[0], "*")
					v = strings.Join(t[1:], " ")
				}
			}

			operator := "contains"
			if string(v[0]) == "-" {
				operator = "notContains"
				v = v[1:]
			}

			fg = append(fg, &sc.ApiDimensionFilter{
				Dimension:  dim,
				Expression: v,
				Operator:   operator,
			})
		}
	}

	return fg
}

func formatSearchTable(rows []*sc.ApiDataRow) []SearchConsoleResponse {
	results := []SearchConsoleResponse{}
	for _, v := range rows {
		results = append(results, SearchConsoleResponse{
			Dimension:   strings.Join(v.Keys, ","),
			Clicks:      v.Clicks,
			Impressions: v.Impressions,
			Ctr:         v.Ctr,
			Position:    v.Position,
		})
	}

	return results
}
