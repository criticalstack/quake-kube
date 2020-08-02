package net

import (
	"net"

	"github.com/pkg/errors"
)

// DetectHostIPv4 attempts to determine the host IPv4 address by finding the
// first non-loopback device with an assigned IPv4 address.
func DetectHostIPv4() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", errors.WithStack(err)
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() == nil {
				continue
			}
			return ipnet.IP.String(), nil
		}
	}
	return "", errors.New("cannot detect host IPv4 address")
}
