package util

import (
	"net"
	"strings"

	"github.com/pkg/errors"
)

// ValidateAddress validates the given address and will return an error if the
// format is not <IP>:<PORT>
func ValidateAddress(address string) error {
	// TODO: this list is not extensive and needs to be changed once we allow DNS
	// names for external metrics endpoints
	const invalidChars = `abcdefghijklmnopqrstuvwxyz/\ `

	address = strings.ToLower(address)
	if strings.ContainsAny(address, invalidChars) {
		return errors.New("invalid character detected (required format: <IP>:<PORT>)")
	}

	// 	check if port if specified
	if !strings.Contains(address, ":") {
		return errors.New("no port specified")
	}

	h, p, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}

	if h == "" {
		return errors.New("no IP listen address specified")
	}

	if p == "" {
		return errors.New("no port specified")
	}

	return nil
}
