package platform

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func Test_FetchJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth":
			if r.Header.Get("Authorization") != "Bearer secret" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			fmt.Fprint(w, `{"ok": true}`)
		case "/html":
			fmt.Fprint(w, "<html></html>")
		default:
			fmt.Fprint(w, `{"total": 42, "rate": 73.5, "items": [{"name": "a", "count": 3}, {"name": "b", "count": 5.9}]}`)
		}
	}))
	defer srv.Close()

	t.Setenv("DEVDASH_TEST_TOKEN", "secret")
	if _, err := FetchJSON(srv.URL+"/auth", "Authorization: Bearer ${DEVDASH_TEST_TOKEN}"); err != nil {
		t.Errorf("expected the header with the environment variable, actual %v", err)
	}
	if _, err := FetchJSON(srv.URL+"/auth", ""); err == nil || !strings.Contains(err.Error(), "401") {
		t.Errorf("expected an error with the status, actual %v", err)
	}
	if _, err := FetchJSON(srv.URL+"/html", ""); err == nil {
		t.Error("expected an error for invalid JSON")
	}
	if _, err := FetchJSON(srv.URL, "no colon"); err == nil {
		t.Error("expected an error for an invalid header")
	}

	root, err := FetchJSON(srv.URL, "")
	if err != nil {
		t.Fatal(err)
	}

	if v, err := JSONValue(root, "total"); err != nil || v != "42" {
		t.Errorf("expected 42, actual %q, %v", v, err)
	}
	if _, err := JSONValue(root, "missing"); err == nil {
		t.Error("expected an error for a missing path")
	}

	table, err := JSONTable(root, "items", []string{"name", " count"}, 1)
	if err != nil || !reflect.DeepEqual(table, [][]string{{"a", "3"}}) {
		t.Errorf("expected one row, actual %v, %v", table, err)
	}

	labels, values, err := JSONBars(root, "items", "name", "count")
	if err != nil || !reflect.DeepEqual(labels, []string{"a", "b"}) || !reflect.DeepEqual(values, []int{3, 5}) {
		t.Errorf("expected the bars a=3 and b=5, actual %v %v, %v", labels, values, err)
	}

	if _, _, err := JSONBars(root, "total", "", ""); err == nil {
		t.Error("expected an error for a path which isn't an array")
	}
}
