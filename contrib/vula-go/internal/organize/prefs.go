package organize

import (
	"fmt"
	"regexp"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
)

var (
	localDomainRegex = regexp.MustCompile(`^[a-zA-Z0-9.-]+\.$`)
	ifacePrefixRegex = regexp.MustCompile("^[a-zA-Z0-9_]+$")
)

func validatePrefs(p *core.Prefs) (errors []string) {
	// check allow / forbidden conflicts
	for _, allowed := range p.SubnetsAllowed {
		for _, forbidden := range p.SubnetsForbidden {
			if allowed.Overlaps(forbidden) {
				errors = append(errors, fmt.Sprintf("allow / forbidden conflict: %s / %s", allowed, forbidden))
			}
		}
	}

	/*
		for _, net := range p.SubnetsAllowed {
			for _, n := range p.SubnetsAllowed {
				if net.Overlaps(n) {
					errors = append(errors, fmt.Sprintf("conflict in subnets_allowed: %s <--> %s", net, n))
				}
			}
		}
		for _, net := range p.SubnetsForbidden {
			for _, n := range p.SubnetsForbidden {
				if net.Overlaps(n) {
					errors = append(errors, fmt.Sprintf("conflict in subnets_forbidden: %s <--> %s", net, n))
				}
			}
		}
	*/
	for _, domain := range p.LocalDomains {
		if !localDomainRegex.MatchString(domain) {
			errors = append(errors, fmt.Sprintf("invalid local domain: %s", domain))
		}
	}

	for _, prefix := range p.IfacePrefixAllowed {
		if !ifacePrefixRegex.MatchString(prefix) {
			errors = append(errors, fmt.Sprintf("invalid interface prefix: %s", prefix))
		}
	}

	return
}
