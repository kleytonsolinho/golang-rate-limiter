package utils

import (
	"net"
	"net/http"
	"strings"
)

func GetIP(r *http.Request) string {
	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		ip, _, _ = net.SplitHostPort(ip)
	}
	return ip
}
