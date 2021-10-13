package main

import (
	"blackJack/config"
	"blackJack/log"
	"blackJack/runner"
	"fmt"
)

func main() {
	runner.ShowBanner()
	options := config.ParseOptions()

	var err error
	r, err := runner.New(options)
	if err != nil{
		log.Fatal(err.Error())
	}

	if options.FaviconUrl != ""{
		// 分析网站icon指纹
		faviconHash, err := r.GetFaviconHash(options.FaviconUrl)
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