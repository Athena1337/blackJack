package runner

import (
	"blackJack/log"
	"blackJack/utils"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

func GetFaviconHash(HashURL string, timeOut int, proxy string) (string, error){
	content, err := HttpReqForIcon(HashURL, timeOut, proxy)
	if err != nil {
		return  "", fmt.Errorf("get favcion error")
	}else{
		ret := utils.Mmh3Hash32(utils.StandBase64(content))
		return ret, nil
	}
}

func HttpReqForIcon(requrl string, timeOut int, proxy string) ([]byte, error) {
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
		return nil, err
	}
	req.Header.Set("User-Agent", utils.GetUserAgent())
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}