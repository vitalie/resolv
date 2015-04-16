package resolv

import (
	"math/rand"
	"net"
)

// peekRandom peeks a random name server from nss list
// returning the selected server and the remaining servers.
func peekRandom(nss []string) (string, []string) {
	n := len(nss) - 1

	// rand.Intn panics if n <= 0.
	if n == 0 {
		return nss[0], nil
	}

	i := rand.Intn(n)
	return nss[i], append(nss[:i], nss[i+1:]...)
}

func toIPv4(s string) (net.IP, bool) {
	ip := net.ParseIP(s)
	return ip, ip != nil && ip.To4() != nil
}

func toIPv6(s string) (net.IP, bool) {
	ip := net.ParseIP(s)
	return ip, ip != nil && ip.To16() != nil
}
