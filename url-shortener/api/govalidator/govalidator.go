package govalidator

import (
	"net/url"
	"strings"
)

func IsURL(str string) bool {
	u, err := url.ParseRequestURI(str)
	if err != nil {
		return false
	}
	scheme := strings.ToLower(u.Scheme)

	if scheme != "http" && scheme != "https" {
		return false
	}
	if u.Host == "" {
		return false
	}
	return true
}
