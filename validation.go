package main

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var reserved = map[string]bool{
	"short":    true,
	"reserved": true,
}
var re = regexp.MustCompile(`^[a-z0-9_-]+$`)

func validateURL(rawURL string) (string, error) {

	rawURL = strings.TrimSpace(rawURL)

	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return "", err
	}

	switch u.Scheme {
	case "http", "https":
		if u.Hostname() == "" {
			return "", errors.New("Missing Host")
		}

	default:
		return "", errors.New("Invalid URL")
	}

	return u.String(), nil

}

func validateAlias(alias string) (string, error) {
	if len(alias) < 3 || len(alias) > 32 {
		return "", errors.New("Invalid Alias Length")
	}
	if _, ok := reserved[alias]; ok {
		return "", errors.New("Invalid Alias")
	}
	if !re.MatchString(alias) {
		return "", errors.New("Invalid Alias")
	}

	if alias[0] == '-' || alias[0] == '_' {
		return "", errors.New("Invalid Alias")
	}
	return alias, nil
}
