package main

import (
	"github.com/Athena1337/blackJack/cli"
	"github.com/Athena1337/blackJack/runner"
	log "github.com/t43Wiu6/tlog"
)

func main() {
	runner.ShowBanner()
	cli.Parse()
	log.Infof("Done, Good Luck !")
}
