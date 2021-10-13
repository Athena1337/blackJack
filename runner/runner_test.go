package runner

import (
	"blackJack/config"
	"testing"
	"time"
)

func TestRunner(t *testing.T) {
	options := &config.Options{
		TargetUrl: "google.com",
		TimeOut:   30 * time.Second,
		Threads:   50,
		RetryMax:  5,
		IsDebug: true,
	}
	r, _ := New(options)
	r.CreateRunner()
}
