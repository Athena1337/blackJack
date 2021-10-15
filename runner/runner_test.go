package runner

import (
	"blackJack/config"
	"testing"
)

func TestRunner(t *testing.T) {
	options := config.DefaultOption
	options.TargetUrl = "https://mmwater.mmzqoa.net/"
	r, _ := New(&options)
	r.CreateRunner()
}
