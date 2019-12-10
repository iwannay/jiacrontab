package jiacrontabd

import (
	"jiacrontab/pkg/util"
	"os"

	"github.com/iwannay/log"

	"container/list"
	"net"
	"strconv"
	"strings"
)

func writeFile(fPath string, content *[]byte) {
	f, err := util.TryOpen(fPath, os.O_APPEND|os.O_CREATE|os.O_RDWR)
	if err != nil {
		log.Errorf("writeLog: %v", err)
		return
	}
	defer f.Close()
	f.Write(*content)
}

func GetIntranetIpList() *list.List {
	ipList := list.New()
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return ipList
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipList.PushBack(ipnet.IP.String())
			}

		}
	}
	return ipList
}

func isIpBelong(ip, cidr string) bool {
	ipAddr := strings.Split(ip, `.`)
	if len(ipAddr) < 4 {
		return false
	}
	if ip == cidr {
		return true
	}
	cidrArr := strings.Split(cidr, `/`)
	if len(cidrArr) < 2 {
		return false
	}
	var tmp = make([]string, 0)
	for key, value := range strings.Split(`255.255.255.0`, `.`) {
		iint, _ := strconv.Atoi(value)

		iint2, _ := strconv.Atoi(ipAddr[key])

		tmp = append(tmp, strconv.Itoa(iint&iint2))
	}
	return strings.Join(tmp, `.`) == cidrArr[0]
}

func checkIpInWhiteList(whiteIpStr string) bool {
	myIps := GetIntranetIpList()
	whiteIpList := strings.Split(whiteIpStr, `,`)
	if len(whiteIpList) == 0 {
		return true
	}
	for item := myIps.Front(); nil != item; item = item.Next() {
		for i := range whiteIpList {
			isBelong := isIpBelong(item.Value.(string), whiteIpList[i])
			if isBelong {
				return true
			}
		}
	}
	return false
}
