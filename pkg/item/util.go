package item

import "golang.org/x/net/publicsuffix"

func removeSubdomain(host string) string {
	domain, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		return host
	}
	return domain
}
