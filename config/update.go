package config

import (
	"blackJack/utils"
	"github.com/t43Wiu6/tlog"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

var (
	Dir_bigger = "dir_bigger.txt" // 15000
	Dict_mini  = "dir_mini.txt"   // 2000
	Dir        = "dir.txt"        // 5000
	Api        = "api/per_root_api.txt"
	File       = "file/per_root_file.txt"
	Rfile      = "file/per_dir_file_replace.txt"
	Asp_dir    = "asp/asp_dir.txt"
	Asp_file   = "asp/asp_file.txt"
	Aspx_dir   = "aspx/aspx_dir.txt"
	Aspx_file  = "aspx/aspx_file.txt"
	Java_dir   = "java/action_dir.txt"
	Java_file  = "java/action_file.txt"
	Jsp_dir    = "jsp/jsp_dir.txt"
	Jsp_file   = "jsp/jsp_file.txt"
	Php_dir    = "php/php_dir.txt"
	Php_file = "php/php_file.txt"
)

func DownloadAll() (err error){
	log.Warn("try to download from github...")
	err = DownloadDictFile()
	if err != nil {
		return
	}

	err = DownloadFinger()
	return
}

func DownloadFinger()(err error){
	fingerUrl := "https://raw.githubusercontent.com/Athena1337/blackJack/main/finger.json"
	home, _ := os.UserHomeDir()
	fingerPath := filepath.Join(home, ".blackJack", "finger.json")
	if !utils.FolderExists(filepath.Join(home, ".blackJack")){
		err = os.Mkdir( filepath.Join(home, ".blackJack"), os.ModePerm)
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

func DownloadDictFile() (err error){
	log.Warnf("Start to Download dict from github")
	err = download(Dict_mini)
	if err != nil {
		log.Warnf("Download dict from github error: %s", err)
	}
	err = download(Api)
	if err != nil {
		log.Warnf("Download dict from github error: %s", err)
	}
	err = download(File)
	if err != nil {
		log.Warnf("Download dict from github error: %s", err)
	}
	return
}

func Check() bool{
	home, _ := os.UserHomeDir()
	if !utils.FolderExists(filepath.Join(home, ".blackJack")){
		return false
	}
	return true
}

func download(name string) (err error){
	baseUrl := "https://raw.githubusercontent.com/t43Wiu6/blackJack-Dicts/main/"
	home, _ := os.UserHomeDir()
	filePath, err := utils.GetCurrentAbPathByCaller()
	filePath = filepath.Join(home, ".blackJack", name)

	if !utils.FolderExists(filepath.Join(home, ".blackJack")){
		err = os.Mkdir( filepath.Join(home, ".blackJack"), os.ModePerm)
		if err != nil {
			return
		}
	}

	if !utils.FolderExists(filepath.Join(home, ".blackJack", "api")){
		err = os.Mkdir( filepath.Join(home, ".blackJack", "api"), os.ModePerm)
		if err != nil {
			return
		}
	}

	if !utils.FolderExists(filepath.Join(home, ".blackJack", "file")){
		err = os.Mkdir( filepath.Join(home, ".blackJack", "file"), os.ModePerm)
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