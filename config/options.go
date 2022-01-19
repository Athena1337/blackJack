package config

import (
	"github.com/t43Wiu6/tlog"
	"os"
	"sync"
	"time"
)

func SetEnv(isDebug bool) {
	log.DEBUG = isDebug
}

type OutputFile struct {
	sync.Mutex
	File *os.File
}

type Options struct {
	TargetUrl      string
	IndexUrl       string
	FaviconUrl     string
	ErrorUrl       string
	OrigProtocol   string
	Urls           []string
	UrlFile        string
	IsDebug        bool
	EnableDirBrute bool
	TimeOut        time.Duration
	Threads        int
	RetryMax       int
	Output         string
	JSONOutput     bool
	Proxy          string
	OutputFile     OutputFile
}

var DefaultOption = Options{
	TargetUrl: "https://google.com",
	IsDebug:   true,
	TimeOut:   30 * time.Second,
	Threads:   50,
	RetryMax:  5,
}
