package runner

import (
	"github.com/Athena1337/blackJack/utils"
	"bytes"
	"context"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	pdhttputil "github.com/projectdiscovery/httputil"
	"github.com/t43Wiu6/tlog"
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
	URL           *url.URL
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
	request, err := newRequest(method, url)
get_response:
	if err != nil {
		return
	}
	var do *http.Response
	if redirect {
		do, err = r.client.Do(request)
	} else {
		do, err = r.noRedirectClient.Do(request)
	}

	if err != nil {
		return
	}

	bodyBak, err := ioutil.ReadAll(do.Body)
	do.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBak))

	resp.Title, _ = utils.ExtractTitle(do)
	// 拒绝one-shot, 还原Body, 后面还要ExtractTitle
	do.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBak))

	var respbody []byte
	// websockets don't have a readable body
	if do.StatusCode != http.StatusSwitchingProtocols {
		respbody, err = ioutil.ReadAll(do.Body)
		if err != nil {
			return
		}
	}
	// 后面还要DumpRaw
	do.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBak))

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
	resp.Headers = do.Header.Clone()
	resp.Host = request.Host
	resp.URL = request.URL

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

	return
}

// newRequest from url
func newRequest(method, targetURL string) (req *http.Request, err error) {
	req, err = http.NewRequest(method, targetURL, nil)
	if err != nil {
		return
	}

	// set default user agent
	req.Header.Add("User-Agent", utils.GetUserAgent())
	// 检测shiro指纹
	req.Header.Add("Cookie", fmt.Sprintf("rememberMe=%s", utils.RandStringBytes(67)))
	// set default encoding to accept utf8
	req.Header.Add("Accept-Charset", "utf-8")
	return
}

func RequestByChrome(url string, screenPath string) (title []string, err error) {
	// create context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.WindowSize(1980, 1080),
		chromedp.Flag("ignore-certificate-errors", "1"),
		chromedp.Flag("enable-features", "NetworkService"),
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36`),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 创建超时时间
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	// 缓存对象
	var buf []byte
	// 运行截屏
	if err = chromedp.Run(ctx, fullScreenshot(url, 100, &buf)); err != nil {
		return
	}
	// 保存文件
	if err = ioutil.WriteFile(screenPath, buf, 0644); err != nil {
		log.Error(err.Error())
	}
	return
}

// fullScreenshot 全屏截图
func fullScreenshot(url string, quality int64, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		////延时：等待有些页面有js自动跳转，待js跳转后再执行截图操作
		chromedp.Sleep(5 * time.Second),
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			*res, err = page.CaptureScreenshot().WithQuality(quality).WithClip(&page.Viewport{
				X:      0,
				Y:      0,
				Width:  1980,
				Height: 1080,
				Scale:  1,
			}).Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}),
	}
}
