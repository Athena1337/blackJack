package runner

import (
	"blackJack/config"
	"testing"
)

func TestHttpReqWithNoRedirect(t *testing.T) {
	options := config.DefaultOption
	r, _ := New(&options)
	resp, err := r.Request("GET","https://bing.com", false)
	if err != nil {
		t.Errorf("HttpReqWithNoRedirect test error")
	}
	if err != nil || resp.StatusCode != 301{
		t.Errorf("HttpReqWithNoRedirect test error")
	}
}


func TestHttpReq(t *testing.T) {
	options := config.DefaultOption
	r, _ := New(&options)
	resp, err := r.Request("GET","https://bing.com", true)
	if err != nil {
		t.Errorf("HttpReq test error")
	}
	if err != nil || resp.StatusCode != 200 || resp.Title != "必应"{
		t.Errorf("HttpReq test error")
	}
}

