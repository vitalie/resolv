package resolv

import (
	"math/rand"
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
