package platform

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/tidwall/gjson"
)

// maxJSONSize of a response.
const maxJSONSize = 10 << 20

// FetchJSON from a URL with a GET request. The header, like "Authorization: Bearer ${TOKEN}", can use environment
// variables.
func FetchJSON(url string, header string) (gjson.Result, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return gjson.Result{}, err
	}

	req.Header.Set("Accept", "application/json")
	if header != "" {
		name, value, ok := strings.Cut(os.ExpandEnv(header), ":")
		if !ok {
			return gjson.Result{}, fmt.Errorf("the header %q needs to be written \"Name: value\"", header)
		}
		req.Header.Set(strings.TrimSpace(name), strings.TrimSpace(value))
	}

	resp, err := NewHTTPClient().Do(req)
	if err != nil {
		return gjson.Result{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxJSONSize))
	if err != nil {
		return gjson.Result{}, fmt.Errorf("can't read the response of %s: %w", url, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return gjson.Result{}, fmt.Errorf("%s answered %s", url, resp.Status)
	}

	if !gjson.ValidBytes(body) {
		return gjson.Result{}, fmt.Errorf("%s didn't answer valid JSON", url)
	}

	return gjson.ParseBytes(body), nil
}

// JSONValue at the path, as a text. See https://github.com/tidwall/gjson/blob/master/SYNTAX.md for the paths.
func JSONValue(root gjson.Result, path string) (string, error) {
	v := root
	if path != "" {
		v = root.Get(path)
	}
	if !v.Exists() {
		return "", fmt.Errorf("nothing at the path %q", path)
	}

	return v.String(), nil
}

// JSONTable of the array at the path: one row per element, one column per field.
func JSONTable(root gjson.Result, path string, fields []string, limit int) ([][]string, error) {
	items, err := jsonArray(root, path)
	if err != nil {
		return nil, err
	}

	table := [][]string{}
	for _, item := range items[:min(limit, len(items))] {
		row := []string{}
		for _, f := range fields {
			row = append(row, item.Get(strings.TrimSpace(f)).String())
		}
		table = append(table, row)
	}

	return table, nil
}

// JSONBars of the array at the path: the label and the value of each element.
func JSONBars(root gjson.Result, path string, label string, value string) (labels []string, values []int, err error) {
	items, err := jsonArray(root, path)
	if err != nil {
		return nil, nil, err
	}

	for i, item := range items {
		v := item
		if value != "" {
			v = item.Get(value)
		}

		l := fmt.Sprint(i + 1)
		if label != "" {
			l = item.Get(label).String()
		}

		labels = append(labels, l)
		values = append(values, int(v.Float()))
	}

	return labels, values, nil
}

func jsonArray(root gjson.Result, path string) ([]gjson.Result, error) {
	v := root
	if path != "" {
		v = root.Get(path)
	}
	if !v.IsArray() {
		return nil, fmt.Errorf("the path %q isn't an array", path)
	}

	return v.Array(), nil
}
