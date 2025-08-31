package ftm

import (
	"net"
	"strings"
)

// IpType validates IPv4/IPv6.
type IpType struct{ BaseType }

func NewIpType() *IpType {
	return &IpType{BaseType{name: "ip", group: "ips", label: "IP Address", matchable: true, pivot: true, maxLength: 64}}
}
func (t *IpType) Validate(value string) bool { return net.ParseIP(value) != nil }
func (t *IpType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	s, ok := sanitizeText(text)
	if !ok {
		return "", false
	}
	ip := net.ParseIP(strings.TrimSpace(s))
	if ip == nil {
		return "", false
	}
	return ip.String(), true
}
