package brute

import (
	"blackJack/config"
	"github.com/pterm/pterm"
	. "github.com/t43Wiu6/tlog"
	"testing"
)

func TestStart(t *testing.T) {
	DEBUG = false
	d := &DirBrute{
		IndexUrl: "http://192.168.22.176:8080/",
		ErrorUrl: "http://192.168.22.176:8080/ljaisdkhfkjashdfkjahsdjkf",
		Options: &config.DefaultOption,
	}
	var output chan []string
	spinnerLiveText, _ := pterm.DefaultSpinner.Start("[DirBrute] Waiting to Brute Force")
	d.Start(output, spinnerLiveText)
}