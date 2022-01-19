package runner

import (
	"blackJack/brute"
	"blackJack/config"
	"blackJack/finger"
	"blackJack/utils"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/projectdiscovery/cdncheck"
	"github.com/projectdiscovery/fastdialer/fastdialer"
	pdhttputil "github.com/projectdiscovery/httputil"
	"github.com/pterm/pterm"
	"github.com/remeh/sizedwaitgroup"
	"github.com/t43Wiu6/tlog"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	focusOn []string
	CONFIG  finger.Config
	dirWg   = sizedwaitgroup.New(10)
)

// Runner A user options
type Runner struct {
	noRedirectClient *http.Client
	client           *http.Client
	Dialer           *fastdialer.Dialer
	options          *config.Options
	printer          *pterm.SpinnerPrinter
	DirStatus        brute.DirStatus
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
	DirBruteRs    []string `json:"dir"`
	Technologies  []string `json:"technologies,omitempty"`
}

func New(options *config.Options) (*Runner, error) {
	runner := &Runner{
		options: options,
	}
	if options.EnableDirBrute {
		spinnerLiveText, _ := pterm.DefaultSpinner.Start("[DirBrute] Waiting to Brute Force")
		runner.printer = spinnerLiveText
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
	var err error
	if !config.Check() {
		log.Error("unable to found finger & dict file")
		errs := config.DownloadAll()
		if errs != nil {
			log.Fatal("unable to download automatically")
		}
	}
	CONFIG, err = finger.LoadFinger()
	if err != nil {
		log.Fatalf("LoadFinger failed: %v", err)
	}
	config.SetEnv(r.options.IsDebug)

	// runner
	log.Warnf("Default threads: %d", r.options.Threads)
	wg := sizedwaitgroup.New(r.options.Threads) // set threads , 50 by default
	r.generate(&wg)
	wg.Wait()

	// 扫描结束 追加重点资产列表
	if r.options.Output != "" && len(focusOn) != 0 {
		r.options.OutputFile.Lock()
		defer r.options.OutputFile.Unlock()
		_, err = r.options.OutputFile.File.WriteString(fmt.Sprintf("%s", "\n重点资产: \n"))
		if err != nil {
			log.Fatalf("Could not write output file '%s': %s", r.options.Output, err)
		}

		for _, fo := range focusOn {
			_, err := r.options.OutputFile.File.WriteString(fo + "\n")
			if err != nil {
				continue
			}
		}
	}
}

// generate 读取文件或直接处理url并开始调度
func (r *Runner) generate(wg *sizedwaitgroup.SizedWaitGroup) {
	if r.options.TargetUrl != "" {
		log.Infof("single target: %s", r.options.TargetUrl)
		wg.Add()
		go r.process(r.options.TargetUrl, wg)
	} else {
		urls, err := utils.ReadFile(r.options.UrlFile)
		if err != nil {
			log.Fatal("Can't read url file")
		} else {
			log.Infof("Read %d's url", len(urls))
			for _, u := range urls {
				wg.Add()
				go r.process(u, wg)
			}
		}
	}
}

// process 请求获取每个url内容并分析后输出
func (r *Runner) process(url string, wg *sizedwaitgroup.SizedWaitGroup) () {
	defer wg.Done()
	options := &config.Options{}
	origProtocol := options.OrigProtocol
	log.Debug(url)
	ch := make(chan []string)
	faviconHash, headerContent, urlContent, resultContent, err := r.scan(url, origProtocol, ch)
	if err != nil {
		return
	}

	raw := makeOutput(analyze(faviconHash, headerContent, urlContent, resultContent))
	r.options.OutputFile.Lock()
	defer r.options.OutputFile.Unlock()
	if r.options.Output != "" && r.options.EnableDirBrute {
		// 如果开启目录爆破功能，等待目录爆破完毕一起写入文件
		dirbResult := <-ch
		r.DirStatus.DoneJob = r.DirStatus.DoneJob + 1
		if len(dirbResult) < 1 {
			return
		}

		// 先写入 指纹探测结果
		_, err = r.options.OutputFile.File.WriteString(raw + "\n")
		if err != nil {
			log.Fatalf("Could not write output file '%s': %s", r.options.Output, err)
		}
		// 后写入 目录爆破结果
		for _, raw := range dirbResult {
			_, err = r.options.OutputFile.File.WriteString(raw + "\n")
			pterm.DefaultBasicText.Print(raw + "\n")
			if err != nil {
				log.Errorf("Could not write output file '%s': %s", r.options.Output, err)
			}
		}
	} else if r.options.Output == "" && r.options.EnableDirBrute {
		dirbResult := <-ch
		r.DirStatus.DoneJob = r.DirStatus.DoneJob + 1
		if len(dirbResult) < 1 {
			return
		}
		// 没有输出文件 直接打印目录爆破结果 指纹结果已经输出过
		for _, raw := range dirbResult {
			pterm.DefaultBasicText.Print(raw + "\n")
		}
	} else if r.options.Output != "" && !r.options.EnableDirBrute {
		// 仅指纹识别
		_, err = r.options.OutputFile.File.WriteString(raw + "\n")
		if err != nil {
			log.Fatalf("Could not open output file '%s': %s", r.options.Output, err)
		}
	}
}

// scan 扫描单个url
func (r *Runner) scan(url string, origProtocol string, ch chan []string) (faviconHash string, headerContent []http.Header, urlContent []string, resultContent *Result, err error) {
	var indexUrl string
	var faviconUrl string
	var errorUrl string
	var urls []string

	// 适配协议
	if utils.ValidateUrl(url) == "" {
		log.Fatal("unable to identify url")
	}
	if utils.ValidateUrl(url) == "http" {
		url = strings.Split(url, "://")[1]
		log.Debug("validate rs is http")
		origProtocol = "http"
	}
	if utils.ValidateUrl(url) == "https" {
		url = strings.Split(url, "://")[1]
		log.Debug("validate rs is https")
		origProtocol = "https"
	}
	if strings.HasSuffix(url, "/") {
		url = strings.Split(url, "/")[0]
	}
	log.Debugf("got target: %s", url)
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
	errorUrl = fmt.Sprintf("%s://%s/%s", prot, targetUrl, utils.RandStringBytes(20))

	if utils.ValidateUrl(indexUrl) != "" && !retried {
		urls = append(urls, indexUrl)
	}
	if utils.ValidateUrl(errorUrl) != "" && !retried {
		urls = append(urls, errorUrl)
	}

	// 获得网站主页内容和不存在页面的内容
	for k, v := range urls { //range returns both the index and valu
		resp, errs := r.Request("GET", v, true)
		log.Debugf("GetContent: %v , %d", v, resp.StatusCode)
		// 如果是https，则拥有一次重试机会，避免协议不匹配问题
		if errs != nil && !retried && origProtocol == "https" {
			log.Debugf("request url %s error", v)
			retried = true
			goto retry
		}
		if errs != nil && retried || err != nil && origProtocol == "http" || resp.ContentLength == 0 {
			err = errors.New("i/o timeout")
			log.Debugf("request url %s error: %s", v, err)
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
	log.Debugf("GetContent with NoRedirect: %s", indexUrl)
	resp, err := r.Request("GET", indexUrl, false)

	if err != nil {
		log.Debugf("%v", err)
	} else {
		urlContent = append(urlContent, string(resp.Data))
		headerContent = append(headerContent, resp.Headers)
	}

	// 获取icon指纹
	log.Debugf("GetIconHash: %s", faviconUrl)
	faviconHash, err = r.GetFaviconHash(faviconUrl)
	if err == nil && faviconHash != "" {
		log.Debugf("GetIconHash: %s %s success", faviconUrl, faviconHash)
	} else if err != nil {
		log.Debugf("GetIconHash Error %s", err)
	}

	// CDN检测
	cdn, err := cdncheck.NewWithCache()
	if err != nil {
		log.Debugf("%s", err)
	} else {
		if found, err := cdn.Check(net.ParseIP(targetUrl)); found && err == nil {
			resultContent.CDN = "isCDN"
		}
	}

	// 调度目录爆破
	if r.options.EnableDirBrute {
		d := &brute.DirBrute{
			IndexUrl: indexUrl,
			ErrorUrl: errorUrl,
			Options:  r.options,
		}
		dirWg.Add()
		r.DirStatus.AllJob = r.DirStatus.AllJob + 1
		go d.Start(ch, r.printer, &dirWg, &r.DirStatus)
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
		r.Title = strings.Replace(string(resp.Data), "\n", "", -1)
		r.Title = strings.Replace(r.Title, " ", "", -1)
	}
	return
}

// makeOutput 分析指纹并生成输出
func makeOutput(resp Result) string {
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
	row := pterm.NewStyle(pterm.Bold).Sprintf("%s ", resp.URL)
	row += fmt.Sprintf("[%s] ", pterm.NewStyle(pterm.FgMagenta, pterm.Bold).Sprintf("%d", resp.StatusCode))
	row += fmt.Sprintf("[%s] ", pterm.NewStyle(pterm.FgCyan).Sprintf("%d", resp.ContentLength))
	row += fmt.Sprintf("[%s] ", pterm.NewStyle(pterm.Bold).Sprintf("%s", resp.Title))
	raw := fmt.Sprintf("%s [%d]  [%d]  [%s] ", resp.URL, resp.StatusCode, resp.ContentLength, resp.Title)

	if technology != "" {
		row += fmt.Sprintf("[%s] ", technology)
		raw += fmt.Sprintf("[%s] ", technology)
	}

	if finger != "" {
		row += fmt.Sprintf("[%s] ", pterm.NewStyle(pterm.FgRed, pterm.Bold).Sprintf("%s", finger))
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
	pterm.DefaultBasicText.Print(row + "\n")
	return raw
}

// analyze content by fingerprint
func analyze(faviconHash string, headerContent []http.Header, indexContent []string, resultContent *Result) Result {
	configs := CONFIG.Rules
	result := resultContent
	var f finger.Finger

	for _, f = range configs {
		for _, p := range f.Fingerprint {
			result.Finger = append(result.Finger, detect(f.Name, faviconHash, headerContent, indexContent, p)...)
		}
	}

	// range all header extract `X-Powered-By` and `Header` value
	for _, h := range headerContent {
		if h.Get("X-Powered-By") != "" {
			result.Technologies = append(result.Technologies, utils.StringArrayToString(h.Values("X-Powered-By")))
		}
		if h.Get("Server") != "" {
			result.Technologies = append(result.Technologies, utils.StringArrayToString(h.Values("Server")))
		}
		break
	}
	return *result
}

func detect(name string, faviconHash string, headerContent []http.Header, indexContent []string, mf finger.MetaFinger) (rs []string) {
	if len(headerContent) == 0 && len(indexContent) == 0 {
		return
	}
	if mf.Method == "keyword" {
		flag := detectKeywords(headerContent, indexContent, mf)
		if flag && !utils.StringArrayContains(rs, name) {
			rs = append(rs, name)
		}
	}

	if mf.Method == "faviconhash" && faviconHash != "" {
		if mf.Keyword[0] == faviconHash && !utils.StringArrayContains(rs, name) {
			rs = append(rs, name)
		}
	}
	return
}

func detectKeywords(headerContent []http.Header, indexContent []string, mf finger.MetaFinger) bool {
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
					if !strings.Contains(utils.StringArrayToString(v), k) {
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
getResponse:
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
			goto getResponse
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
