package app

import (
	"net"
	"reflect"
	"testing"
)

func TestBuildHTTPAccessURLs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		host           string
		interfaceCIDRs []string
		want           []string
	}{
		{
			name: "IPv4 wildcard includes local and network addresses",
			host: "0.0.0.0",
			interfaceCIDRs: []string{
				"127.0.0.1/8",
				"192.168.1.25/24",
				"10.0.0.8/24",
				"169.254.10.20/16",
				"192.168.1.25/24",
				"2001:db8::25/64",
			},
			want: []string{
				"http://localhost:8080",
				"http://10.0.0.8:8080",
				"http://192.168.1.25:8080",
			},
		},
		{
			name: "IPv6 wildcard includes IPv6 addresses",
			host: "::",
			interfaceCIDRs: []string{
				"::1/128",
				"2001:db8::25/64",
				"192.168.1.25/24",
			},
			want: []string{
				"http://[::1]:8080",
				"http://[2001:db8::25]:8080",
			},
		},
		{
			name: "specific host is used directly",
			host: "127.0.0.1",
			interfaceCIDRs: []string{
				"192.168.1.25/24",
			},
			want: []string{"http://127.0.0.1:8080"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			interfaceAddrs := make([]net.Addr, 0, len(tt.interfaceCIDRs))
			for _, cidr := range tt.interfaceCIDRs {
				interfaceAddrs = append(interfaceAddrs, parseTestIPNet(t, cidr))
			}

			got := buildHTTPAccessURLs(tt.host, "8080", interfaceAddrs)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("access URLs = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func parseTestIPNet(t *testing.T, cidr string) *net.IPNet {
	t.Helper()

	ip, network, err := net.ParseCIDR(cidr)
	if err != nil {
		t.Fatalf("parse CIDR %q: %v", cidr, err)
	}
	network.IP = ip
	return network
}
