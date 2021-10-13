package runner

import (
	"blackJack/log"
	"blackJack/utils"
	"flag"
	"time"
)

type Options struct{
	targetUrl    string
	indexUrl     string
	FaviconUrl   string
	errorUrl     string
	origProtocol string
	urls         []string
	urlFile      string
	isDebug      bool
	TimeOut      time.Duration
	Threads      int
	RetryMax	 int
	Output       string
	JSONOutput   bool
	Proxy        string
}

func ParseOptions() *Options {
	options := &Options{}
	flag.StringVar(&options.targetUrl ,"u", "", "single target url")
	flag.StringVar(&options.urlFile, "l", "", "the list file contain mutilple target url")
	flag.BoolVar(&options.isDebug, "d", false, "enable debug mode")
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
	if options.urlFile != "" && !utils.FileNameIsGlob(options.urlFile) && !utils.FileExists(options.urlFile) {
		log.Fatal("File does not exist!")
	}

	if options.targetUrl == "" && options.urlFile == "" {
		log.Error("Usage: -h to see the help info")
		log.Fatal("Require target url or url file!")
	}

	if options.isDebug{
		log.Debug("Enable Debug mode")
	}

	if options.origProtocol == ""{
		options.origProtocol = "https||http"
	}
}