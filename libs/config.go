package libs

import (
	. "blackJack/log"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

// 获取当前执行文件绝对路径（go run）
func getCurrentAbPathByCaller() string {
	ex, err := os.Executable()
	if err != nil {
		Fatal("Cann't Read Config File")
	}
	exPath := filepath.Dir(ex)
	return exPath
}

func LoadFinger() (error,Config){
	var configs Config
	filePath := getCurrentAbPathByCaller()
	jsonPath := path.Join(filePath, "../","finger.json")
	dat, err := ioutil.ReadFile(jsonPath)

	if err != nil{
		Fatal("Cann't Read Config File")
		return err, configs
	}
	err = json.Unmarshal(dat, &configs)
	if err != nil {
		Error(fmt.Sprintf("%s",err))
		return err, configs
	}
	Info(fmt.Sprintf("Totaly load finger %d 's", len(configs.Fingerprint)))
	return nil, configs
}

func SetEnv(isDebug bool){
	DEBUG = isDebug
}

