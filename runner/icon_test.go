package runner

import (
	"blackJack/config"
	"blackJack/log"
	"crypto/tls"
	"github.com/hashicorp/go-retryablehttp"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

func TestGetFaviconHash(t *testing.T) {

	runner := &Runner{}
	runner.options = &config.DefaultOption

	// 创建不允许重定向的http client
	var redirectFunc = func(_ *http.Request, _ []*http.Request) error {
		// Tell the http client to not follow redirect
		return http.ErrUseLastResponse
	}
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = runner.options.RetryMax

	transport := &http.Transport{
		DialContext:         runner.Dialer.Dial,
		MaxIdleConnsPerHost: -1,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // don't check cert
		},
		DisableKeepAlives: true,
	}

	if runner.options.Proxy != "" {
		proxyUrl, err := url.Parse(runner.options.Proxy)
		if err != nil {
			log.Error("Proxy Url Can Not Identify, Droped")
		} else {
			transport.Proxy = http.ProxyURL(proxyUrl)
		}
	}

	runner.client = retryClient.StandardClient() // *http.Client
	runner.client.Timeout = runner.options.TimeOut
	runner.client.Transport = transport
	req, _ := runner.NewRequest("GET", runner.options.TargetUrl)

	runner.noRedirectClient = retryClient.StandardClient() // *http.Client
	runner.noRedirectClient.CheckRedirect = redirectFunc
	runner.client.Timeout = runner.options.TimeOut
	runner.client.Transport = transport

	resp, err := runner.client.Do(req)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close() //nolint
	log.Error(string(body))
	//hash, err := runner.GetFaviconHash(url)
	//if err != nil && hash != "1693998826"{
	//	t.Errorf("Analyze test error")
	//}
}
