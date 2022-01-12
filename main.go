package main

import (
	"blackJack/cli"
	"blackJack/runner"
)

func main() {
	runner.ShowBanner()
	cli.Parse()
}
