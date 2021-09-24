package main

import (
	"blackJack/log"
	"blackJack/runner"
	"fmt"
)

func main() {
	runner.ShowBanner()
	options := runner.ParseOptions()
	r := runner.Init(options)
	if options.FaviconUrl != ""{
		// 分析网站icon指纹
		faviconHash, err := runner.GetFaviconHash(options.FaviconUrl,  options.TimeOut, options.Proxy)
		if err != nil && faviconHash != ""{
			log.Info(fmt.Sprintf("faviconHash: %s", faviconHash))
		}else{
			log.Info(fmt.Sprintf("%s",err))
		}
	}else{
		// 创建指纹扫描任务
		r.CreateRunner()
	}
}