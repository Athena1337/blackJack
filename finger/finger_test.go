package finger

import (
	"testing"
)

func TestLoadFinger(t *testing.T){
	finger, err := LoadFinger()
	if err != nil && len(finger.Rules) == 0{
		t.Errorf("LoadFinger test error")
	}
}