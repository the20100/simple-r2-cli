package cmd

import "net/url"

func buildParams(pairs ...string) url.Values {
	p := url.Values{}
	for i := 0; i+1 < len(pairs); i += 2 {
		if pairs[i+1] != "" {
			p.Set(pairs[i], pairs[i+1])
		}
	}
	return p
}
