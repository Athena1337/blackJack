package libs

import (
	"blackJack/log"
	"crypto/tls"
	"fmt"
	"github.com/projectdiscovery/cdncheck"
	"net"
	"net/http"
	"net/url"
	"time"
)


/*
Must be A
*/
type Finger struct {
	Cms string
	Method string
	Location string
	Keyword []string
}

type Config struct {
	Fingerprint []Finger
}

// Result of a scan
type Result struct {
	Raw              string
	URL              string `json:"url,omitempty"`
	Location         string `json:"location,omitempty"`
	Title            string `json:"title,omitempty"`
	Host             string `json:"host,omitempty"`
	ContentLength    int64  `json:"content-length,omitempty"`
	StatusCode       int    `json:"status-code,omitempty"`
	VHost            string `json:"vhost,omitempty"`
	CDN              string `json:"cdn,omitempty"`
	Finger			 []string `json:"finger,omitempty"`
	Technologies     []string `json:"technologies,omitempty"`
}

func HttpReqWithNoRedirect(requrl string, timeOut int, proxy string) (http.Header, string, Result, error){
	var body string
	result := &Result{
		Raw: "None",
		URL: "None",
		Location: "None", //
		Title: "None",
		Host: "None", //
		ContentLength: 0,
		StatusCode: 0,
		VHost: "noVhost", //
		CDN: "noCDN", //
		Technologies: []string{},
	}

	var client http.Client

	if proxy != ""{
		proxyUrl, err := url.Parse(proxy)
		if err != nil{
			log.Error("Proxy Url Can Not Identify, Droped")
			client = http.Client{
				Timeout: time.Second * time.Duration(timeOut),
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // don't check cert
				},
				CheckRedirect: 	func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}
		}else{
			client = http.Client{
				Timeout: time.Second * time.Duration(timeOut),
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // don't check cert
					Proxy: http.ProxyURL(proxyUrl),
				},
				CheckRedirect: 	func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}
		}
	}else{
		client = http.Client{
			Timeout: time.Second * time.Duration(timeOut),
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // don't check cert
			},
			CheckRedirect: 	func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}


	req, err := http.NewRequest("GET", requrl, nil) //nolint
	if err != nil {
		return nil, "", *result, err
	}
	req.Header.Set("User-Agent", GetUserAgent())
	// 检测shiro指纹
	req.Header.Set("Cookie", "rememberMe=6gYvaCGZaDXt1c0xwriXj/Uvz6g8OMT3VSaAK4WL0Fvqvkcm0nf3CfTwkWWTT4EjeSS")

	resp, err := client.Do(req)
	if err != nil{
		return nil, "", *result, err
	}else{
		result.URL = requrl
		result.Host = req.Host
		result.StatusCode = resp.StatusCode
		result.ContentLength = resp.ContentLength
		result.Title, body = ExtractTitle(resp)
	}
	defer resp.Body.Close() //nolint


	cdn, err := cdncheck.NewWithCache()
    if err != nil {
        log.Debug(fmt.Sprintf("%s", err))
    }else{
		if found, err := cdn.Check(net.ParseIP(req.Host)); found && err == nil {
			result.CDN = "isCDN"
		}
		if err != nil {
			return nil, "", *result, err
		}
	}
	return resp.Header, body, *result, nil
}

/*
HttpReq 从 URL 中获取内容
*/
func HttpReq(requrl string, timeOut int, proxy string) (http.Header, string, Result, error) {
	var body string
	result := &Result{
		Raw: "None",
		URL: "None",
		Location: "None", //
		Title: "None",
		Host: "None", //
		ContentLength: 0,
		StatusCode: 0,
		VHost: "noVhost", //
		CDN: "noCDN", //
		Technologies: []string{},
	}
	var client http.Client

	if proxy != ""{
		proxyUrl, err := url.Parse(proxy)
		if err != nil{
			log.Error("Proxy Url Can Not Identify, Droped")
			client = http.Client{
				Timeout: time.Second * time.Duration(timeOut),
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // don't check cert
				},
			}
		}else{
			client = http.Client{
				Timeout: time.Second * time.Duration(timeOut),
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // don't check cert
					Proxy: http.ProxyURL(proxyUrl),
				},
			}
		}
	}else{
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
	req.Header.Set("User-Agent", GetUserAgent())
	req.Header.Set("Cookie", "rememberMe=6gYvaCGZaDXt1c0xwriXj/Uvz6g8OMT3VSaAK4WL0Fvqvkcm0nf3CfTwkWWTT4EjeSS")

	resp, err := client.Do(req)
	if err != nil{
		return nil, "", *result, err
	}else{
		result.URL = requrl
		result.Host = req.Host
		result.StatusCode = resp.StatusCode
		result.ContentLength = resp.ContentLength
		result.Title, body = ExtractTitle(resp)
	}
	defer resp.Body.Close() //nolint


	cdn, err := cdncheck.NewWithCache()
    if err != nil {
        log.Debug(fmt.Sprintf("%s", err))
    }else{
		if found, err := cdn.Check(net.ParseIP(req.Host)); found && err == nil {
			result.CDN = "isCDN"
		}
		if err != nil {
			return nil, "", *result, err
		}
	}
	return resp.Header, body, *result, nil
}


