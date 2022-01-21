package utils

import (
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"testing"
)

func TestIsSimilarHash0(t *testing.T) {
	if IsSimilarHash(9832410273512798190, 12133804658672123316) {
		fmt.Println("yes")
	} else {
		t.Fatal("")
	}
}

func TestIsSimilarHash2(t *testing.T) {
	var rs []uint64
	test := []string{
		"?wsdl",
		"asdkfjaksjdfkajsdkfjkajsdkfj",
		"",
	}
	retryClient := retryablehttp.NewClient()
	client := retryClient.StandardClient()

	for _, t := range test {
		resp, err := client.Get("http://192.168.22.176:8080/" + t)
		if err != nil {
			return
		}
		data := DumpHttpResponse(resp)
		hash1 := GetHash(data)
		fmt.Println(t, " : ", hash1)
		for _, k := range rs {
			if IsSimilarHash(hash1, k) {
				fmt.Println("yes")
				fmt.Println(k)
			}
		}
		rs = append(rs, hash1)
	}
}
