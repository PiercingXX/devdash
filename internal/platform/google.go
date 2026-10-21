package platform

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// RequestTimeout of the requests to the APIs and websites.
const RequestTimeout = 30 * time.Second

// googleEndpointEnv replaces the endpoint of the Google APIs, to test DevDash with a fake API.
const googleEndpointEnv = "DEVDASH_GOOGLE_API_ENDPOINT"

// NewHTTPClient with a timeout.
func NewHTTPClient() *http.Client {
	return &http.Client{Timeout: RequestTimeout}
}

// readKeyfile of a Google service account, from its path, or from the config directory of DevDash.
func readKeyfile(keyfile string) ([]byte, error) {
	data, err := os.ReadFile(keyfile)
	if err == nil {
		return data, nil
	}

	if !filepath.IsAbs(keyfile) {
		if data, errHome := os.ReadFile(filepath.Join(xdg.ConfigHome, "devdash", keyfile)); errHome == nil {
			return data, nil
		}
	}

	return nil, fmt.Errorf("reading keyfile %q failed: %w", keyfile, err)
}

// googleClientOptions to connect to a Google API with a service account's keyfile.
func googleClientOptions(keyfile string, scope string) ([]option.ClientOption, error) {
	data, err := readKeyfile(keyfile)
	if err != nil {
		return nil, err
	}

	config, err := google.JWTConfigFromJSON(data, scope)
	if err != nil {
		return nil, fmt.Errorf("creating JWT config from json keyfile %q failed: %w", keyfile, err)
	}

	client := config.Client(context.Background())
	client.Timeout = RequestTimeout

	opts := []option.ClientOption{option.WithHTTPClient(client)}
	if endpoint := os.Getenv(googleEndpointEnv); endpoint != "" {
		opts = append(opts, option.WithEndpoint(endpoint))
	}

	return opts, nil
}

// googleError keeps the code and the message of an error of a Google API, without the body of the response,
// which can be a whole HTML page.
func googleError(err error) error {
	var gerr *googleapi.Error
	if !errors.As(err, &gerr) {
		return err
	}

	msg := gerr.Message
	if msg == "" {
		msg = http.StatusText(gerr.Code)
	}

	return fmt.Errorf("error %d of the Google API: %s", gerr.Code, strings.TrimSpace(msg))
}
