package root

import (
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"net"
	"strconv"
	"strings"
)

func parseHosts(hosts string) (hostList []string) {
	if strings.Contains(hosts, ",") {
		hostList = strings.Split(hosts, ",")
	} else {
		hostList = strings.Split(hosts, " ")
	}
	return
}

func parseHostRange(hostRange string) (hostList []string, err error) {
	hostR := strings.TrimSpace(hostRange)
	ip := strings.Split(hostR, "-")
	firstIP := net.ParseIP(ip[0])
	lastIP := net.ParseIP(ip[1])

	if firstIP == nil || lastIP == nil {
		err = errors.New("请输入正确的IP地址")
		return
	}

	firstIPNum := iPToInt(firstIP)
	lastIPNum := iPToInt(lastIP)

	for i := firstIPNum; i <= lastIPNum; i++ {
		switch {
		case strings.HasSuffix(intToIP(i).String(), ".0"):
			continue
		case strings.HasSuffix(intToIP(i).String(), ".1"):
			continue
		case strings.HasSuffix(intToIP(i).String(), ".255"):
			continue
		default:
			hostList = append(hostList, intToIP(i).String())
		}
	}
	return
}

func parseHostNet(hostNet string) (hostList []string, err error) {
	fmt.Printf("host: %v, type: %T", hostNet, hostNet)
	_, ipNet, e := net.ParseCIDR(hostNet)
	if e != nil {
		err = errors.New(fmt.Sprintf("网段解析失败,err: %s", e))
		return
	}
	firstIP, lastIP := networkRange(ipNet)
	ipRange := fmt.Sprintf("%s-%s", firstIP.String(), lastIP.String())
	hostList, err = parseHostRange(ipRange)
	return
}

func parseFile(group string) (conns []Connection, err error) {
	vars := viper.GetStringMapString(fmt.Sprintf("all.%s.vars", group))
	c := &Connection{}
	stringMap := viper.Get(fmt.Sprintf("all.%s.hosts", group))
	s, ok := stringMap.([]interface{})
	if ok {
		for _, value := range s {
			switch v := value.(type) {
			case string:
				c.Host = v
				c.Port = vars["port"]
				c.Username = vars["username"]
				c.Password = vars["password"]
				c.Key = vars["key"]
				conns = append(conns, *c)
			case map[interface{}]interface{}:
				for key, v2 := range v {
					c.Host = key.(string)
					v3 := v2.(map[interface{}]interface{})
					for key, value := range v3 {
						switch key.(string) {
						case "username":
							c.Username = value.(string)
						case "password":
							c.Password = value.(string)
						case "port":
							p, ok := value.(string)
							if ok {
								c.Port = p
							} else {
								c.Port = strconv.Itoa(value.(int))
							}
						case "key":
							c.Key = value.(string)
						default:
							err = errors.New("invalid key")
						}
					}
					conns = append(conns, *c)
				}
			}
		}
	}
	return
}

func ParseAllGroups() (conns []Connection, err error) {
	var c []Connection
	allGroups := viper.GetStringMap("all")
	for key := range allGroups {
		if conns, err = parseFile(key); err != nil {
			return
		}
		c = append(c, conns...)
	}
	conns = removeDuplicateElement(c)
	return
}

func ParseGroups(groups string) (conns []Connection, err error) {
	var c []Connection
	if strings.Contains(groups, ",") {
		groupList := strings.Split(groups, ",")
		for _, value := range groupList {
			if conns, err = parseFile(value); err != nil {
				return
			}
			c = append(c, conns...)
		}
		conns = removeDuplicateElement(c)
	} else {
		if conns, err = parseFile(groups); err != nil {
			return
		}
	}
	return
}

func removeDuplicateElement(conns []Connection) []Connection {
	result := make([]Connection, 0, len(conns))
	temp := map[string]struct{}{}
	for _, item := range conns {
		if _, ok := temp[item.Host]; !ok {
			temp[item.Host] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func iPToInt(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip.To4())
}

func intToIP(n uint32) net.IP {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, n)
	return net.IP(b)
}

func networkRange(network *net.IPNet) (net.IP, net.IP) {
	netIP := network.IP.To4()
	firstIP := netIP.Mask(network.Mask)
	lastIP := net.IPv4(0, 0, 0, 0).To4()
	for i := 0; i < len(lastIP); i++ {
		lastIP[i] = netIP[i] | ^network.Mask[i]
	}
	return firstIP, lastIP
}

func GenHostList(hosts, hostRange, hostNet string) (hostList []string, err error) {
	if hosts != "" {
		hostList = parseHosts(hosts)
	}
	if hostRange != "" {
		hostList, err = parseHostRange(hostRange)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	if hostNet != "" {
		hostList, err = parseHostNet(hostNet)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	return
}
