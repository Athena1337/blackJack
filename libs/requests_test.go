package libs

import (
	"testing"
)

func TestHttpReqWithNoRedirect(t *testing.T) {
	url := "https://bing.com"
	_, _, _, err := HttpReqWithNoRedirect(url,50,"")
	if err != nil {
		t.Errorf("HttpReqWithNoRedirect test error")
	}
}


func TestHttpReq(t *testing.T) {
	url := "https://bing.com"
	_, _, r, err := HttpReq(url,50,"")
	if err != nil && r.Title != "必应" {
		t.Errorf("HttpReq test error")
	}
}

