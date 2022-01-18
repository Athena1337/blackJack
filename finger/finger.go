package finger

import (
	"blackJack/utils"
	"encoding/json"
	"github.com/t43Wiu6/tlog"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	Rules []Finger
}

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

func LoadFinger() (configs Config, err error) {
	filePath, err := utils.GetCurrentAbPathByCaller()
	if err == nil{
		home, _ := os.UserHomeDir()
		filePath = filepath.Join(home, ".blackJack", "finger.json")
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
	log.Infof("Totally load finger %d 's", a)
	return
}
