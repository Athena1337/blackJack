package runner

import (
	"blackJack/libs"
	"testing"
)

func TestRunner(t *testing.T){
	utils.SetEnv(true)
	_, FINGER= utils.LoadFinger()
	url := "google.com"
	faviconHash, headerContent, urlContent, resultContent := scan(url, "", 50, "https")
	result := analyze(faviconHash, headerContent, urlContent, resultContent)
	output("",result)
	if result.Title != "Google"{
		t.Errorf("Analyze test error")
	}
}
