package utils

import (
	"gopkg.in/go-dedup/simhash.v2"
)

func GetHash(resp []byte) (hash uint64){
	sh := simhash.NewSimhash()
	data := TrimHtml(string(resp))
	hash = sh.GetSimhash(sh.NewWordFeatureSet([]byte(data)))
	return
}

func IsSimilarHash(hash1 uint64, hash2 uint64) bool{
	rs := simhash.Compare(hash1, hash2)
	// 差异大于8，认定为不同页面
	if rs > 8 {
		return false
	}
	return true
}

// IsEqualHash 判断页面hash是否相同
func IsEqualHash(hash1 uint64, hash2 uint64) bool{
	rs := simhash.Compare(hash1, hash2)
	if rs == 0 {
		return true
	}
	return false

}
