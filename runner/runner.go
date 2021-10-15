package runner

import (
	"blackJack/config"
	"blackJack/log"
	. "blackJack/utils"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	. "github.com/logrusorgru/aurora"
	"github.com/projectdiscovery/cdncheck"
	"github.com/projectdiscovery/fastdialer/fastdialer"
	pdhttputil "github.com/projectdiscovery/httputil"
	"github.com/remeh/sizedwaitgroup"
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
var CONFIG config.Config

// Runner A user options
type Runner struct {
	noRedirectClient *http.Client
	client           *http.Client
	Dialer           *fastdialer.Dialer
	options          *config.Options
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

func New(options *config.Options) (*Runner, error) {
	runner := &Runner{
		options: options,
	}

	// 创建不允许重定向的http client
	var redirectFunc = func(_ *http.Request, _ []*http.Request) error {
		// Tell the http client to not follow redirect
		return http.ErrUseLastResponse
	}
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = runner.options.RetryMax

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
			DualStack: true,
		}).DialContext,
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

	runner.noRedirectClient = retryClient.StandardClient() // *http.Client
	runner.noRedirectClient.CheckRedirect = redirectFunc
	runner.noRedirectClient.Timeout = runner.options.TimeOut
	runner.noRedirectClient.Transport = transport
	return runner, nil
}

// CreateRunner 创建扫描
func (r *Runner) CreateRunner() {
	_, CONFIG = config.LoadFinger()
	config.SetEnv(r.options.IsDebug)

	// output routine
	wgoutput := sizedwaitgroup.New(1)
	wgoutput.Add()
	output := make(chan Result)
	go r.output(output, &wgoutput)

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
	if r.options.TargetUrl != "" {
		log.Info(fmt.Sprintf("single target: %s", r.options.TargetUrl))
		wg.Add()
		go r.process(output, r.options.TargetUrl, wg)
	} else {
		urls, err := ReadFile(r.options.UrlFile)
		if err != nil {
			log.Fatal("Cann't read url file")
		} else {
			log.Info(fmt.Sprintf("Read %d's url totaly", len(urls)))
			for _, u := range urls {
				wg.Add()
				go r.process(output, u, wg)
			}
		}
	}
}

// process 请求获取每个url内容用于后续分析
func (r *Runner) process(output chan Result, url string, wg *sizedwaitgroup.SizedWaitGroup) () {
	defer wg.Done()
	options := &config.Options{}
	origProtocol := options.OrigProtocol
	log.Debug(url)
	faviconHash, headerContent, urlContent, resultContent, err := r.scan(url, origProtocol)
	if err == nil {
		output <- analyze(faviconHash, headerContent, urlContent, resultContent)
	}
}

// scan 扫描单个url
// 对单个目标
// 1. 请求主页一次
// 2. 不允许重定向地请求主页一次
// 3. 请求icon一次
// 4. 请求不存在的页面一次
// return icon指纹 响应头列表 响应体列表 结果样例
func (r *Runner) scan(url string, origProtocol string) (faviconHash string, headerContent []http.Header, urlContent []string, resultContent *Result, err error) {
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
		err = errors.New("no a valid domain or ip")
		return
	}
	targetUrl := url
	prot := "https" // have no protocol, use https default
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
		resp, errs := r.Request("GET", v, true)
		// 如果是https，则拥有一次重试机会，避免协议不匹配问题
		if errs != nil && !retried && origProtocol == "https" {
			log.Debug(fmt.Sprintf("request url %s error", v))
			retried = true
			goto retry
		}
		if errs != nil && retried || err != nil && origProtocol == "http" || resp.ContentLength == 0 {
			err = errors.New("i/o timeout")
			return
		} else {
			urlContent = append(urlContent, string(resp.Data))
			headerContent = append(headerContent, resp.Headers)
			if k == 0 {
				resultContent = makeResultTemplate(resp)
			}
		}
	}

	// 以非重定向的方式获得网站内容
	// 同时分析两种情况，避免重定向跳转导致获得失败，避免反向代理和CDN导致收集面缩小
	log.Debug("GetContent with NoRedirect: " + indexUrl)
	resp, err := r.Request("GET", indexUrl, false)

	if err != nil {
		log.Debug(fmt.Sprintf("%s", err))
	} else {
		urlContent = append(urlContent, string(resp.Data))
		headerContent = append(headerContent, resp.Headers)
	}

	// 获取icon指纹
	log.Debug("GetIconHash: " + faviconUrl)
	faviconHash, err = r.GetFaviconHash(faviconUrl)
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

func makeResultTemplate(resp Response) (r *Result) {
	r = &Result{
		Raw:           resp.Raw,
		URL:           resp.URL.String(),
		Location:      "None", //
		Title:         resp.Title,
		Host:          resp.Host, //
		ContentLength: int64(resp.ContentLength),
		StatusCode:    resp.StatusCode,
		VHost:         "noVhost", //
		CDN:           "noCDN",   //
		Finger:        []string{},
		Technologies:  []string{},
	}
	if resp.Title == "" && resp.ContentLength < 120 {
		r.Title = string(resp.Data)
	}
	return
}

// output 输出处理
func (r *Runner) output(output chan Result, wgoutput *sizedwaitgroup.SizedWaitGroup) {
	defer wgoutput.Done()
	for resp := range output {
		var f *os.File
		var finger string
		var technology string
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

		row := fmt.Sprintf("%s ", Bold(resp.URL))
		row += fmt.Sprintf("[%d] ", Bold(Magenta(resp.StatusCode)))
		row += fmt.Sprintf("[%d] ", Cyan(resp.ContentLength))
		row += fmt.Sprintf("[%s] ", Bold(Cyan(resp.Title)))
		raw := fmt.Sprintf("%s [%d]  [%d]  [%s] ", resp.URL, resp.StatusCode, resp.ContentLength, resp.Title)

		if technology != "" {
			row += fmt.Sprintf("[%s] ", technology)
			raw += fmt.Sprintf("[%s] ", technology)
		}

		if finger != "" {
			row += fmt.Sprintf("[%s] ", Bold(Red(finger)))
			raw += fmt.Sprintf("[%s] ", finger)
			focusOn = append(focusOn, row)
		}

		if resp.VHost != "noVhost" {
			row += fmt.Sprintf("[%s] ", resp.VHost)
			raw += fmt.Sprintf("[%s] ", resp.VHost)
		}
		if resp.CDN != "noCDN" {
			row += fmt.Sprintf("[%s] ", resp.CDN)
			raw += fmt.Sprintf("[%s] ", resp.CDN)
		}
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
	var f config.Finger

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

func detect(name string, faviconHash string, headerContent []http.Header, indexContent []string, mf config.MetaFinger) (rs []string) {
	if len(headerContent) == 0 && len(indexContent) == 0 {
		return
	}
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

func detectKeywords(headerContent []http.Header, indexContent []string, mf config.MetaFinger) bool {
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

func (r *Runner) Do(req *http.Request) (*Response, error) {
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
		if err != nil {
			return nil, err
		}
	}

	closeErr := httpresp.Body.Close()
	if closeErr != nil {
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
