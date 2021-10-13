package config

import (
	"blackJack/log"
	"blackJack/utils"
	"flag"
	"time"
)

type Options struct{
	TargetUrl    string
	IndexUrl     string
	FaviconUrl   string
	ErrorUrl     string
	OrigProtocol string
	Urls         []string
	UrlFile      string
	IsDebug      bool
	TimeOut      time.Duration
	Threads      int
	RetryMax     int
	Output       string
	JSONOutput   bool
	Proxy        string
}

func ParseOptions() *Options {
	options := &Options{}
	flag.StringVar(&options.TargetUrl,"u", "", "single target url")
	flag.StringVar(&options.UrlFile, "l", "", "the list file contain mutilple target url")
	flag.BoolVar(&options.IsDebug, "d", false, "enable debug mode")
	flag.DurationVar(&options.TimeOut, "time", 30 * time.Second, "request timeout")
	flag.IntVar(&options.Threads, "t",  50, "request thread, default 50")
	flag.IntVar(&options.RetryMax, "r", 5, "Max Retry attempts")
	flag.StringVar(&options.Output, "o", "", "output file")
	//flag.StringVar(&options.origProtocol, "p", "", "http/https protocal")
	flag.StringVar(&options.Proxy, "p", "", "http proxy ,Ex: http://127.0.0.1:8080")
	flag.StringVar(&options.FaviconUrl, "i", "", "Analyse target favicon fingerprint")
	flag.Parse()
	options.validateOptions()
	return options
}

func (options *Options) validateOptions() {
	if options.UrlFile != "" && !utils.FileNameIsGlob(options.UrlFile) && !utils.FileExists(options.UrlFile) {
		log.Fatal("File does not exist!")
	}

	if options.TargetUrl == "" && options.UrlFile == "" {
		log.Error("Usage: -h to see the help info")
		log.Fatal("Require target url or url file!")
	}

	if options.IsDebug {
		log.Debug("Enable Debug mode")
	}

	if options.OrigProtocol == ""{
		options.OrigProtocol = "https||http"
	}
}