package libs

import (
	"testing"
)

func TestLoadFinger(t *testing.T){
	err, finger := LoadFinger()
	if err != nil && len(finger.Fingerprint) == 0{
		t.Errorf("LoadFinger test error")
	}
}

