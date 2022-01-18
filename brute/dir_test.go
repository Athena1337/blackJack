package brute

import (
	"blackJack/config"
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
	d.Start()
}
