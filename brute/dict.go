package brute

import (
	"blackJack/utils"
	"github.com/t43Wiu6/tlog"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

var (
	dir_bigger = "dir_bigger.txt" // 15000
	dict_mini = "dir_mini.txt" // 2000
	dir = "dir.txt" // 5000
	api = "api/per_root_api.txt"
	file = "file/per_root_file.txt"
	rfile = "file/per_dir_file_replace.txt"
	asp_dir = "asp/asp_dir.txt"
	asp_file = "asp/asp_file.txt"
	aspx_dir = "aspx/aspx_dir.txt"
	aspx_file = "aspx/aspx_file.txt"
	java_dir = "java/action_dir.txt"
	java_file = "java/action_file.txt"
	jsp_dir = "jsp/jsp_dir.txt"
	jsp_file = "jsp/jsp_file.txt"
	php_dir = "php/php_dir.txt"
	php_file = "php/php_file.txt"
)

func DownloadDict()(err error){
	err = download(dict_mini)
	if err != nil {
		return
	}
	err = download(api)
	if err != nil {
		return
	}
	err = download(file)
	if err != nil {
		return
	}
	return
}

func PrepareDict() (dicts []string, err error){
	dict, err := loadDict(dict_mini)
	if err != nil {
		return
	}
	dicts = append(dicts, dict...)

	dict, err = loadDict(api)
	if err != nil {
		return
	}
	dicts = append(dicts, dict...)

	dict, err = loadDict(file)
	if err != nil {
		return
	}
	dicts = append(dicts, dict...)

	log.Infof("Generate dict %d 's", len(dicts))
	return
}

func loadDict(name string)(dict []string, err error) {
	filePath, err := utils.GetCurrentAbPathByCaller()
	if err == nil{
		home, _ := os.UserHomeDir()
		filePath = filepath.Join(home, "blackJack", name)
		dict, err = utils.ReadFile(filePath)
		if err != nil {
			log.Warnf("dict not found, unable to read dict file: %s", filePath)
			return nil, err
		}
	}
	return
}

func download(name string) (err error){
	baseUrl := "https://raw.githubusercontent.com/t43Wiu6/blackJack-Dicts/main/"
	home, _ := os.UserHomeDir()
	filePath, err := utils.GetCurrentAbPathByCaller()
	filePath = filepath.Join(home, "blackJack", name)

	if !utils.FolderExists(filepath.Join(home, "blackJack")){
		err = os.Mkdir( filepath.Join(home, "blackJack"), os.ModePerm)
		if err != nil {
			return
		}
	}

	if !utils.FolderExists(filepath.Join(home, "blackJack", "api")){
		err = os.Mkdir( filepath.Join(home, "blackJack", "api"), os.ModePerm)
		if err != nil {
			return
		}
	}

	if !utils.FolderExists(filepath.Join(home, "blackJack", "file")){
		err = os.Mkdir( filepath.Join(home, "blackJack", "file"), os.ModePerm)
		if err != nil {
			return
		}
	}

	resp, err := http.Get(baseUrl + name)
	if err != nil {
		log.Error("download failed, please check you network...")
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if utils.FileExists(filePath){
		err := os.Remove(filePath)
		if err != nil {
			return err
		}
	}
	err = ioutil.WriteFile(filePath, data, 0666)
	if err != nil {
		log.Warnf("unable to write dict file, %s", err)
	}
	return
}