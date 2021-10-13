package runner

import (
	"blackJack/log"
	"blackJack/utils"
	"crypto/tls"
	pdhttputil "github.com/projectdiscovery/httputil"
	"github.com/projectdiscovery/retryablehttp-go"
	"net/http"
	"net/url"
	"time"
)

func (r *Runner) Request(method string, url string)(headers http.Header, rawResp []byte, err error){
	request, err := r.NewRequest(method,url)
	if err != nil{
		return
	}
	do, err := r.client.Do(request)
	if err != nil{
		return
	}
	_, rawResp, err = pdhttputil.DumpResponseHeadersAndRaw(do)
	headers = request.Header.Clone()
	return
}

// NewRequest from url
func (r *Runner) NewRequest(method, targetURL string) (req *retryablehttp.Request, err error) {
	req, err = retryablehttp.NewRequest(method, targetURL, nil)
	if err != nil {
		return
	}

	// set default user agent
	req.Header.Set("User-Agent", utils.GetUserAgent())
	// 检测shiro指纹
	req.Header.Set("Cookie", "rememberMe=6gYvaCGZaDXt1c0xwriXj/Uvz6g8OMT3VSaAK4WL0Fvqvkcm0nf3CfTwkWWTT4EjeSS")
	// set default encoding to accept utf8
	req.Header.Add("Accept-Charset", "utf-8")
	return
}


/*
Must be A
*/
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


	req, err := http.NewRequest("GET", requrl, nil) //nolint
	if err != nil {
		return nil, "", *result, err
	}
	req.Header.Set("User-Agent", utils.GetUserAgent())


	resp, err := client.Do(req)
	if err != nil{
		return nil, "", *result, err
	}else{
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
	req.Header.Set("User-Agent", utils.GetUserAgent())
	req.Header.Set("Cookie", "rememberMe=6gYvaCGZaDXt1c0xwriXj/Uvz6g8OMT3VSaAK4WL0Fvqvkcm0nf3CfTwkWWTT4EjeSS")

	resp, err := client.Do(req)
	if err != nil{
		return nil, "", *result, err
	}else{
		result.URL = requrl
		result.Host = req.Host
		result.StatusCode = resp.StatusCode
		result.ContentLength = resp.ContentLength
		result.Title, body = utils.ExtractTitle(resp)
	}
	defer resp.Body.Close() //nolint
	return resp.Header, body, *result, nil
}


