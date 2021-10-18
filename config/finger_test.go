package config

import (
	"testing"
)

func TestLoadFinger(t *testing.T){
	finger, err := LoadFinger()
	if err != nil && len(finger.Rules) == 0{
		t.Errorf("LoadFinger test error")
	}
}

func TestDownloadFinger(t *testing.T){
	err := DownloadFinger()
	if err != nil{
		t.Errorf("DownloadFinger test error")
	}
}