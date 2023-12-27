package internal

import (
	"errors"
	"os"
)

var inEnclave = false

func init() {
	// Determine if we're running inside an enclave.
	if _, err := os.Stat("/dev/nsm"); err == nil {
		inEnclave = true
	} else if errors.Is(err, os.ErrNotExist) {
		inEnclave = false
	} else {
		// We encountered an unknown error.  Let's assume that we are not
		// inside an enclave.
		inEnclave = false
	}
	maybeSeedEntropy()
}
