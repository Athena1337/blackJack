package runner

import (
	"blackJack/config"
	"testing"
)

func TestGetFaviconHash(t *testing.T) {
	options := config.DefaultOption
	r, _ := New(&options)
	hash, err := r.GetFaviconHash("https://google.com/favicon.ico")
	if err != nil && hash != "1693998826" {
		t.Errorf("Analyze test error")
	}
}
