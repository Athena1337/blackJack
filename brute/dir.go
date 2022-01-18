package brute

import (
	"blackJack/config"
	"blackJack/utils"
	"crypto/tls"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/remeh/sizedwaitgroup"
	"github.com/t43Wiu6/tlog"
	"gopkg.in/go-dedup/simhash.v2"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type DirBrute struct {
	sync.Mutex
	IndexUrl      string
	ErrorUrl      string
	Options       *config.Options
	client        *http.Client
	errorContent  []byte
	indexContent []byte
	list          []string
	ua            string
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

// compare 根据simhash算法计算相似性，不相似即返回false
func compare(resp1 []byte, resp2 []byte) (bool, error) {
	// TODO
	sh := simhash.NewSimhash()

	data1 := utils.TrimHtml(string(resp1))
	data2 := utils.TrimHtml(string(resp2))

	hashes1 := sh.GetSimhash(sh.NewWordFeatureSet([]byte(data1)))
	hashes2 := sh.GetSimhash(sh.NewWordFeatureSet([]byte(data2)))
	rs := simhash.Compare(hashes1, hashes2)

	// 差异大于10，认定为不同页面
	if rs > 10 {
		return false, nil
	}
	return true, nil
}

func (dir *DirBrute) request(dict string, wg *sizedwaitgroup.SizedWaitGroup) {
	defer wg.Done()

	dictUrl := fmt.Sprintf("%s%s", dir.IndexUrl, dict)
	log.Debugf("[DirBrute] Get url : %s", dictUrl)
	req, err := http.NewRequest("GET", dictUrl, nil)
	if err != nil {
		return
	}

	// set default user agent
	req.Header.Add("User-Agent", dir.ua)

	resp, err := dir.client.Do(req)
	if err != nil {
		return
	}

	// 不相似
	dir.Lock()
	defer dir.Unlock()
	data := utils.DumpHttpResponse(resp)

	if flag, err := compare(data, dir.errorContent); err == nil && !flag {
		if flag2, err := compare(data, dir.indexContent); err == nil && !flag2{
			template := fmt.Sprintf("[!] [%d] - %d - %s", resp.StatusCode, resp.ContentLength, req.URL.String())
			fmt.Println(template)
			dir.list = append(dir.list, template)
		}
	} else if err != nil {
		log.Errorf("Compare content error: %v", err)
	}
	log.Debugf("The similarity of the page %s to NotFoundPage is less than 3", resp.Request.URL.String())
}

func (dir *DirBrute) detectPseudo() (err error) {
	resp, err := dir.client.Get(dir.ErrorUrl)
	if err != nil {
		return
	}
	data := utils.DumpHttpResponse(resp)
	dir.errorContent = data

	resp, err = dir.client.Get(dir.IndexUrl)
	if err != nil {
		return
	}
	data = utils.DumpHttpResponse(resp)
	dir.indexContent = data
	return
}

//func (dir *DirBrute) Output() {
//	return dir.list
//}

func (dir *DirBrute) Start(output chan []string) {
	dir.init()
	dicts, err := PrepareDict()
	if err != nil {
		log.Errorf("PrepareDict Failed: %v", err)
		return
	}

	err = dir.detectPseudo()
	if err != nil {
		log.Errorf("Request Pseudo 404 page Failed: %v", err)
		return
	}

	wg := sizedwaitgroup.New(dir.Options.Threads) // set threads , 50 by default
	for _, dict := range dicts {
		wg.Add()
		go dir.request(dict, &wg)
	}

	wg.Wait()
	output <- dir.list
}
