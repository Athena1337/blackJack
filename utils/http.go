package utils

import (
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"github.com/t43Wiu6/tlog"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

func TrimHtml(src string) string {
	//将HTML标签全转换成小写
	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllStringFunc(src, strings.ToLower)
	//去除STYLE
	re, _ = regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
	src = re.ReplaceAllString(src, "")
	//去除SCRIPT
	re, _ = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	src = re.ReplaceAllString(src, "")
	//去除所有尖括号内的HTML代码，并换成换行符
	re, _ = regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllString(src, "\n")
	//去除连续的换行符
	re, _ = regexp.Compile("\\s{2,}")
	src = re.ReplaceAllString(src, "\n")
	return strings.TrimSpace(src)
}

func DumpHttpResponse(resp *http.Response) []byte{
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("func httpDump read resp body err: %v", err)
	} else {
		acceptEncode := resp.Header["Content-Encoding"]
		var respBodyBin bytes.Buffer
		w := bufio.NewWriter(&respBodyBin)
		w.Write(respBody)
		w.Flush()
		for _, compress := range acceptEncode {
			switch compress {
			case "gzip":
				r, err := gzip.NewReader(&respBodyBin)
				if err != nil {
					log.Debugf("gzip reader err: %v", err)
				} else {
					defer r.Close()
					respBody, _ = ioutil.ReadAll(r)
				}
				break
			case "deflate":
				r := flate.NewReader(&respBodyBin)
				defer r.Close()
				respBody, _ = ioutil.ReadAll(r)
				break
			}
		}
		// log.Debugf("DumpContent %s : %s", resp.Request.URL.String() ,string(respBody))
	}
	return respBody
}