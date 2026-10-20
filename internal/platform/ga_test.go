package platform

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/pkg/errors"
	ga "google.golang.org/api/analyticsdata/v1beta"
)

func Test_Format(t *testing.T) {
	testCases := []struct {
		name        string
		expectedVal []int
		fixtureFile string
		expectedDim []string
		formater    func([]string) string
		wantErr     bool
	}{
		{
			name:        "format users",
			expectedVal: []int{370, 414, 387, 202},
			expectedDim: []string{"02-05", "02-06", "02-07", "02-08"},
			fixtureFile: "./testdata/fixtures/ga4_users.json",
			formater:    func(dim []string) string { return formatTimeLabel("date", dim[0]) },
			wantErr:     false,
		},
		{
			name:        "custom bar metric with aggregation",
			expectedVal: []int{35, 40, 29},
			expectedDim: []string{"03-01", "03-02", "03-03"},
			fixtureFile: "./testdata/fixtures/ga4_bar_metric_aggregations.json",
			formater:    func(dim []string) string { return formatTimeLabel("date", dim[0]) },
			wantErr:     false,
		},
		{
			name:        "custom bar metric country",
			expectedVal: []int{1, 5, 1, 8, 3, 1, 1, 3, 3, 1, 1},
			expectedDim: []string{
				"Belgium",
				"Ukraine",
				"Slovakia",
				"United States",
				"Japan",
				"Netherlands",
				"Sweden",
				"China",
				"Germany",
				"Brazil",
				"France",
			},
			fixtureFile: "./testdata/fixtures/ga4_bar_metric_country.json",
			formater:    func(dim []string) string { return dim[1] },
			wantErr:     false,
		},
		{
			name:        "custom bar metric with pages",
			expectedVal: []int{3, 1, 1, 1, 2, 1, 1, 1, 1},
			expectedDim: []string{"03-20", "03-25", "03-28", "03-30", "03-31", "04-01", "04-03", "04-04", "04-05"},
			fixtureFile: "./testdata/fixtures/ga4_bar_pages.json",
			formater:    func(dim []string) string { return formatTimeLabel("date", dim[0]) },
			wantErr:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ret := &ga.RunReportResponse{}
			fixtures := ReadFixtureFile(tc.fixtureFile, t)
			err := json.Unmarshal(fixtures, ret)
			if err != nil {
				t.Error(err)
			}

			dim, val, err := formatBar(ret.Rows, tc.formater)
			if (err != nil) != tc.wantErr {
				t.Errorf("Error '%v' even if wantErr is %t", err, tc.wantErr)
				return
			}

			if tc.wantErr == false && !reflect.DeepEqual(dim, tc.expectedDim) {
				t.Errorf("Expected %v, actual %v", tc.expectedDim, dim)
			}

			if tc.wantErr == false && !reflect.DeepEqual(val, tc.expectedVal) {
				t.Errorf("Expected %v, actual %v", tc.expectedVal, val)
			}
		})
	}
}

func Test_FormatStackedBar(t *testing.T) {
	testCases := []struct {
		name        string
		values      map[string][]int
		fixtureFile string
		expectedDim []string
		formater    func([]string) string
		wantErr     bool
	}{
		{
			name: "format new vs returning",
			values: map[string][]int{
				newVisitor: {
					11245,
					13966,
					13804,
				},
				returningVisitor: {
					1386,
					1472,
					1633,
				},
			},
			expectedDim: []string{
				"2019-04",
				"2019-05",
				"2019-06",
			},
			fixtureFile: "./testdata/fixtures/ga4_new_returning.json",
			formater:    func(dim []string) string { return formatTimeLabel("yearMonth", dim[0]) },
			wantErr:     false,
		},
		{
			name: "format bar devices",
			values: map[string][]int{
				"desktop": {11309, 10239, 1454},
				"mobile":  {6916, 3050, 359},
				"tablet":  {210, 92, 16},
			},
			expectedDim: []string{
				"2020-02",
				"2020-03",
				"2020-04",
			},
			fixtureFile: "./testdata/fixtures/ga4_bar_devices.json",
			formater:    func(dim []string) string { return formatTimeLabel("yearMonth", dim[0]) },
			wantErr:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := &ga.RunReportResponse{}
			fixtures := ReadFixtureFile(tc.fixtureFile, t)
			err := json.Unmarshal(fixtures, r)
			if err != nil {
				t.Error(err)
			}

			dim, values, err := formatStackedBar(r.Rows, tc.formater)
			if (err != nil) != tc.wantErr {
				t.Errorf("Error '%v' even if wantErr is %t", err, tc.wantErr)
				return
			}

			if tc.wantErr == false && !reflect.DeepEqual(dim, tc.expectedDim) {
				t.Errorf("Expected %v, actual %v", tc.expectedDim, dim)
			}

			if tc.wantErr == false && !reflect.DeepEqual(values, tc.values) {
				t.Errorf("Expected %v, actual %v", tc.values, values)
			}
		})
	}
}

func Test_FormatStackedBarMissingValues(t *testing.T) {
	row := func(date, header, value string) *ga.Row {
		return &ga.Row{
			DimensionValues: []*ga.DimensionValue{{Value: date}, {Value: header}},
			MetricValues:    []*ga.MetricValue{{Value: value}},
		}
	}
	rows := []*ga.Row{
		row("202001", "desktop", "10"),
		row("202001", "mobile", "5"),
		row("202002", "desktop", "7"),
	}

	dim, values, err := formatStackedBar(rows, func(dim []string) string { return formatTimeLabel("yearMonth", dim[0]) })
	if err != nil {
		t.Fatal(err)
	}

	expectedDim := []string{"2020-01", "2020-02"}
	expectedValues := map[string][]int{
		"desktop": {10, 7},
		"mobile":  {5, 0},
	}
	if !reflect.DeepEqual(dim, expectedDim) {
		t.Errorf("Expected %v, actual %v", expectedDim, dim)
	}
	if !reflect.DeepEqual(values, expectedValues) {
		t.Errorf("Expected %v, actual %v", expectedValues, values)
	}
}

func Test_FormatBarReturning(t *testing.T) {
	ret := &ga.RunReportResponse{}
	fixtures := ReadFixtureFile("./testdata/fixtures/ga4_users_returning.json", t)
	if err := json.Unmarshal(fixtures, ret); err != nil {
		t.Fatal(err)
	}

	// The API filters the returning users; simulate it.
	rows := []*ga.Row{}
	for _, r := range ret.Rows {
		if r.DimensionValues[1].Value == returningVisitor {
			rows = append(rows, r)
		}
	}

	dim, val, err := formatBar(rows, func(dim []string) string { return formatTimeLabel("date", dim[0]) })
	if err != nil {
		t.Fatal(err)
	}

	expectedVal := []int{87, 78, 73, 58, 54, 20}
	expectedDim := []string{"03-04", "03-05", "03-06", "03-07", "03-08", "03-09"}
	if !reflect.DeepEqual(dim, expectedDim) {
		t.Errorf("Expected %v, actual %v", expectedDim, dim)
	}
	if !reflect.DeepEqual(val, expectedVal) {
		t.Errorf("Expected %v, actual %v", expectedVal, val)
	}
}

func Test_formatTable(t *testing.T) {
	testCases := []struct {
		name        string
		expectedDim []string
		expectedVal [][]string
		fixtureFile string
		formater    func([]string) string
	}{
		{
			name:        "format traffic sources",
			expectedDim: []string{"google", "(direct)", "indiehackers.com"},
			expectedVal: [][]string{
				{"453", "485", "453", "462"},
				{"132", "180", "132", "155"},
				{"15", "29", "15", "27"},
			},
			fixtureFile: "./testdata/fixtures/ga4_table_traffic_sources.json",
			formater:    func(dim []string) string { return dim[0] },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ret := &ga.RunReportResponse{}
			fixtures := ReadFixtureFile(tc.fixtureFile, t)
			err := json.Unmarshal(fixtures, ret)
			if err != nil {
				t.Error(err)
			}

			dim, val := formatTable(ret.Rows, tc.formater)

			if !reflect.DeepEqual(dim, tc.expectedDim) {
				t.Errorf("Expected %v, actual %v", tc.expectedDim, dim)
			}

			if !reflect.DeepEqual(val, tc.expectedVal) {
				t.Errorf("Expected %v, actual %v", tc.expectedVal, val)
			}
		})
	}
}

func ReadFixtureFile(file string, t *testing.T) (data []byte) {
	data, err := os.ReadFile(file)
	if err != nil {
		t.Error(errors.Wrapf(err, "can't read file %s", file))
	}

	return data
}

func Test_mapMetrics(t *testing.T) {
	testCases := []struct {
		name     string
		m        []string
		expected []*ga.Metric
		wantErr  bool
	}{
		{
			name: "aliases",
			m: []string{
				"sessions",
				"page_views",
				"users",
				"bounces",
			},
			expected: []*ga.Metric{
				{Name: "sessions"},
				{Name: "screenPageViews"},
				{Name: "totalUsers"},
				{Name: "bounces", Expression: "sessions-engagedSessions"},
			},
		},
		{
			name: "universal analytics and GA4 names",
			m: []string{
				"ga:sessionsPerUser",
				"ga:avgSessionDuration",
				" eventCount",
			},
			expected: []*ga.Metric{
				{Name: "sessionsPerUser"},
				{Name: "averageSessionDuration"},
				{Name: "eventCount"},
			},
		},
		{
			name:    "removed metric",
			m:       []string{"sessions", "unique_page_views"},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := mapMetrics(tc.m)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Error '%v' even if wantErr is %t", err, tc.wantErr)
			}
			if !tc.wantErr && !reflect.DeepEqual(tc.expected, actual) {
				t.Errorf("Expected %v, actual %v", tc.expected, actual)
			}
		})
	}
}

func Test_mapDimension(t *testing.T) {
	testCases := []struct {
		name     string
		d        string
		expected string
	}{
		{
			name:     "alias",
			d:        "page_path",
			expected: "pagePath",
		},
		{
			name:     "universal analytics name",
			d:        "ga:pagePath",
			expected: "pagePath",
		},
		{
			name:     "renamed universal analytics name",
			d:        "ga:userType",
			expected: "newVsReturning",
		},
		{
			name:     "GA4 name",
			d:        "sessionDefaultChannelGroup",
			expected: "sessionDefaultChannelGroup",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := mapDimension(tc.d)
			if !reflect.DeepEqual(tc.expected, actual) {
				t.Errorf("Expected %v, actual %v", tc.expected, actual)
			}
		})
	}
}

func Test_mapHeaders(t *testing.T) {
	testCases := []struct {
		name     string
		m        []string
		expected []string
		el       string
	}{
		{
			name: "happy case",
			el:   "Pages",
			expected: []string{
				"Pages",
				"Sessions",
				"Page Views",
				"Users",
				"someRandomExpr",
				"hey",
			},
			m: []string{
				"sessions",
				"page_views",
				"users",
				"ga:someRandomExpr",
				" hey",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := mapHeaders(tc.el, tc.m)

			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("Expected %v, actual %v", tc.expected, actual)
			}
		})
	}
}

func Test_mapOrderBy(t *testing.T) {
	testCases := []struct {
		name       string
		m          []string
		dimensions []*ga.Dimension
		expected   []*ga.OrderBy
	}{
		{
			name: "happy case",
			m: []string{
				"sessions asc",
				"page_views desc",
				"users",
				"ga:pageviews",
				"page_path asc",
			},
			dimensions: []*ga.Dimension{{Name: "pagePath"}},
			expected: []*ga.OrderBy{
				{Metric: &ga.MetricOrderBy{MetricName: "sessions"}},
				{Metric: &ga.MetricOrderBy{MetricName: "screenPageViews"}, Desc: true},
				{Metric: &ga.MetricOrderBy{MetricName: "totalUsers"}, Desc: true},
				{Metric: &ga.MetricOrderBy{MetricName: "screenPageViews"}, Desc: true},
				{Dimension: &ga.DimensionOrderBy{DimensionName: "pagePath"}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := mapOrderBy(tc.m, tc.dimensions)
			if !reflect.DeepEqual(tc.expected, actual) {
				t.Errorf("Expected %v, actual %v", tc.expected, actual)
			}
		})
	}
}

func Test_formatTimeLabel(t *testing.T) {
	testCases := []struct {
		dimension string
		value     string
		expected  string
	}{
		{"date", "20261020", "10-20"},
		{"yearMonth", "202610", "2026-10"},
		{"year", "2026", "2026"},
		{"date", "(other)", "(other)"},
	}

	for _, tc := range testCases {
		t.Run(tc.dimension+"/"+tc.value, func(t *testing.T) {
			if actual := formatTimeLabel(tc.dimension, tc.value); actual != tc.expected {
				t.Errorf("Expected %v, actual %v", tc.expected, actual)
			}
		})
	}
}

func Test_formatValue(t *testing.T) {
	testCases := map[string]string{
		"42":                 "42",
		"0.5312":             "0.53",
		"123.45678":          "123.46",
		"not a number. ever": "not a number. ever",
	}

	for in, expected := range testCases {
		if actual := formatValue(in); actual != expected {
			t.Errorf("formatValue(%q): expected %v, actual %v", in, expected, actual)
		}
	}
}

func Test_containsFilter(t *testing.T) {
	if f := containsFilter("pagePath", []string{" ", ""}); f != nil {
		t.Errorf("Expected nil filter for empty expressions, actual %v", f)
	}

	single := containsFilter("pagePath", []string{" /blog "})
	if single.Filter == nil || single.Filter.StringFilter.Value != "/blog" {
		t.Errorf("Expected a single filter on /blog, actual %v", single)
	}

	multiple := containsFilter("pagePath", []string{"/a", "/b"})
	if multiple.OrGroup == nil || len(multiple.OrGroup.Expressions) != 2 {
		t.Errorf("Expected an OR group of 2 filters, actual %v", multiple)
	}

	excluded := containsFilter("pagePath", []string{"-/tag/"})
	if excluded.NotExpression == nil || excluded.NotExpression.Filter.StringFilter.Value != "/tag/" {
		t.Errorf("Expected a NOT filter on /tag/, actual %v", excluded)
	}

	// Every excluded value and at least one of the included values.
	mixed := containsFilter("pagePath", []string{"/a", "-/tag/", "/b", "-/draft/"})
	if mixed.AndGroup == nil || len(mixed.AndGroup.Expressions) != 3 {
		t.Fatalf("Expected an AND group of 3 filters, actual %v", mixed)
	}
	e := mixed.AndGroup.Expressions
	if e[0].NotExpression == nil || e[1].NotExpression == nil || e[2].OrGroup == nil || len(e[2].OrGroup.Expressions) != 2 {
		t.Errorf("Expected 2 NOT filters and an OR group of 2 filters, actual %v", mixed)
	}

	if f := containsFilter("pagePath", []string{"-", " - "}); f != nil {
		t.Errorf("Expected nil filter for empty exclusions, actual %v", f)
	}
}

func Test_propertyName(t *testing.T) {
	for in, expected := range map[string]string{
		"123456":            "properties/123456",
		" 123456 ":          "properties/123456",
		"properties/123456": "properties/123456",
	} {
		if actual := propertyName(in); actual != expected {
			t.Errorf("propertyName(%q): expected %v, actual %v", in, expected, actual)
		}
	}
}
