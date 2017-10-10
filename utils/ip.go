package utils

import (
	"net"

	"github.com/pkg/errors"
)

var (
	ErrNoIpFound = errors.New("no ip found")
)

func GetLocalNetworkIp() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", errors.Wrap(err, "failed to list network interfaces")
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", ErrNoIpFound
}
