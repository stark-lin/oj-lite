// Builds browser-ready HTTP URLs for the addresses served by the application.

package app

import (
	"fmt"
	"net"
	"sort"
)

func discoverHTTPAccessURLs(configuredHost string, listenerAddr net.Addr) ([]string, error) {
	listenerHost, port, err := net.SplitHostPort(listenerAddr.String())
	if err != nil {
		return nil, fmt.Errorf("split listener address %q: %w", listenerAddr.String(), err)
	}

	host := configuredHost
	if host == "" {
		host = listenerHost
	}
	if !isUnspecifiedHost(host) {
		return []string{httpURL(host, port)}, nil
	}

	interfaceAddrs, err := net.InterfaceAddrs()
	urls := buildHTTPAccessURLs(host, port, interfaceAddrs)
	if err != nil {
		return urls, fmt.Errorf("list network interface addresses: %w", err)
	}

	return urls, nil
}

func buildHTTPAccessURLs(host, port string, interfaceAddrs []net.Addr) []string {
	if !isUnspecifiedHost(host) {
		return []string{httpURL(host, port)}
	}

	includeIPv4 := true
	includeIPv6 := true
	localHost := "localhost"
	if wildcardIP := net.ParseIP(host); wildcardIP != nil {
		if wildcardIP.To4() != nil {
			includeIPv6 = false
		} else {
			includeIPv4 = false
			localHost = "::1"
		}
	}

	localURL := httpURL(localHost, port)
	seen := map[string]struct{}{localURL: {}}
	networkURLs := make([]string, 0, len(interfaceAddrs))
	for _, addr := range interfaceAddrs {
		ip := interfaceIP(addr)
		if ip == nil || ip.IsLoopback() || !ip.IsGlobalUnicast() {
			continue
		}
		if ip.To4() != nil && !includeIPv4 {
			continue
		}
		if ip.To4() == nil && !includeIPv6 {
			continue
		}

		url := httpURL(ip.String(), port)
		if _, ok := seen[url]; ok {
			continue
		}
		seen[url] = struct{}{}
		networkURLs = append(networkURLs, url)
	}

	sort.Strings(networkURLs)
	return append([]string{localURL}, networkURLs...)
}

func isUnspecifiedHost(host string) bool {
	if host == "" {
		return true
	}

	ip := net.ParseIP(host)
	return ip != nil && ip.IsUnspecified()
}

func interfaceIP(addr net.Addr) net.IP {
	switch value := addr.(type) {
	case *net.IPAddr:
		return value.IP
	case *net.IPNet:
		return value.IP
	default:
		return nil
	}
}

func httpURL(host, port string) string {
	return "http://" + net.JoinHostPort(host, port)
}
