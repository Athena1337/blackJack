package cli

import (
	"blackJack/config"
	"blackJack/runner"
	"blackJack/utils"
	"fmt"
	"github.com/t43Wiu6/tlog"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
	"time"
)

var options = &config.Options{}

func Parse(){
	app := cli.NewApp()
	app.Name = "blackJack"
	app.Usage = "Usage Menu"
	app.HideVersion = false
	app.Flags = Init()
	app.Action = Action

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func Action(c *cli.Context) error {
	r, err := runner.New(options)
	if err != nil {
		log.Fatal(err.Error())
	}

	if c.Bool("d") {
		log.Debug("Enable Debug mode")
	}
	if c.String("OrigProtocol") == ""{
		options.OrigProtocol = "https||http"
	}

	if c.String("l") != "" && !utils.FileExists(c.String("l")) {
		log.Fatal("Url File does not exist!")
	}

	if c.String("u") == "" && c.String("l") == "" && c.String("i") == ""{
		cli.ShowAppHelp(c)
	}else if c.String("u") != ""{
		r.CreateRunner()
	}else if c.String("l") != ""{
		r.CreateRunner()
	}else if c.String("i") != ""{
		// 分析网站icon指纹
		if !strings.Contains(options.FaviconUrl, "http"){
			options.FaviconUrl = fmt.Sprintf("%s://%s", "https",options.FaviconUrl)
		}
		if !strings.Contains(options.FaviconUrl, "favicon.ico"){
			options.FaviconUrl = fmt.Sprintf("%s/%s", options.FaviconUrl, "favicon.ico")
		}
		faviconHash, err := r.GetFaviconHash(options.FaviconUrl)
		if err !=nil && strings.Contains(options.FaviconUrl, "https"){
			err = nil
			options.FaviconUrl = strings.Replace(options.FaviconUrl, "https", "http", 1)
			faviconHash, err = r.GetFaviconHash(options.FaviconUrl)
		}
		if err == nil && faviconHash != "" {
			log.Infof("url: %s, faviconHash: %s", options.FaviconUrl, faviconHash)
		} else {
			log.Errorf("%s", err)
		}
	}
	return nil
}

func Init() []cli.Flag{
	flags := []cli.Flag{
		&cli.BoolFlag{
			Name:        "d, debug",
			Value:       false,
			Usage:       "Enable debug mode",
			Destination: &options.IsDebug,
		},
		&cli.StringFlag{
			Name:        "u, url",
			Value:       "",
			Usage:       "Single target url",
			Destination: &options.TargetUrl,
		},
		&cli.StringFlag{
			Name:        "l, list",
			Value:       "",
			Usage:       "The list file contain mutilple target url",
			Destination: &options.UrlFile,
		},
		&cli.IntFlag{
			Name:        "t, threads",
			Value:       50,
			Usage:       "Request thread",
			Destination: &options.Threads,
		},
		&cli.DurationFlag{
			Name:        "time",
			Value:       30 * time.Second,
			Usage:       "Request timeout",
			Destination: &options.TimeOut,
		},
		&cli.IntFlag{
			Name:        "r, retry",
			Value:       5,
			Usage:       "Max Retry attempts",
			Destination: &options.RetryMax,
		},
		&cli.StringFlag{
			Name:        "o, output",
			Value:       "",
			Usage:       "Output file",
			Destination: &options.Output,
		},
		&cli.StringFlag{
			Name:        "p, proxy",
			Value:       "",
			Usage:       "http proxy ,Ex: http://127.0.0.1:8080",
			Destination: &options.Proxy,
		},
		&cli.StringFlag{
			Name:        "i, icon",
			Value:       "",
			Usage:       "Analyse target favicon fingerprint",
			Destination: &options.FaviconUrl,
		},
		&cli.BoolFlag{
			Name:        "b, brute",
			Value:       false,
			Usage:       "Enable DirBrute for analyse target",
			Destination: &options.EnableDirBrute,
		},
	}
	return flags
}
