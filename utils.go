package redis

import (
	"strings"

	"github.com/miekg/dns"
)

func Key(dn, prefix string) string {
	labels := dns.SplitDomainName(dn)
	for i, j := 0, len(labels)-1; i < j; i, j = i+1, j-1 {
		labels[i], labels[j] = labels[j], labels[i]
	}
	if prefix != "" {
		labels = append([]string{prefix}, labels...)
	}
	return strings.Join(labels, ":")
}

func AnyKey(key string) string {
	parts := strings.Split(key, ":")
	parts[len(parts)-1] = "*"
	return strings.Join(parts, ":")
}

// Split255 splits a string into 255 byte chunks.
func Split255(s string) []string {
	if len(s) < 255 {
		return []string{s}
	}
	sx := []string{}
	p, i := 0, 255
	for {
		if i <= len(s) {
			sx = append(sx, s[p:i])
		} else {
			sx = append(sx, s[p:])
			break
		}
		p, i = p+255, i+255
	}

	return sx
}
