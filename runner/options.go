package runner

import (
	"blackJack/libs"
	"blackJack/log"
	"flag"
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
	TimeOut      int
	Threads      int
	Output       string
	JSONOutput   bool
	Proxy        string
}

func ParseOptions() *Options {
	options := &Options{}
	flag.StringVar(&options.targetUrl ,"u", "", "single target url")
	flag.StringVar(&options.urlFile, "l", "", "the list file contain mutilple target url")
	flag.BoolVar(&options.isDebug, "d", false, "enable debug mode")
	flag.IntVar(&options.TimeOut, "time", 5, "request timeout")
	flag.IntVar(&options.Threads, "t",  50, "request thread, default 50")
	flag.StringVar(&options.Output, "o", "", "output file")
	//flag.StringVar(&options.origProtocol, "p", "", "http/https protocal")
	flag.StringVar(&options.Proxy, "p", "", "http proxy ,Ex: http://127.0.0.1:8080")
	flag.StringVar(&options.FaviconUrl, "i", "", "Analyse target favicon fingerprint")
	flag.Parse()
	options.validateOptions()
	return options
}

func (options *Options) validateOptions() {
	if options.urlFile != "" && !libs.FileNameIsGlob(options.urlFile) && !libs.FileExists(options.urlFile) {
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