package brute

import (
	"blackJack/config"
	"blackJack/utils"
	"github.com/t43Wiu6/tlog"
	"os"
	"path/filepath"
)

func PrepareDict() (dicts []string, err error){
	dict, err := loadDict(config.Dict_mini)
	if err != nil {
		return
	}
	dicts = append(dicts, dict...)

	dict, err = loadDict(config.Api)
	if err != nil {
		return
	}
	dicts = append(dicts, dict...)

	dict, err = loadDict(config.File)
	if err != nil {
		return
	}
	dicts = append(dicts, dict...)

	log.Debugf("Generate dict %d 's", len(dicts))
	return
}

func loadDict(name string)(dict []string, err error) {
	filePath, err := utils.GetCurrentAbPathByCaller()
	if err == nil{
		home, _ := os.UserHomeDir()
		filePath = filepath.Join(home, ".blackJack", name)
		dict, err = utils.ReadFile(filePath)
		if err != nil {
			log.Warnf("dict not found, unable to read dict file: %s", filePath)
			return
		}
	}
	return
}