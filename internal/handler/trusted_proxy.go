package handler

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
)

func ParseTrustedProxies(value string) ([]netip.Prefix, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parts := strings.Split(value, ",")
	prefixes := make([]netip.Prefix, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("TRUSTED_PROXIES chua CIDR rong")
		}

		prefix, err := netip.ParsePrefix(part)
		if err != nil {
			return nil, fmt.Errorf("CIDR proxy khong hop le %q: %w", part, err)
		}
		prefixes = append(prefixes, prefix.Masked())
	}

	return prefixes, nil
}

func NewAuthHandlerWithTrustedProxies(
	userService authUserService,
	trustedProxies []netip.Prefix,
) *AuthHandler {
	h := NewAuthHandler(userService)
	h.trustedProxies = append([]netip.Prefix(nil), trustedProxies...)
	return h
}

func parseRemoteIP(remoteAddr string) (netip.Addr, bool) {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		addr, parseErr := netip.ParseAddr(host)
		if parseErr == nil {
			return addr.Unmap(), true
		}
	}

	addr, err := netip.ParseAddr(strings.Trim(remoteAddr, "[]"))
	if err != nil {
		return netip.Addr{}, false
	}
	return addr.Unmap(), true
}

func isTrustedProxy(addr netip.Addr, prefixes []netip.Prefix) bool {
	for _, prefix := range prefixes {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

func getClientIPWithTrustedProxies(
	r *http.Request,
	trustedProxies []netip.Prefix,
) string {
	remoteIP, ok := parseRemoteIP(r.RemoteAddr)
	if !ok {
		return r.RemoteAddr
	}

	if len(trustedProxies) == 0 || !isTrustedProxy(remoteIP, trustedProxies) {
		return remoteIP.String()
	}

	header := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if header == "" {
		return remoteIP.String()
	}

	items := strings.Split(header, ",")
	addresses := make([]netip.Addr, 0, len(items))

	for _, item := range items {
		addr, err := netip.ParseAddr(strings.TrimSpace(item))
		if err != nil {
			// Header sai dinh dang: bo qua toan bo.
			return remoteIP.String()
		}
		addresses = append(addresses, addr.Unmap())
	}

	// Duyet tu proxy gan server nhat ve client.
	for i := len(addresses) - 1; i >= 0; i-- {
		if !isTrustedProxy(addresses[i], trustedProxies) {
			return addresses[i].String()
		}
	}

	return remoteIP.String()
}

// Mac dinh khong tin X-Forwarded-For.
func getClientIP(r *http.Request) string {
	return getClientIPWithTrustedProxies(r, nil)
}
