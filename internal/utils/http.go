package utils

import (
	"fmt"
	"net/http"
)

// GetURL composes a URL from http.Request object
func GetURL(r *http.Request) string {
	url := r.URL.RequestURI()

	// If the URL is relative, RequestURI will return only the path of it.
	if !r.URL.IsAbs() {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		url = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.URL.String())
	}

	return url
}
