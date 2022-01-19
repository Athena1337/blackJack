package main

import (
	"blackJack/cli"
	"blackJack/runner"
	log "github.com/t43Wiu6/tlog"
)

func main() {
	runner.ShowBanner()
	cli.Parse()
	log.Infof("Done, Good Luck !")
}
