package runner

import (
	. "blackJack/config"
	"blackJack/log"
	. "blackJack/utils"
	"fmt"
	. "github.com/logrusorgru/aurora"
	"net/http"
	"os"
	"strings"
	//"errors"
	"github.com/remeh/sizedwaitgroup"
)

var focusOn []string
var CONFIG Config

// Runner A user options
type Runner struct {
	options *Options
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

func Init(options *Options) *Runner {
	runner := &Runner{
		options: options,
	}
	return runner
}

// CreateRunner 创建扫描
func (r *Runner) CreateRunner() {
	_, CONFIG = LoadFinger()
	SetEnv(r.options.isDebug)

	log.Warn(fmt.Sprintf("Default threads: %d", r.options.Threads))
	wg := sizedwaitgroup.New(r.options.Threads) // set threads , 50 by default
	r.generate(&wg)
	wg.Wait()
	if len(focusOn) != 0{
		fmt.Println(" ")
		log.Info(fmt.Sprintf("%s",Bold(Green("重点资产: "))))
		for _,f := range focusOn{
			fmt.Println(f)
		}
	}
}

// generate all target url
func (r *Runner) generate(wg *sizedwaitgroup.SizedWaitGroup) {
	if r.options.targetUrl != "" {
		log.Info(fmt.Sprintf("single target: %s", r.options.targetUrl))
		wg.Add()
		go r.process(r.options, r.options.targetUrl, wg)
	} else {
		urls, err := ReadFile(r.options.urlFile)
		if err != nil {
			log.Fatal("Cann't read url file")
		} else {
			log.Info(fmt.Sprintf("Read %d's url totaly", len(urls)))
			for _, u := range urls {
				wg.Add()
				go r.process(r.options, u, wg)
			}
		}
	}
}

// process 请求获取每个url内容用于后续分析
func (r *Runner) process(ret *Options, url string, wg *sizedwaitgroup.SizedWaitGroup) () {
	defer wg.Done()
	options := &Options{}
	proxy := ret.Proxy
	timeOut := options.TimeOut
	origProtocol := options.origProtocol
	log.Debug(options.origProtocol)
	log.Debug(url)
	faviconHash, headerContent, urlContent, resultContent := scan(url, proxy, timeOut, origProtocol)
	output(r.options.Output, analyze(faviconHash, headerContent, urlContent, resultContent))
}

// scan 扫描单个url
// return icon指纹 响应头列表 响应体列表 结果样例
func scan(url string, proxy string, timeOut int, origProtocol string) (faviconHash string, headerContent []http.Header, urlContent []string, resultContent *Result) {
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

	// 获得网站内容
	for k, v := range urls { //range returns both the index and value
		log.Debug("GetContent: " + v)
		header, content, result, err := HttpReq(v, timeOut, proxy)
		// 如果是https，则拥有一次重试机会，避免协议不匹配问题
		if err != nil && !retried && origProtocol == "https" {
			log.Debug(fmt.Sprintf("request url %s error", v))
			retried = true
			goto retry
		} else {
			urlContent = append(urlContent, content)
			headerContent = append(headerContent, header)
			if k == 0 {
				resultContent = &result
			}
		}
	}

	// 以非重定向的方式获得网站内容
	// 同时分析两种情况，避免重定向跳转导致获得失败，避免反向代理和CDN导致收集面缩小
	log.Debug("GetContent with NoRedirect: " + indexUrl)
	header, content, _, err := HttpReqWithNoRedirect(indexUrl, timeOut, proxy)
	if err != nil {
		log.Error(fmt.Sprintf("%s", err))
	} else {
		urlContent = append(urlContent, content)
		headerContent = append(headerContent, header)
	}

	// 获取icon指纹
	log.Debug("GetIconHash: " + faviconUrl)
	faviconHash, err = GetFaviconHash(faviconUrl, timeOut, proxy)
	if err == nil && faviconHash != "" {
		log.Debug(fmt.Sprintf("GetIconHash: %s %s success", faviconUrl, faviconHash))
	} else if err != nil {
		log.Warn(fmt.Sprintf("GetIconHash Error %s", err))
	}
	return
}

// output 输出处理
func output(Output string, resp Result) {
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

	row := fmt.Sprintf("[%s]", Bold(Green(resp.URL)))
	row += fmt.Sprintf("[%d]", Bold(Cyan(resp.StatusCode)))
	row += fmt.Sprintf("[%s]", Bold(Magenta(resp.Title)))
	row += fmt.Sprintf("[%s]", Bold(Red(finger)))
	row += fmt.Sprintf("[%s]", technology)

	raw := fmt.Sprintf("[%s] [%d] [%s] [%s] [%s]", resp.URL, resp.StatusCode, resp.Title, finger, technology)

	if finger != ""{
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
	if Output != "" {
		var err error
		f, err = os.Create(Output)
		if err != nil {
			log.Fatal(fmt.Sprintf("Could not create output file '%s': %s", Output, err))
		}
		f.WriteString(raw + "\n")
		defer f.Close() //nolint
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

	for _,f = range configs {
		for _,p := range f.Fingerprint{
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

func detect(name string, faviconHash string, headerContent []http.Header, indexContent []string, mf MetaFinger) (rs []string){
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

func detectKeywords(headerContent []http.Header, indexContent []string, mf MetaFinger) bool{
	// if keyword in body
	if mf.Location == "body" {
		for _, k := range mf.Keyword {
			for _, c := range indexContent {
				// make sure Both conditions are satisfied
				// TODO  && mf.StatusCode == header
				if !strings.Contains(c, k){
					return false
				}
			}
		}
	}

	if mf.Location == "header"{
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