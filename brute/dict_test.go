package brute

import (
	"testing"
)

func TestPrepareDict(t *testing.T) {
	dicts, err  := PrepareDict()
	if err != nil && len(dicts) > 0{
		t.Error("PrepareDict Error")
	}
}

func TestDownloadDict(t *testing.T) {
	err := DownloadDict()
	if err != nil {
		t.Error("DownloadDict Error")
	}
}