package utils

import (
	"crypto/sha512"
	"fmt"
)

const (
	maxChainLength = 28
	chainPrefix    = "CNI-"
	prefixLength   = len(chainPrefix)
)

// Generates a chain name to be used with iptables.
// Ensures that the generated chain name is less than
// maxChainLength chars in length
func FormatChainName(name string, id string) string {
	chainBytes := sha512.Sum512([]byte(name + id))
	chain := fmt.Sprintf("%s%x", chainPrefix, chainBytes)
	return chain[:maxChainLength]
}
