package runner

import (
	. "blackJack/config"
	"blackJack/log"
	. "blackJack/utils"
	"crypto/tls"
	"fmt"
	. "github.com/logrusorgru/aurora"
	"github.com/projectdiscovery/cdncheck"
	"github.com/projectdiscovery/fastdialer/fastdialer"
	pdhttputil "github.com/projectdiscovery/httputil"
	"github.com/projectdiscovery/retryablehttp-go"
	"github.com/remeh/sizedwaitgroup"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var focusOn []string
var CONFIG Config

// Runner A user options
type Runner struct {
	bj               *BlackJack
	noRedirectClient *retryablehttp.Client
	client           *retryablehttp.Client
	Dialer           *fastdialer.Dialer
	options          *Options
}

// Response contains the response to a server
type Response struct {
	StatusCode    int
	Headers       map[string][]string
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

// Result of a scan
type Result struct {
	Raw           string
	URL           string   `json:"url,omitempty"`
	Location      string   `json:"location,omitempty"`
	Title         string   `json:"title,omitempty"`
	Host          string   `json:"host,omitempty"`
	ContentLength int64    `json:"content-length,omitempty"`
	StatusCode    int      `json:"status-code,omitempty"`
	VHost         string   `json:"vhost,omitempty"`
	CDN           string   `json:"cdn,omitempty"`
	Finger        []string `json:"finger,omitempty"`
	Technologies  []string `json:"technologies,omitempty"`
}

func New(options *Options) (*Runner, error) {
	runner := &Runner{
		options: options,
	}
	// 创建不允许重定向的http client
	var redirectFunc = func(_ *http.Request, _ []*http.Request) error {
		// Tell the http client to not follow redirect
		return http.ErrUseLastResponse
	}

	var retryablehttpOptions = retryablehttp.DefaultOptionsSpraying
	retryablehttpOptions.Timeout = options.TimeOut
	retryablehttpOptions.RetryMax = options.RetryMax

	transport := &http.Transport{
		DialContext:         runner.Dialer.Dial,
		MaxIdleConnsPerHost: -1,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // don't check cert
		},
		DisableKeepAlives: true,
	}

	if options.Proxy != "" {
		proxyUrl, err := url.Parse(options.Proxy)
		if err != nil {
			log.Error("Proxy Url Can Not Identify, Droped")
			return nil, err
		} else {
			transport.Proxy = http.ProxyURL(proxyUrl)
		}
	}

	runner.client = retryablehttp.NewWithHTTPClient(&http.Client{
		Transport:     transport,
		Timeout:       30 * time.Second,
		CheckRedirect: redirectFunc,
	}, retryablehttpOptions)

	runner.noRedirectClient = retryablehttp.NewWithHTTPClient(&http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}, retryablehttpOptions)

	return runner, nil
}

// CreateRunner 创建扫描
func (r *Runner) CreateRunner(options *Options) {
	_, CONFIG = LoadFinger()
	SetEnv(r.options.isDebug)

	// output routine
	wgoutput := sizedwaitgroup.New(1)
	wgoutput.Add()
	output := make(chan Result)
	r.output(output, &wgoutput)

	// runner
	log.Warn(fmt.Sprintf("Default threads: %d", r.options.Threads))
	wg := sizedwaitgroup.New(r.options.Threads) // set threads , 50 by default
	r.generate(output, &wg)
	wg.Wait()
	close(output)
	wgoutput.Wait()

	if len(focusOn) != 0 {
		fmt.Println(" ")
		log.Info(fmt.Sprintf("%s", Bold(Green("重点资产: "))))
		for _, f := range focusOn {
			fmt.Println(f)
		}
	}
}

// generate all target url
func (r *Runner) generate(output chan Result, wg *sizedwaitgroup.SizedWaitGroup) {
	if r.options.targetUrl != "" {
		log.Info(fmt.Sprintf("single target: %s", r.options.targetUrl))
		wg.Add()
		go r.process(output, r.options, r.options.targetUrl, wg)
	} else {
		urls, err := ReadFile(r.options.urlFile)
		if err != nil {
			log.Fatal("Cann't read url file")
		} else {
			log.Info(fmt.Sprintf("Read %d's url totaly", len(urls)))
			for _, u := range urls {
				wg.Add()
				go r.process(output, r.options, u, wg)
			}
		}
	}
}

// process 请求获取每个url内容用于后续分析
func (r *Runner) process(output chan Result, ret *Options, url string, wg *sizedwaitgroup.SizedWaitGroup) () {
	defer wg.Done()
	options := &Options{}
	proxy := ret.Proxy
	origProtocol := options.origProtocol
	log.Debug(options.origProtocol)
	log.Debug(url)
	faviconHash, headerContent, urlContent, resultContent := r.scan(url, proxy, origProtocol)
	output <- analyze(faviconHash, headerContent, urlContent, resultContent)
}

// scan 扫描单个url
// return icon指纹 响应头列表 响应体列表 结果样例
func (r *Runner)scan(url string, proxy string, origProtocol string) (faviconHash string, headerContent []http.Header, urlContent []string, resultContent *Result) {
	var indexUrl string
	var faviconUrl string
	var errorUrl string
	var urls []string

	// 适配协议
	if ValidateUrl(url) == "" {
		log.Fatal("unable to identify url")
	}
	if ValidateUrl(url) == "http" {
		url = strings.Split(url, "://")[1]
		log.Debug("validate rs is http")
		origProtocol = "http"
	}
	if ValidateUrl(url) == "https" {
		url = strings.Split(url, "://")[1]
		log.Debug("validate rs is https")
		origProtocol = "https"
	}
	if strings.HasSuffix(url, "/") {
		url = strings.Split(url, "/")[0]
	}
	log.Debug("got target: " + url)
	if !strings.Contains(url, ".") {
		log.Error("no a valid domain or ip")
		//return Result{}, errors.New("no a valid domain or ip")
	}
	targetUrl := url
	prot := "https"
	retried := false
retry:
	if origProtocol == "https" {
		prot = "https"
	}
	if origProtocol == "http" {
		prot = "http"
	}

	if retried && origProtocol == "https" {
		prot = "http"
	}
	log.Debug(origProtocol)

	indexUrl = fmt.Sprintf("%s://%s", prot, targetUrl)
	faviconUrl = fmt.Sprintf("%s://%s/%s", prot, targetUrl, "favicon.ico")
	errorUrl = fmt.Sprintf("%s://%s/%s", prot, targetUrl, RandStringBytes(20))

	if ValidateUrl(indexUrl) != "" && !retried {
		urls = append(urls, indexUrl)
	}
	if ValidateUrl(errorUrl) != "" && !retried {
		urls = append(urls, errorUrl)
	}

	// 获得网站主页内容和不存在页面的内容
	for k, v := range urls { //range returns both the index and value
		log.Debug("GetContent: " + v)
		header, content, err := r.Request("GET",v)
		// 如果是https，则拥有一次重试机会，避免协议不匹配问题
		if err != nil && !retried && origProtocol == "https" {
			log.Debug(fmt.Sprintf("request url %s error", v))
			retried = true
			goto retry
		} else {
			urlContent = append(urlContent, string(content...))
			headerContent = append(headerContent, header)
			if k == 0 {
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
				resultContent = &result
			}
		}
	}

	// 以非重定向的方式获得网站内容
	// 同时分析两种情况，避免重定向跳转导致获得失败，避免反向代理和CDN导致收集面缩小
	log.Debug("GetContent with NoRedirect: " + indexUrl)
	header, content, _, err := HttpReqWithNoRedirect(indexUrl, timeOut, proxy)
	if err != nil {
		log.Debug(fmt.Sprintf("%s", err))
	} else {
		urlContent = append(urlContent, content)
		headerContent = append(headerContent, header)
	}

	// 获取icon指纹
	log.Debug("GetIconHash: " + faviconUrl)
	faviconHash, err = GetFaviconHash(faviconUrl, proxy)
	if err == nil && faviconHash != "" {
		log.Debug(fmt.Sprintf("GetIconHash: %s %s success", faviconUrl, faviconHash))
	} else if err != nil {
		log.Debug(fmt.Sprintf("GetIconHash Error %s", err))
	}

	// CDN检测
	cdn, err := cdncheck.NewWithCache()
	if err != nil {
		log.Debug(fmt.Sprintf("%s", err))
	} else {
		if found, err := cdn.Check(net.ParseIP(targetUrl)); found && err == nil {
			resultContent.CDN = "isCDN"
		}
	}
	return
}

// output 输出处理
func (r *Runner)output(output chan Result, wgoutput *sizedwaitgroup.SizedWaitGroup) {
	defer wgoutput.Done()
	for resp := range output{
		var f *os.File
		var finger string
		var technology string
		if resp.URL == "None" {
			return
		}
		for k, v := range resp.Finger {
			if k == 0 {
				finger = v
			} else {
				finger = finger + "," + v
			}
		}

		for k, v := range resp.Technologies {
			if k == 0 {
				technology = v
			} else {
				technology = technology + "," + v
			}
		}

		row := fmt.Sprintf("[%s]", Bold(Green(resp.URL)))
		row += fmt.Sprintf("[%d]", Bold(Cyan(resp.StatusCode)))
		row += fmt.Sprintf("[%s]", Bold(Magenta(resp.Title)))
		row += fmt.Sprintf("[%s]", Bold(Red(finger)))
		row += fmt.Sprintf("[%s]", technology)

		raw := fmt.Sprintf("[%s] [%d] [%s] [%s] [%s]", resp.URL, resp.StatusCode, resp.Title, finger, technology)

		if finger != "" {
			focusOn = append(focusOn, row)
		}

		if resp.VHost != "noVhost" {
			row += fmt.Sprintf("[%s]", resp.VHost)
			raw += fmt.Sprintf("[%s]", resp.VHost)
		}
		if resp.CDN != "noCDN" {
			row += fmt.Sprintf("[%s]", resp.CDN)
			raw += fmt.Sprintf("[%s]", resp.CDN)
		}

		//row = fmt.Sprintf("[%s] [%d] [%s] [%s] [%s] [%s] [%s]", Bold(Green(resp.URL)), Bold(Cyan(resp.StatusCode)), Bold(Magenta(resp.Title)), Bold(Red(finger)), resp.Host, resp.VHost, resp.CDN)
		fmt.Println(row)

		if r.options.Output != "" {
			var err error
			f, err = os.Create(r.options.Output)
			if err != nil {
				log.Fatal(fmt.Sprintf("Could not create output file '%s': %s", r.options.Output, err))
			}
			f.WriteString(raw + "\n")
			defer f.Close() //nolint
		}
	}
}

// analyze content by fingerprint
//
// analyze with 3 ways
// fisrstly, Judgment keyword from finger
// secondly,  detect faviconhash
// last, extract `X-Powered-By` and `Header` in header
//
// append to the result.Technologies[] if hitted
func analyze(faviconHash string, headerContent []http.Header, indexContent []string, resultContent *Result) Result {
	configs := CONFIG.Rules
	result := resultContent
	var f Finger

	for _, f = range configs {
		for _, p := range f.Fingerprint {
			result.Finger = append(result.Finger, detect(f.Name, faviconHash, headerContent, indexContent, p)...)
		}
	}

	// range all header extract `X-Powered-By` and `Header` value
	for _, h := range headerContent {
		if h.Get("X-Powered-By") != "" {
			//log.Debug(StringArrayToString(h.Values("X-Powered-By")), true)
			result.Technologies = append(result.Technologies, StringArrayToString(h.Values("X-Powered-By")))
		}
		if h.Get("Server") != "" {
			//log.Debug(StringArrayToString(h.Values("Server")), true)
			result.Technologies = append(result.Technologies, StringArrayToString(h.Values("Server")))
		}
		break
	}
	return *result
}

func detect(name string, faviconHash string, headerContent []http.Header, indexContent []string, mf MetaFinger) (rs []string) {
	if mf.Method == "keyword" {
		flag := detectKeywords(headerContent, indexContent, mf)
		if flag && !StringArrayContains(rs, name) {
			rs = append(rs, name)
		}
	}

	if mf.Method == "faviconhash" && faviconHash != "" {
		if mf.Keyword[0] == faviconHash && !StringArrayContains(rs, name) {
			rs = append(rs, name)
		}
	}
	return
}

func detectKeywords(headerContent []http.Header, indexContent []string, mf MetaFinger) bool {
	// if keyword in body
	if mf.Location == "body" {
		for _, k := range mf.Keyword {
			for _, c := range indexContent {
				// make sure Both conditions are satisfied
				// TODO  && mf.StatusCode == header
				if !strings.Contains(c, k) {
					return false
				}
			}
		}
	}

	if mf.Location == "header" {
		for _, k := range mf.Keyword {
			for _, h := range headerContent {
				for _, v := range h {
					if !strings.Contains(StringArrayToString(v), k) {
						return false
					}
				}
			}
		}
	}
	return true
}

func (r *Runner) Do(req *retryablehttp.Request) (*Response, error) {
	timeStart := time.Now()

	var gzipRetry bool
get_response:
	httpresp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}

	var resp Response

	resp.Headers = httpresp.Header.Clone()

	// httputil.DumpResponse does not handle websockets
	headers, rawResp, err := pdhttputil.DumpResponseHeadersAndRaw(httpresp)
	if err != nil {
		// Edge case - some servers respond with gzip encoding header but uncompressed body, in this case the standard library configures the reader as gzip, triggering an error when read.
		// The bytes slice is not accessible because of abstraction, therefore we need to perform the request again tampering the Accept-Encoding header
		if !gzipRetry && strings.Contains(err.Error(), "gzip: invalid header") {
			gzipRetry = true
			req.Header.Set("Accept-Encoding", "identity")
			goto get_response
		}
	}
	resp.Raw = string(rawResp)
	resp.RawHeaders = string(headers)

	var respbody []byte
	// websockets don't have a readable body
	if httpresp.StatusCode != http.StatusSwitchingProtocols {
		var err error
		respbody, err = ioutil.ReadAll(httpresp.Body)
		if err != nil{
			return nil, err
		}
	}

	closeErr := httpresp.Body.Close()
	if closeErr != nil{
		return nil, closeErr
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
	resp.StatusCode = httpresp.StatusCode
	// number of words
	resp.Words = len(strings.Split(respbodystr, " "))
	// number of lines
	resp.Lines = len(strings.Split(respbodystr, "\n"))

	resp.Duration = time.Since(timeStart)
	return &resp, nil
}
