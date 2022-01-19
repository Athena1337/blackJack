package brute

import (
	"blackJack/config"
	"blackJack/utils"
	"crypto/tls"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pterm/pterm"
	"github.com/remeh/sizedwaitgroup"
	"github.com/t43Wiu6/tlog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type DirStatus struct {
	DoneJob int
	AllJob  int
}

type DirBrute struct {
	sync.Mutex
	IndexUrl string
	ErrorUrl string
	Options  *config.Options
	client   *http.Client
	Simhash  []uint64
	list     []string
	ua       string
}

func (dir *DirBrute) init() {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = dir.Options.RetryMax

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

	if dir.Options.Proxy != "" {
		proxyUrl, err := url.Parse(dir.Options.Proxy)
		if err != nil {
			log.Error("Proxy Url Can Not Identify, Droped")
		} else {
			transport.Proxy = http.ProxyURL(proxyUrl)
		}
	}
	dir.client = retryClient.StandardClient() // *http.Client
	dir.client.Timeout = dir.Options.TimeOut
	dir.client.Transport = transport
	dir.ua = utils.GetUserAgent()
}

func (dir *DirBrute) Start(output chan []string, printer *pterm.SpinnerPrinter, dirWg *sizedwaitgroup.SizedWaitGroup, process *DirStatus) {
	defer dirWg.Done()

	log.Debugf("[DirBrute] Start to Brute Force %s", dir.IndexUrl)
	dir.init()
	dicts, err := PrepareDict()
	if err != nil {
		output <- dir.list
		log.Errorf("PrepareDict Failed: %v", err)
		return
	}

	err = dir.detectPseudo()
	if err != nil {
		output <- dir.list
		log.Errorf("Request Pseudo 404 page Failed: %v", err)
		return
	}

	wg := sizedwaitgroup.New(dir.Options.Threads) // set threads , 50 by default
	for _, dict := range dicts {
		wg.Add()
		go dir.request(dict, &wg, printer, process)
	}
	wg.Wait()
	output <- dir.list
}

func (dir *DirBrute) request(dict string, wg *sizedwaitgroup.SizedWaitGroup, printer *pterm.SpinnerPrinter, process *DirStatus) {
	defer wg.Done()

	var dictUrl string
	if strings.HasSuffix(dir.IndexUrl, "/") {
		dictUrl = fmt.Sprintf("%s%s", dir.IndexUrl, dict)
	} else {
		dictUrl = fmt.Sprintf("%s/%s", dir.IndexUrl, dict)
	}

	log.Debugf("[DirBrute] Get url : %s", dictUrl)
	req, err := http.NewRequest("GET", dictUrl, nil)
	if err != nil {
		return
	}
	printer.UpdateText(fmt.Sprintf("[DirBrute] [%d/%d] Brute Force Target : %s Path : %s", process.DoneJob, process.AllJob, dir.IndexUrl, dict))

	// set default user agent
	req.Header.Add("User-Agent", dir.ua)

	resp, err := dir.client.Do(req)
	if err != nil {
		return
	}
	data := utils.DumpHttpResponse(resp)

	// nginx \ tengine 反代
	if strings.Contains(string(data), "403 Forbidden") && resp.StatusCode == 403 {
		return
	}
	if resp.StatusCode < 300 && resp.StatusCode != 200 {
		return
	}
	if resp.StatusCode > 399 && resp.StatusCode != 500 && resp.StatusCode != 403 {
		return
	}

	// 不相似
	h := utils.GetHash(data)
	if !dir.isInBlackList(h) {
		template := fmt.Sprintf("[!] [%d] [%d] - %s", resp.StatusCode, len(data), req.URL.String())
		dir.list = append(dir.list, template)
	}
	log.Debugf("The similarity of the page %s to NotFoundPage is less than 10", resp.Request.URL.String())
}

// isInBlackList 根据simhash算法计算相似性，不相似即返回false
func (dir *DirBrute) isInBlackList(h uint64) bool {
	// TODO
	for _, hash := range dir.Simhash {
		// hash类似， 认为是两个近似页面
		if utils.IsSimilarHash(hash, h) {
			// 追加
			dir.Lock()
			dir.Simhash = append(dir.Simhash, h)
			dir.Unlock()
			return true
		}
	}
	return false
}

// detectPseudo 计算主页和不存在页面的simhash 用于检测
func (dir *DirBrute) detectPseudo() (err error) {
	dir.Lock()
	defer dir.Unlock()
	resp, err := dir.client.Get(dir.ErrorUrl)
	if err != nil {
		return
	}
	data := utils.DumpHttpResponse(resp)
	hash := utils.GetHash(data)
	dir.Simhash = append(dir.Simhash, hash)

	resp, err = dir.client.Get(dir.IndexUrl)
	if err != nil {
		return
	}
	data = utils.DumpHttpResponse(resp)
	hash = utils.GetHash(data)
	dir.Simhash = append(dir.Simhash, hash)
	return
}
