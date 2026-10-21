package platform

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"errors"
)

type Feedly struct {
	Address string
	Client  *http.Client
}

type FeedlyResponse struct {
	Results []map[string]any `json:"results"`
}

const (
	version = "v3"
	domain  = "https://feedly.com"
	method  = "search/feeds"
)

func NewFeedly(address string) *Feedly {
	return &Feedly{
		Address: address,
		Client:  NewHTTPClient(),
	}
}

func (f *Feedly) Subscribers() (string, error) {
	url := f.createAPIURL()
	resp, err := f.Client.Get(url)
	if err != nil {
		return "", fmt.Errorf("error while fetching feedly API: %w", err)
	}

	defer resp.Body.Close()
	c, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error while reading feedly API response: %w", err)
	}

	subscribers, err := extractSubscribers(c)
	if err != nil {
		return "", err
	}

	return subscribers, nil
}

func (f *Feedly) createAPIURL() string {
	params := "n=50&fullTerm=false&organic=true&promoted=true"
	return fmt.Sprintf(
		"%s/%s/%s?q=%s&%s",
		domain,
		version,
		method,
		url.QueryEscape(f.Address),
		params,
	)
}

func extractSubscribers(data []byte) (string, error) {
	fr := &FeedlyResponse{}
	err := json.Unmarshal(data, fr)
	if err != nil {
		return "", fmt.Errorf("error while unmarshal feedly API response: %w", err)
	}

	if len(fr.Results) == 0 {
		return "0", nil
	}

	if _, ok := fr.Results[0]["subscribers"]; !ok {
		return "", errors.New("wrong format for Feedly API")
	}

	s := fr.Results[0]["subscribers"].(float64)
	subs := strconv.FormatFloat(s, 'f', 0, 64)

	return subs, nil
}
