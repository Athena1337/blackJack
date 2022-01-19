package utils

import (
	"fmt"
	"github.com/t43Wiu6/tlog"
	"net"
	"sort"
	"strings"
)

// CheckCIDR 计算给出的一组ip或域名列表的CIDR重合值
func CheckCIDR(D []string) string{
	type tmp struct {
		CIDR  string
		Count int
	}

	var (
		lst []tmp
		rs  string
	)
	/* 创建集合 */
	cidrMap := make(map[string]int)
	for _, name := range D {
		addr, err := net.ResolveIPAddr("ip", name)
		if err != nil {
			log.Debugf("Resolvtion error, %v" , err.Error())
			continue
		}

		n := strings.Split(addr.String(), ".")
		n = n[:len(n)-1]
		p := strings.Join(n, ".") + ".0/24"

		/* 查看元素在集合中是否存在 */
		_, ok := cidrMap[p]
		/* 如果 ok 是 true, 则存在，否则不存在 */
		if ok {
			cidrMap[p] = cidrMap[p] + 1
			//fmt.Println("the cidr", p, " count is :", cidr)
		} else {
			cidrMap[p] = 1
		}
	}

	for k, v := range cidrMap {
		lst = append(lst, tmp{k, v})
	}

	sort.Slice(lst, func(i, j int) bool {
		return lst[i].Count > lst[j].Count // 降序
	})

	for _, vv := range lst {
		rs += fmt.Sprintf("%s : %d", vv.CIDR, vv.Count)
	}
	return rs
}
