package helpers

import (
	"fmt"
	"net"
	"net/netip"
	"strings"
)

// IsURL checks if the given string starts with http:// or https://.
// It performs a simple prefix check to identify URLs.
func IsURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// ToNetIPAddr converts a remote address string to a netip.Addr.
// It handles addresses with or without ports and validates IP format.
func ToNetIPAddr(remoteAddress string) (*netip.Addr, error) {
	var host string

	// Try to split if remoteAddress contains port (e.g. "192.168.0.1:5000")
	if strings.Contains(remoteAddress, ":") {
		h, _, err := net.SplitHostPort(remoteAddress)
		if err != nil {
			// It might still be a plain IP like "::1" or malformed
			host = remoteAddress
		} else {
			host = h
		}
	} else {
		host = remoteAddress
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP: %s", remoteAddress)
	}

	ipAddr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return nil, fmt.Errorf("failed to convert %s to netip.Addr", remoteAddress)
	}

	return &ipAddr, nil
}
