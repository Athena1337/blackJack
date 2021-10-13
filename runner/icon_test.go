package runner

import (
	"blackJack/config"
	"testing"
	"time"
)

func TestGetFaviconHash(t *testing.T){
	options := &config.Options{
		TargetUrl: "google.com",
		TimeOut:  30 * time.Second,
		Threads: 50,
		RetryMax: 5,
	}

	r := &Runner{
		options: options,
	}

	url := "https://bing.com"
	hash, err := r.GetFaviconHash(url)
	if err != nil && hash != "1693998826"{
		t.Errorf("Analyze test error")
	}
}