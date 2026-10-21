package platform

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// fakeGoogle is a fake of the Google APIs: it gives a token, records the requests, and answers with the response of
// each path.
type fakeGoogle struct {
	*httptest.Server
	mu        sync.Mutex
	requests  map[string][]map[string]any
	responses map[string]string
}

func newFakeGoogle(t *testing.T, responses map[string]string) *fakeGoogle {
	f := &fakeGoogle{requests: map[string][]map[string]any{}, responses: responses}
	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			_, _ = io.WriteString(w, `{"access_token": "fake", "token_type": "Bearer", "expires_in": 3600}`)
			return
		}

		if r.Header.Get("Authorization") != "Bearer fake" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		body := map[string]any{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.mu.Lock()
		f.requests[r.URL.Path] = append(f.requests[r.URL.Path], body)
		f.mu.Unlock()

		resp, ok := f.responses[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			_, _ = io.WriteString(w, `{"error": {"code": 403, "message": "User does not have sufficient permissions for this property."}}`)
			return
		}
		_, _ = io.WriteString(w, resp)
	}))
	t.Cleanup(f.Close)
	t.Setenv(googleEndpointEnv, f.URL+"/")

	return f
}

// keyfile of a fake service account, which gets its token from the fake API.
func (f *fakeGoogle) keyfile(t *testing.T) string {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	pk := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	content, _ := json.Marshal(map[string]string{
		"type":         "service_account",
		"private_key":  string(pk),
		"client_email": "devdash@test.iam.gserviceaccount.com",
		"token_uri":    f.URL + "/token",
	})

	path := filepath.Join(t.TempDir(), "keyfile.json")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

func Test_analyticsWithFakeAPI(t *testing.T) {
	f := newFakeGoogle(t, map[string]string{
		"/v1beta/properties/123:runReport": `{
			"dimensionHeaders": [{"name": "date"}],
			"metricHeaders": [{"name": "totalUsers"}, {"name": "engagementRate"}],
			"rows": [
				{"dimensionValues": [{"value": "20261019"}], "metricValues": [{"value": "120"}, {"value": "0.5"}]},
				{"dimensionValues": [{"value": "20261020"}], "metricValues": [{"value": "80"}, {"value": "0.25"}]}
			]
		}`,
		"/v1beta/properties/123:runRealtimeReport": `{"rows": [{"metricValues": [{"value": "17"}]}]}`,
	})

	a, err := NewAnalyticsClient(f.keyfile(t))
	if err != nil {
		t.Fatal(err)
	}

	val := AnalyticValues{PropertyID: "123", StartDate: "2026-10-19", EndDate: "2026-10-20", TimePeriod: "day", Metrics: []string{"users", "engagement_rate"}}
	dim, values, err := a.BarMetric(val)
	if err != nil {
		t.Fatal(err)
	}
	// The values of the metrics are added: 120 users + 50%, 80 users + 25%.
	if !reflect.DeepEqual(dim, []string{"10-19", "10-20"}) || !reflect.DeepEqual(values, []int{170, 105}) {
		t.Errorf("unexpected bars %v %v", dim, values)
	}

	req := f.requests["/v1beta/properties/123:runReport"][0]
	metrics, _ := json.Marshal(req["metrics"])
	if string(metrics) != `[{"name":"totalUsers"},{"name":"engagementRate"}]` {
		t.Errorf("unexpected metrics in the request %s", metrics)
	}

	_, _, rows, err := a.Table(AnalyticValues{PropertyID: "123", StartDate: "2026-10-19", EndDate: "2026-10-20", Metrics: []string{"users", "engagement_rate"}, Dimensions: []string{"date"}, Orders: []string{"users desc"}}, "Date")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rows[0], []string{"120", "50.00%"}) {
		t.Errorf("expected the rate in percent, actual %v", rows)
	}

	users, err := a.RealTimeUsers("123")
	if err != nil || users != "17" {
		t.Errorf("expected 17 real time users, actual %q, %v", users, err)
	}

	_, err = a.SimpleMetric(AnalyticValues{PropertyID: "456", StartDate: "2026-10-19", EndDate: "2026-10-20", Metrics: []string{"sessions"}})
	if err == nil || !strings.Contains(err.Error(), "error 403 of the Google API: User does not have sufficient permissions") {
		t.Errorf("expected a short error, actual %v", err)
	}
}

func Test_searchConsoleWithFakeAPI(t *testing.T) {
	f := newFakeGoogle(t, map[string]string{
		"/webmasters/v3/sites/https:%2F%2Fexample.com%2F/searchAnalytics/query": `{"rows": [
			{"keys": ["terminal dashboard"], "clicks": 10, "impressions": 200, "ctr": 0.05, "position": 3.2}
		]}`,
	})
	// The site in the path is escaped.
	f.responses["/webmasters/v3/sites/https://example.com//searchAnalytics/query"] = f.responses["/webmasters/v3/sites/https:%2F%2Fexample.com%2F/searchAnalytics/query"]

	s, err := NewSearchConsoleClient(f.keyfile(t))
	if err != nil {
		t.Fatal(err)
	}

	rows, err := s.Table("2026-10-01", "2026-10-31", 5, "https://example.com/", "query", "-devdash,*country fra")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Dimension != "terminal dashboard" || rows[0].Clicks != 10 {
		t.Errorf("unexpected rows %+v", rows)
	}

	var req map[string]any
	for _, r := range f.requests {
		req = r[0]
	}
	filters, _ := json.Marshal(req["dimensionFilterGroups"])
	expected := `[{"filters":[{"dimension":"query","expression":"devdash","operator":"notContains"},{"dimension":"country","expression":"fra","operator":"contains"}]}]`
	if string(filters) != expected {
		t.Errorf("expected the filters %s, actual %s", expected, filters)
	}
}
