package runner

import (
	"blackJack/config"
	"testing"
)

func TestRunner(t *testing.T) {
	//options := config.DefaultOption
	//r, _ := New(&options)
	//r.CreateRunner()

	options := config.DefaultOption
	options.TargetUrl = ""
	options.UrlFile = "C:/Users/sherd/Desktop/urls.txt"
	r, _ := New(&options)
	r.CreateRunner()
}
