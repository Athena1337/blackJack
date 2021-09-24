package runner

import (
	"testing"
)

func TestGetFaviconHash(t *testing.T){
	url := "https://bing.com"
	hash, err := GetFaviconHash(url,50,"")
	if err != nil && hash != "1693998826"{
		t.Errorf("Analyze test error")
	}
}