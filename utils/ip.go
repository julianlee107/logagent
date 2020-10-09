package utils

import (
	"net"
	"strings"
)

func GetOutboundIP(addr string) (ip string, err error) {
	conn, err := net.Dial("udp", addr)
	defer conn.Close()
	if err != nil {
		return
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip = strings.Split(localAddr.IP.String(), ":")[0]
	return
}
