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
		IndexUrl: "https://os.alipayobjects.com",
		ErrorUrl: "https://os.alipayobjects.com/ljaisdkhfkjashdfkjahsdjkf",
		Options: &config.DefaultOption,
	}
	var output chan []string
	spinnerLiveText, _ := pterm.DefaultSpinner.Start("[DirBrute] Waiting to Brute Force")
	ds := DirStatus{
		AllJob: 0,
		DoneJob: 0,
	}
	d.Start(output, spinnerLiveText, &ds)
}