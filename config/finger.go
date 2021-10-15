package config

import (
	. "blackJack/log"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
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

// 获取当前执行文件绝对路径
func getCurrentAbPathByCaller() string {
	ex, err := os.Executable()
	if err != nil {
		Fatal("Cann't Read Config File")
	}
	exPath := filepath.Dir(ex)
	return exPath
}

func LoadFinger() (error, Config) {
	var configs Config
	filePath := getCurrentAbPathByCaller()
	jsonPath := path.Join(filePath, "finger.json")
	dat, err := ioutil.ReadFile(jsonPath)

	if err != nil {
		Fatal(fmt.Sprintf("Cann't Read Config File: %s %s", filePath, jsonPath))
		return err, configs
	}
	err = json.Unmarshal(dat, &configs)
	if err != nil {
		Error(fmt.Sprintf("%s", err))
		return err, configs
	}
	a := 0
	for _, k := range configs.Rules {
		a = a + len(k.Fingerprint)
	}
	Info(fmt.Sprintf("Totaly load finger %d 's", a))
	return nil, configs
}

func SetEnv(isDebug bool) {
	DEBUG = isDebug
}
