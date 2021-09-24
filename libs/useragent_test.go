package libs

import (
	"testing"
)

func TestGetUserAgent(t *testing.T) {
	ua := GetUserAgent()
	if ua == ""{
		t.Errorf("useragent error")
	}
}