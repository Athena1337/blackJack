package utils

import (
	"bytes"
	"fmt"
	"github.com/t43Wiu6/tlog"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	. "net/http"
	"regexp"
	"strings"
)

var (
	cutset                       = "\n\t\v\f\r"
	reTitle       *regexp.Regexp = regexp.MustCompile(`(?im)<\s*title.*>(.*?)<\s*/\s*title>`)
	reTitle2      *regexp.Regexp = regexp.MustCompile(`document\.title=(.+);`)
	reContentType *regexp.Regexp = regexp.MustCompile(`(?im)\s*charset="(.*?)"|charset=(.*?)"\s*`)
)

// ExtractTitle from a response
func ExtractTitle(r *Response) (title string, body_content string) {
	bodyBak, err := ioutil.ReadAll(r.Body)
	// Try to parse the DOM
	titleDom, err, body := getTitleWithDom(r)
	// fix body for re extract title
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBak))
	// convert io.ReadCloser to string
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	newStr := buf.String()
	body_content = ByteArrayToString(body)
	// In case of error fallback to regex
	if err != nil {
		for _, match := range reTitle.FindAllString(newStr, -1) {
			title = match
			break
		}
	} else {
		title = renderNode(titleDom)
	}

	title = html.UnescapeString(trimTitleTags(title))
	// remove unwanted chars
	title = strings.TrimSpace(strings.Trim(title, cutset))

	// support overwrite title with js
	if title == ""{
		for _, match := range reTitle2.FindAllString(newStr, -1) {
			title = match
			break
		}
		title = strings.Replace(title,"document.title","",-1)
		title = strings.Replace(title,"=","",-1)
		title = strings.Replace(title,"'","",-1)
		title = strings.Replace(title,";","",-1)
		title = strings.Replace(title,"\"","",-1)
	}

	// Non UTF-8
	contentTypes := r.Header.Values("Content-Type")
	if len(contentTypes) > 0 {
		contentType := strings.Join(contentTypes, ";")

		// special cases
		if strings.Contains(strings.ToLower(contentType), "charset=gb2312") ||
			strings.Contains(strings.ToLower(contentType), "charset=gbk") {
			titleUtf8, err := Decodegbk([]byte(title))
			if err != nil {
				return
			}
			return string(titleUtf8), body_content
		}

		// Content-Type from head tag
		var match = reContentType.FindSubmatch(body)
		var mcontentType = ""
		if len(match) != 0 {
			for i, v := range match {
				if string(v) != "" && i != 0 {
					mcontentType = string(v)
				}
			}
			mcontentType = strings.ToLower(mcontentType)
		}
		if strings.Contains(mcontentType, "gb2312") || strings.Contains(mcontentType, "gbk") {
			titleUtf8, err := Decodegbk([]byte(title))
			if err != nil {
				return
			}
			return string(titleUtf8), body_content
		}
	}
	return //nolint
}

func getTitleWithDom(r *Response) (*html.Node, error, []byte) {
	var title *html.Node
	var crawler func(*html.Node)
	crawler = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "title" {
			title = node
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			crawler(child)
		}
	}
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Errorf("Error reading body: %v", err)
	}
	htmlDoc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err, body
	}
	crawler(htmlDoc)
	if title != nil {
		return title, nil, body
	}
	return nil, fmt.Errorf("title not found"), body
}

func renderNode(n *html.Node) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, n) //nolint
	return buf.String()
}

func trimTitleTags(title string) string {
	// trim <title>*</title>
	titleBegin := strings.Index(title, ">")
	titleEnd := strings.Index(title, "</")
	if titleEnd < 0 || titleBegin < 0 {
		return title
	}
	return title[titleBegin+1 : titleEnd]
}
