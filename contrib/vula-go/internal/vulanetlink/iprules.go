package vulanetlink

import (
	"fmt"
	"math"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"github.com/vishvananda/netlink"
)

func syncNetlinkIPRules(routingTable int, firewallMark, ipRulePriority int, dryRun bool) (log []string, err error) {
	if firewallMark < 0 || firewallMark > math.MaxUint32 {
		err = fmt.Errorf("firewallMark too big")
		return
	}

	firewallMarkU32 := uint32(firewallMark)

	for family, familyName := range netlinkAddressFamilies {

		// find existing rule
		rules, err := netlink.RuleList(family)
		if err != nil {
			return log, err
		}
		filteredRules := []netlink.Rule{}
		for _, rule := range rules {
			isMatch := rule.Table == routingTable &&
				rule.Mark == firewallMarkU32 &&
				rule.Priority == ipRulePriority &&
				rule.Invert

			if isMatch {
				filteredRules = append(filteredRules, rule)
			}
		}

		if len(filteredRules) > 0 {
			core.LogDebugf("Found expected existing rule: %v", filteredRules)
			continue
		}

		// existing rule not found. Create new rule
		if !dryRun {
			rule := netlink.NewRule()
			rule.Table = routingTable
			rule.Priority = ipRulePriority
			rule.Mark = firewallMarkU32
			rule.Family = family
			rule.Invert = true

			err = netlink.RuleAdd(rule)
			if err != nil {
				return log, err
			}
		}

		log = append(log, fmt.Sprintf("ip -%s rule add not from all fwmark 0x%x lookup %d", familyName, firewallMark, routingTable))
	}
	return
}

var netlinkAddressFamilies = map[int]string{
	netlink.FAMILY_V4: "4",
	netlink.FAMILY_V6: "6",
}
