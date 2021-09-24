package libs

import (
	"blackJack/log"
	"net/url"
	"strings"
)

func ValidateUrl(targetUrl string) string{
	if strings.HasPrefix(targetUrl, "https") {
		_, err := url.Parse(targetUrl)
		if err != nil {
			log.Error("could not parse request URL: "+targetUrl)
			return ""
		}
		return "https"
	}

	if strings.HasPrefix(targetUrl, "http") {
		_, err := url.Parse(targetUrl)
		if err != nil {
			log.Error("could not parse request URL: "+targetUrl)
			return ""
		}
		return "http"
	}
	return "none"
}