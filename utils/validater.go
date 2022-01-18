package utils

import (
	"github.com/t43Wiu6/tlog"
	"net/url"
	"strings"
)

func ValidateUrl(targetUrl string) string{
	if strings.HasPrefix(targetUrl, "https") {
		_, err := url.Parse(targetUrl)
		if err != nil {
			log.Errorf("could not parse request URL: %s", targetUrl)
			return ""
		}
		return "https"
	}

	if strings.HasPrefix(targetUrl, "http") {
		_, err := url.Parse(targetUrl)
		if err != nil {
			log.Errorf("could not parse request URL: %s", targetUrl)
			return ""
		}
		return "http"
	}
	return "none"
}