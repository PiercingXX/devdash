package platform

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"errors"

	ga "google.golang.org/api/analyticsdata/v1beta"
)

// To create fixtures in order to test the data formatting.
// j, _ := json.Marshal(resp)
// fmt.Println(string(j))

const (
	// Determine if the time will be display as x-axis headers, or something else
	XHeaderTime uint16 = iota
	XHeaderOtherDim

	earliestDate     = "2015-08-14" // This is the earliest date the Google Analytics Data API accepts.
	newVisitor       = "new"
	returningVisitor = "returning"
)

// mappingMetrics maps devdash aliases to GA4 metrics.
// GA4 has no equivalent for some Universal Analytics metrics; they're derived with an expression when possible.
var mappingMetrics = map[string]ga.Metric{
	"sessions":                 {Name: "sessions"},
	"engaged_sessions":         {Name: "engagedSessions"},
	"page_views":               {Name: "screenPageViews"},
	"bounces":                  {Name: "bounces", Expression: "sessions-engagedSessions"},
	"users":                    {Name: "totalUsers"},
	"new_users":                {Name: "newUsers"},
	"active_users":             {Name: "activeUsers"},
	"session_duration":         {Name: "sessionDuration", Expression: "averageSessionDuration*sessions"},
	"average_session_duration": {Name: "averageSessionDuration"},
	"bounce_rate":              {Name: "bounceRate"},
	"engagement_rate":          {Name: "engagementRate"},
}

// removedMetrics existed in Universal Analytics but have no GA4 equivalent.
var removedMetrics = map[string]bool{
	"entrances":         true,
	"unique_page_views": true,
	"uniquePageviews":   true,
}

var mappingTimePeriod = map[string]string{
	"day":   "date",
	"month": "yearMonth",
	"year":  "year",
}

var mappingDimensions = map[string]string{
	"page_path":       "pagePath",
	"landing_page":    "landingPage",
	"referrer":        "pageReferrer",
	"traffic_source":  "sessionSource",
	"user_type":       "newVsReturning",
	"device_category": "deviceCategory",
	"country":         "country",
}

// mappingUANames maps Universal Analytics names (without the "ga:" prefix) to their GA4 equivalent,
// to keep old configurations working when possible.
var mappingUANames = map[string]string{
	"pageViews":          "screenPageViews",
	"pageviews":          "screenPageViews",
	"users":              "totalUsers",
	"avgSessionDuration": "averageSessionDuration",
	"source":             "sessionSource",
	"userType":           "newVsReturning",
	"fullReferrer":       "pageReferrer",
}

var mappingHeader = map[string]string{
	"sessions":                 "Sessions",
	"engaged_sessions":         "Engaged Sessions",
	"page_views":               "Page Views",
	"bounces":                  "Bounces",
	"users":                    "Users",
	"new_users":                "New Users",
	"active_users":             "Active Users",
	"average_session_duration": "Avg. Session Duration",
	"bounce_rate":              "Bounce Rate",
	"engagement_rate":          "Engagement Rate",
}

// Analytics connect to Google Analytics Data API (GA4).
type Analytics struct {
	service *ga.Service
}

// AnalyticValues which can be possibly send to the Google Analytics API.
// This is a pure value object without behavior.
type AnalyticValues struct {
	PropertyID string
	StartDate  string
	EndDate    string
	TimePeriod string
	Global     bool
	Metrics    []string
	Dimensions []string
	Filters    []string
	Orders     []string
	RowLimit   int64
	XHeaders   uint16
}

// NewAnalyticsClient to connect to Google Analytics APIs.
func NewAnalyticsClient(keyfile string) (*Analytics, error) {
	opts, err := googleClientOptions(keyfile, ga.AnalyticsReadonlyScope)
	if err != nil {
		return nil, err
	}

	service, err := ga.NewService(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("creating the analytics data service failed: %w", err)
	}

	return &Analytics{service: service}, nil
}

// SimpleMetric get a value depending on Google Analytics metrics.
func (c *Analytics) SimpleMetric(val AnalyticValues) (string, error) {
	metrics, err := mapMetrics(val.Metrics)
	if err != nil {
		return "", err
	}

	req := &ga.RunReportRequest{
		DateRanges: dateRanges(val),
		Metrics:    metrics,
	}

	resp, err := c.service.Properties.RunReport(propertyName(val.PropertyID), req).Do()
	if err != nil {
		return "", fmt.Errorf("can't get total metric data from Google Analytics from %s to %s: %w", val.StartDate, val.EndDate, googleError(err))
	}

	if len(resp.Rows) != 0 && len(resp.Rows[0].MetricValues) != 0 {
		return formatValue(resp.Rows[0].MetricValues[0].Value), nil
	}

	return "0", nil
}

// BarMetric provides a qualitive dimension linked to a quantitative value, for example a date (dimension) with an int.
func (c *Analytics) BarMetric(val AnalyticValues) ([]string, []int, error) {
	metrics, err := mapMetrics(val.Metrics)
	if err != nil {
		return nil, nil, err
	}

	// The time dimension is always the first one.
	tm := mapTimePeriod(val.TimePeriod)
	dim := append([]*ga.Dimension{{Name: tm}}, mapDimensions(val.Dimensions)...)

	req := &ga.RunReportRequest{
		DateRanges: dateRanges(val),
		Metrics:    metrics,
		Dimensions: dim,
		OrderBys: []*ga.OrderBy{
			{Dimension: &ga.DimensionOrderBy{DimensionName: tm}},
		},
		KeepEmptyRows: true,
	}

	// The filters apply to the dimension: they are ambiguous with more than one.
	if len(val.Dimensions) == 1 && len(val.Filters) != 0 {
		req.DimensionFilter = containsFilter(mapDimension(val.Dimensions[0]), val.Filters)
	}

	resp, err := c.service.Properties.RunReport(propertyName(val.PropertyID), req).Do()
	if err != nil {
		return nil, nil, fmt.Errorf("can't get bar data from Google Analytics from %s to %s: %w", val.StartDate, val.EndDate, googleError(err))
	}

	// Decide of the header X-axis
	f := func(dimValues []string) string {
		if val.XHeaders == XHeaderOtherDim && len(dimValues) > 1 {
			return dimValues[1]
		}

		return formatTimeLabel(tm, dimValues[0])
	}

	return formatBar(resp.Rows, f)
}

// RealTimeUsers return the number of visitor currently on the website.
func (c *Analytics) RealTimeUsers(propertyID string) (string, error) {
	req := &ga.RunRealtimeReportRequest{
		Metrics: []*ga.Metric{{Name: "activeUsers"}},
	}

	resp, err := c.service.Properties.RunRealtimeReport(propertyName(propertyID), req).Do()
	if err != nil {
		return "", fmt.Errorf("can't get real time data from Google Analytics: %w", googleError(err))
	}

	if len(resp.Rows) != 0 && len(resp.Rows[0].MetricValues) != 0 {
		return resp.Rows[0].MetricValues[0].Value, nil
	}

	return "0", nil
}

// Table display dimensions and values.
// The headers on the first row are qualitative dimensions, the values can be qualitative or quantitative.
func (c *Analytics) Table(
	an AnalyticValues,
	firstHeader string,
) (headers []string, dim []string, u [][]string, err error) {
	metrics, err := mapMetrics(an.Metrics)
	if err != nil {
		return nil, nil, nil, err
	}

	dimensions := mapDimensions(an.Dimensions)
	req := &ga.RunReportRequest{
		DateRanges:    dateRanges(an),
		Metrics:       metrics,
		Dimensions:    dimensions,
		OrderBys:      mapOrderBy(an.Orders, dimensions),
		KeepEmptyRows: true,
		Limit:         an.RowLimit,
	}

	if len(an.Filters) > 0 {
		// The same filters apply to every dimension.
		filters := []*ga.FilterExpression{}
		for _, v := range dimensions {
			if f := containsFilter(v.Name, an.Filters); f != nil {
				filters = append(filters, f)
			}
		}

		if len(filters) > 0 {
			req.DimensionFilter = &ga.FilterExpression{
				AndGroup: &ga.FilterExpressionList{Expressions: filters},
			}
		}
	}

	resp, err := c.service.Properties.RunReport(propertyName(an.PropertyID), req).Do()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("can't get table data from Google Analytics from %s to %s: %w", an.StartDate, an.EndDate, googleError(err))
	}

	formater := func(dim []string) string {
		return dim[0]
	}

	headers = mapHeaders(firstHeader, an.Metrics)
	dim, u = formatTable(resp.Rows, formater)
	return
}

// StackedBar returns one dimension set linked with multiple values.
func (c *Analytics) StackedBar(an AnalyticValues) (dim []string, values map[string][]int, err error) {
	metrics, err := mapMetrics(an.Metrics)
	if err != nil {
		return nil, nil, err
	}

	// The time dimension is always the first one, the stacked dimension the second one.
	tm := mapTimePeriod(an.TimePeriod)
	d := append([]*ga.Dimension{{Name: tm}}, mapDimensions(an.Dimensions)...)

	req := &ga.RunReportRequest{
		DateRanges: dateRanges(an),
		Metrics:    metrics,
		Dimensions: d,
		OrderBys: []*ga.OrderBy{
			{Dimension: &ga.DimensionOrderBy{DimensionName: tm}},
		},
		KeepEmptyRows: true,
	}

	resp, err := c.service.Properties.RunReport(propertyName(an.PropertyID), req).Do()
	if err != nil {
		return nil, nil, fmt.Errorf("can't get stacked bar data from Google Analytics from %s to %s: %w", an.StartDate, an.EndDate, googleError(err))
	}

	// Always dates on x-axis
	formater := func(dim []string) string {
		return formatTimeLabel(tm, dim[0])
	}

	return formatStackedBar(resp.Rows, formater)
}

// formatBar to return one slice of dimension which elements are all linked with the elements of another slice with the values.
func formatBar(rows []*ga.Row, dimFormater func(dim []string) string) (dim []string, u []int, err error) {
	dimVal := map[string]int{}
	for _, r := range rows {
		d := dimFormater(dimensionValues(r))

		// Add the dimension only if it was not added already.
		if _, ok := dimVal[d]; !ok {
			dim = append(dim, d)
			dimVal[d] = 0
		}

		for _, m := range r.MetricValues {
			vu, err := parseMetricValue(m.Value)
			if err != nil {
				return nil, nil, err
			}
			dimVal[d] += vu
		}
	}

	for _, v := range dim {
		u = append(u, dimVal[v])
	}

	return dim, u, nil
}

func formatTable(
	rows []*ga.Row,
	dimFormater func(dim []string) string,
) (dim []string, u [][]string) {
	for _, r := range rows {
		dim = append(dim, dimFormater(dimensionValues(r)))

		var g []string
		for _, m := range r.MetricValues {
			g = append(g, formatValue(m.Value))
		}
		u = append(u, g)
	}

	return dim, u
}

// formatStackedBar returns the x-axis dimensions and, for every value of the stacked dimension,
// the values aligned with the x-axis. A missing combination is filled with 0.
func formatStackedBar(
	rows []*ga.Row,
	dimFormater func(dim []string) string,
) (dim []string, values map[string][]int, err error) {
	type point struct {
		dim    string
		header string
		value  int
	}

	points := []point{}
	for _, r := range rows {
		dv := dimensionValues(r)
		if len(dv) < 2 {
			return nil, nil, errors.New("a stacked bar needs a dimension to stack")
		}

		d := dimFormater(dv)
		if !inArray(dim, d) {
			dim = append(dim, d)
		}

		for _, m := range r.MetricValues {
			vu, err := parseMetricValue(m.Value)
			if err != nil {
				return nil, nil, err
			}
			points = append(points, point{dim: d, header: dv[1], value: vu})
		}
	}

	values = make(map[string][]int)
	for _, p := range points {
		if _, ok := values[p.header]; !ok {
			values[p.header] = make([]int, len(dim))
		}
		values[p.header][slices.Index(dim, p.dim)] += p.value
	}

	return dim, values, nil
}

func dimensionValues(r *ga.Row) []string {
	d := make([]string, len(r.DimensionValues))
	for k, v := range r.DimensionValues {
		d[k] = v.Value
	}

	return d
}

// parseMetricValue parses a metric value returned by the API. Decimal values are truncated.
func parseMetricValue(value string) (int, error) {
	if strings.Contains(value, ".") {
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, err
		}
		return int(f), nil
	}

	vu, err := strconv.ParseInt(value, 10, 0)
	if err != nil {
		return 0, err
	}

	return int(vu), nil
}

// formatValue rounds decimal values to two digits.
func formatValue(value string) string {
	if !strings.Contains(value, ".") {
		return value
	}

	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return value
	}

	return strconv.FormatFloat(f, 'f', 2, 64)
}

// formatTimeLabel formats the value of a GA4 time dimension for the x-axis of a chart.
func formatTimeLabel(dimension string, value string) string {
	switch {
	case dimension == "date" && len(value) == 8:
		// YYYYMMDD -> MM-DD
		return value[4:6] + "-" + value[6:]
	case dimension == "yearMonth" && len(value) == 6:
		// YYYYMM -> YYYY-MM
		return value[:4] + "-" + value[4:]
	}

	return value
}

func inArray(haystack []string, needle string) bool {
	return slices.Contains(haystack, needle)
}

func propertyName(propertyID string) string {
	propertyID = strings.TrimSpace(propertyID)
	if strings.HasPrefix(propertyID, "properties/") {
		return propertyID
	}

	return "properties/" + propertyID
}

func dateRanges(val AnalyticValues) []*ga.DateRange {
	startDate := val.StartDate
	if val.Global {
		startDate = earliestDate
	}

	return []*ga.DateRange{
		{StartDate: startDate, EndDate: val.EndDate},
	}
}

// containsFilter matches a dimension containing at least one of the expressions,
// and none of the expressions prefixed with "-".
func containsFilter(dimension string, expressions []string) *ga.FilterExpression {
	contains := func(v string) *ga.FilterExpression {
		return &ga.FilterExpression{
			Filter: &ga.Filter{
				FieldName: dimension,
				StringFilter: &ga.StringFilter{
					MatchType: "CONTAINS",
					Value:     v,
				},
			},
		}
	}

	included, excluded := []*ga.FilterExpression{}, []*ga.FilterExpression{}
	for _, v := range expressions {
		v = strings.TrimSpace(v)
		if after, ok := strings.CutPrefix(v, "-"); ok {
			if after = strings.TrimSpace(after); after != "" {
				excluded = append(excluded, &ga.FilterExpression{NotExpression: contains(after)})
			}
			continue
		}
		if v != "" {
			included = append(included, contains(v))
		}
	}

	group := excluded
	switch len(included) {
	case 0:
	case 1:
		group = append(group, included[0])
	default:
		group = append(group, &ga.FilterExpression{OrGroup: &ga.FilterExpressionList{Expressions: included}})
	}

	switch len(group) {
	case 0:
		return nil
	case 1:
		return group[0]
	}

	return &ga.FilterExpression{AndGroup: &ga.FilterExpressionList{Expressions: group}}
}

// The map functions map the properties of the application to the Google Analytics API params.

// normalizeName removes the Universal Analytics prefix "ga:" and maps the old names to GA4 when possible.
func normalizeName(name string) string {
	name = strings.TrimPrefix(strings.TrimSpace(name), "ga:")
	if n, ok := mappingUANames[name]; ok {
		return n
	}

	return name
}

func mapMetrics(m []string) ([]*ga.Metric, error) {
	gam := make([]*ga.Metric, len(m))

	for k, v := range m {
		metric, err := mapMetric(v)
		if err != nil {
			return nil, err
		}
		gam[k] = metric
	}

	return gam, nil
}

// mapMetric will first try to search an alias for a google analytics metric,
// send the option to GoogleAnalytics otherwise
func mapMetric(metric string) (*ga.Metric, error) {
	metric = strings.TrimSpace(metric)
	if m, ok := mappingMetrics[metric]; ok {
		return &m, nil
	}

	name := normalizeName(metric)
	if removedMetrics[name] {
		return nil, fmt.Errorf("the metric %s doesn't exist in Google Analytics 4", metric)
	}

	return &ga.Metric{Name: name}, nil
}

func metricName(metric string) string {
	m, err := mapMetric(metric)
	if err != nil {
		return normalizeName(metric)
	}

	return m.Name
}

func mapTimePeriod(m string) string {
	m = strings.TrimSpace(m)
	if timePeriod, ok := mappingTimePeriod[m]; ok {
		return timePeriod
	}

	return normalizeName(m)
}

func mapDimension(dim string) string {
	dim = strings.TrimSpace(dim)
	if d, ok := mappingDimensions[dim]; ok {
		return d
	}

	return normalizeName(dim)
}

func mapDimensions(dimensions []string) []*ga.Dimension {
	d := []*ga.Dimension{}
	for _, v := range dimensions {
		d = append(d, &ga.Dimension{Name: mapDimension(v)})
	}

	return d
}

// mapOrderBy maps orders like "sessions desc". A field matching one of the dimensions orders the dimension;
// the field is considered as a metric otherwise.
func mapOrderBy(o []string, dimensions []*ga.Dimension) []*ga.OrderBy {
	gam := make([]*ga.OrderBy, len(o))

	for k, v := range o {
		s := strings.Fields(v)

		// default
		field := "sessions"
		order := "desc"

		if len(s) > 0 {
			field = s[0]
		}

		if len(s) > 1 {
			order = strings.ToLower(s[1])
		}

		desc := order != "asc"

		dim := mapDimension(field)
		if slices.ContainsFunc(dimensions, func(d *ga.Dimension) bool { return d.Name == dim }) {
			gam[k] = &ga.OrderBy{Dimension: &ga.DimensionOrderBy{DimensionName: dim}, Desc: desc}
			continue
		}

		gam[k] = &ga.OrderBy{Metric: &ga.MetricOrderBy{MetricName: metricName(field)}, Desc: desc}
	}

	return gam
}

func mapHeaders(el string, metrics []string) []string {
	h := make([]string, len(metrics)+1)
	h[0] = el

	for k, v := range metrics {
		head, ok := mappingHeader[strings.TrimSpace(v)]
		if !ok {
			h[k+1] = strings.TrimPrefix(strings.TrimSpace(v), "ga:")
			continue
		}
		h[k+1] = head
	}

	return h
}
