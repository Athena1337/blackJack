package runner

import (
	"blackJack/log"
	"blackJack/utils"
	"crypto/tls"
	pdhttputil "github.com/projectdiscovery/httputil"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Response contains the response to a server
type Response struct {
	StatusCode    int
	Headers       map[string][]string
	Host          string
	Title         string
	Data          []byte
	ContentLength int
	Raw           string
	RawHeaders    string
	Words         int
	Lines         int
	CSPData       *CSPData
	HTTP2         bool
	Pipeline      bool
	Duration      time.Duration
}

func (r *Runner) Request(method string, url string, redirect bool) (resp Response, err error) {
	timeStart := time.Now()
	var gzipRetry bool
	request, err := r.NewRequest(method, url)
get_response:
	if err != nil {
		return
	}
	var do *http.Response
	if redirect{
		do, err = r.client.Do(request)
	}else{
		do, err = r.noRedirectClient.Do(request)
	}

	if err != nil {
		return
	}

	rawHeader, rawResp, err := pdhttputil.DumpResponseHeadersAndRaw(do)
	if err != nil {
		// Edge case - some servers respond with gzip encoding header but uncompressed body, in this case the standard library configures the reader as gzip, triggering an error when read.
		// The bytes slice is not accessible because of abstraction, therefore we need to perform the request again tampering the Accept-Encoding header
		if !gzipRetry && strings.Contains(err.Error(), "gzip: invalid header") {
			gzipRetry = true
			request.Header.Set("Accept-Encoding", "identity")
			goto get_response
		}
	}
	resp.Raw = string(rawResp)
	resp.RawHeaders = string(rawHeader)
	resp.Headers = request.Header.Clone()

	var respbody []byte
	// websockets don't have a readable body
	if do.StatusCode != http.StatusSwitchingProtocols {
		respbody, err = ioutil.ReadAll(do.Body)
		if err != nil {
			return
		}
	}

	err = do.Body.Close()
	if err != nil {
		return
	}

	respbodystr := string(respbody)

	// if content length is not defined
	if resp.ContentLength <= 0 {
		// check if it's in the header and convert to int
		if contentLength, ok := resp.Headers["Content-Length"]; ok {
			contentLengthInt, _ := strconv.Atoi(strings.Join(contentLength, ""))
			resp.ContentLength = contentLengthInt
		}

		// if we have a body, then use the number of bytes in the body if the length is still zero
		if resp.ContentLength <= 0 && len(respbodystr) > 0 {
			resp.ContentLength = utf8.RuneCountInString(respbodystr)
		}
	}

	resp.Data = respbody

	// fill metrics
	resp.StatusCode = do.StatusCode
	// number of words
	resp.Words = len(strings.Split(respbodystr, " "))
	// number of lines
	resp.Lines = len(strings.Split(respbodystr, "\n"))

	resp.Duration = time.Since(timeStart)

	resp.Title, _ = utils.ExtractTitle(do)
	return
}

// NewRequest from url
func (r *Runner) NewRequest(method, targetURL string) (req *http.Request, err error) {
	//req, err = retryablehttp.NewRequest(method, targetURL, nil)
	req = &http.Request{}
	req.Method = method
	parse, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	req.URL = parse
	if err != nil {
		return
	}


	// set default user agent
	req.Header.Add("User-Agent", utils.GetUserAgent())
	// 检测shiro指纹
	req.Header.Add("Cookie", "rememberMe=6gYvaCGZaDXt1c0xwriXj/Uvz6g8OMT3VSaAK4WL0Fvqvkcm0nf3CfTwkWWTT4EjeSS")
	// set default encoding to accept utf8
	req.Header.Add("Accept-Charset", "utf-8")
	return
}

/*
Must be A
*/
func HttpReqWithNoRedirect(requrl string, timeOut int, proxy string) (http.Header, string, Result, error) {
	var body string
	result := &Result{
		Raw:           "None",
		URL:           "None",
		Location:      "None", //
		Title:         "None",
		Host:          "None", //
		ContentLength: 0,
		StatusCode:    0,
		VHost:         "noVhost", //
		CDN:           "noCDN",   //
		Technologies:  []string{},
	}

	var client http.Client

	req, err := http.NewRequest("GET", requrl, nil) //nolint
	if err != nil {
		return nil, "", *result, err
	}
	req.Header.Set("User-Agent", utils.GetUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", *result, err
	} else {
		result.URL = requrl
		result.Host = req.Host
		result.StatusCode = resp.StatusCode
		result.ContentLength = resp.ContentLength
		result.Title, body = utils.ExtractTitle(resp)
	}
	defer resp.Body.Close() //nolint
	return resp.Header, body, *result, nil
}

/*
HttpReq 从 URL 中获取内容
*/
func HttpReq(requrl string, timeOut int, proxy string) (http.Header, string, Result, error) {
	var body string
	result := &Result{
		Raw:           "None",
		URL:           "None",
		Location:      "None", //
		Title:         "None",
		Host:          "None", //
		ContentLength: 0,
		StatusCode:    0,
		VHost:         "noVhost", //
		CDN:           "noCDN",   //
		Technologies:  []string{},
	}
	var client http.Client

	if proxy != "" {
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			log.Error("Proxy Url Can Not Identify, Droped")
			client = http.Client{
				Timeout: time.Second * time.Duration(timeOut),
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // don't check cert
				},
			}
		} else {
			client = http.Client{
				Timeout: time.Second * time.Duration(timeOut),
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // don't check cert
					Proxy:           http.ProxyURL(proxyUrl),
				},
			}
		}
	} else {
		client = http.Client{
			Timeout: time.Second * time.Duration(timeOut),
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // don't check cert
			},
		}
	}

	req, err := http.NewRequest("GET", requrl, nil) //nolint
	if err != nil {
		return nil, "", *result, err
	}
	req.Header.Set("User-Agent", utils.GetUserAgent())
	req.Header.Set("Cookie", "rememberMe=6gYvaCGZaDXt1c0xwriXj/Uvz6g8OMT3VSaAK4WL0Fvqvkcm0nf3CfTwkWWTT4EjeSS")

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", *result, err
	} else {
		result.URL = requrl
		result.Host = req.Host
		result.StatusCode = resp.StatusCode
		result.ContentLength = resp.ContentLength
		result.Title, body = utils.ExtractTitle(resp)
	}
	defer resp.Body.Close() //nolint
	return resp.Header, body, *result, nil
}
