package platform

import (
	"errors"
	"fmt"
	"testing"

	"google.golang.org/api/googleapi"
)

func Test_googleError(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected string
	}{
		{
			"message of the API",
			&googleapi.Error{Code: 403, Message: "User does not have sufficient permissions for this property."},
			"error 403 of the Google API: User does not have sufficient permissions for this property.",
		},
		{
			"HTML body without message",
			fmt.Errorf("wrapped: %w", &googleapi.Error{Code: 404, Body: "<!DOCTYPE html><html>...</html>"}),
			"error 404 of the Google API: Not Found",
		},
		{"other error", errors.New("timeout"), "timeout"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if actual := googleError(tc.err).Error(); actual != tc.expected {
				t.Errorf("Expected %q, actual %q", tc.expected, actual)
			}
		})
	}
}
