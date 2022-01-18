package config

import (
	"blackJack/utils"
	"encoding/json"
	"github.com/t43Wiu6/tlog"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type MetaFinger struct {
	Method     string
	Location   string
	StatusCode int
	Keyword    []string
}

type Finger struct {
	Name        string
	Fingerprint []MetaFinger
}

type Config struct {
	Rules []Finger
}

func LoadFinger() (configs Config, err error) {
	filePath, err := utils.GetCurrentAbPathByCaller()
	if err == nil{
		home, _ := os.UserHomeDir()
		filePath = filepath.Join(home, "blackJack", "finger.json")
		dat, errs := ioutil.ReadFile(filePath)
		if errs != nil {
			log.Warnf("finger.json not found, unable to read config file: %s", filePath)
			return configs, errs
		}
		err = json.Unmarshal(dat, &configs)
		if err != nil {
			log.Errorf("%s", err)
			return
		}
	}
	a := 0
	for _, k := range configs.Rules {
		a = a + len(k.Fingerprint)
	}
	log.Infof("Totaly load finger %d 's", a)
	return
}

func DownloadFinger()(err error){
	fingerUrl := "https://raw.githubusercontent.com/Athena1337/blackJack/main/finger.json"
	home, _ := os.UserHomeDir()
	fingerPath := filepath.Join(home, "blackJack", "finger.json")
	if !utils.FolderExists(filepath.Join(home, "blackJack")){
		err = os.Mkdir( filepath.Join(home, "blackJack"), os.ModePerm)
		if err != nil {
			return
		}
	}
	resp, err := http.Get(fingerUrl)
	if err != nil {
		log.Error("download failed, please check you network...")
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if utils.FileExists(fingerPath){
		err := os.Remove(fingerPath)
		if err != nil {
			return err
		}
	}
	err = ioutil.WriteFile(fingerPath, data, 0666)
	if err != nil {
		log.Warn("unable to write config file")
	}
	return
}

func SetEnv(isDebug bool) {
	log.DEBUG = isDebug
}
