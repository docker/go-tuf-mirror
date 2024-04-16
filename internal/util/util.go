package util

import (
	"net/url"
)

// isValidUrl tests a string to determine if it is a well-structured url or not.
func IsValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	return err == nil && u.Scheme != "" && u.Host != ""
}
