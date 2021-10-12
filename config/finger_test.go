package config

import (
	"testing"
)

func TestLoadFinger(t *testing.T){
	err, finger := LoadFinger()
	if err != nil && len(finger.Rules) == 0{
		t.Errorf("LoadFinger test error")
	}
}

