package vulanetlink

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"os/exec"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"github.com/vishvananda/netlink"
)

func syncInterfaces(wgLinkName string, primaryIP netip.Addr, dryRun bool) (log []string, err error) {
	syncLog, err := syncWireguardLink(wgLinkName, dryRun)
	log = append(log, syncLog...)
	if err != nil {
		err = fmt.Errorf("sync wireguard link: %w", err)
		return
	}

	syncLog, err = syncDummyLink(core.DummyLinkName, dryRun)
	log = append(log, syncLog...)
	if err != nil {
		err = fmt.Errorf("sync dummy link: %w", err)
		return
	}

	syncLog, err = linkAddAddr(core.DummyLinkName, netip.PrefixFrom(primaryIP, 128), dryRun)
	log = append(log, syncLog...)
	if err != nil {
		err = fmt.Errorf("sync dummy link addr: %w", err)
	}
	return
}

func syncWireguardLink(linkName string, dryRun bool) (log []string, err error) {
	link, err := netlink.LinkByName(linkName)
	if err != nil {
		var errNotFound netlink.LinkNotFoundError
		if !errors.As(err, &errNotFound) {
			return
		}

		// link does not exist. Create it.
		if !dryRun {
			link := &netlink.Wireguard{
				LinkAttrs: netlink.NewLinkAttrs(),
			}
			link.Name = linkName
			err = netlink.LinkAdd(link)
			if err != nil {
				err = fmt.Errorf("failed to create netlink interface: %w", err)
				return
			}
		}

		log = append(log, "# create interface")
		log = append(log, fmt.Sprintf("ip link add %s type wireguard", linkName))

		link, err = netlink.LinkByName(linkName)
		if err != nil {
			return
		}
	}

	if link.Attrs().OperState != netlink.OperUp {
		if !dryRun {
			err = netlink.LinkSetUp(link)
			if err != nil {
				return
			}
		}
		log = append(log, "# bring up interface")
		log = append(log, fmt.Sprintf("ip link set up %s", linkName))
	}
	return
}

func linkAddAddr(linkName string, prefix netip.Prefix, dryRun bool) (log []string, err error) {
	allAddrs, err := getAllAddrs()
	if err != nil {
		return
	}

	// check if the address is assigned already
	for _, addr := range allAddrs {
		fmt.Println(addr.link, addr.prefix)
		if addr.link == linkName && addr.prefix == prefix {
			return
		}
	}

	log = append(log, fmt.Sprintf("ip addr add %s dev %s", prefix, linkName))
	if !dryRun {
		core.LogInfof("# %s", log[len(log)-1])

		var link netlink.Link
		link, err = netlink.LinkByName(linkName)
		if err != nil {
			return
		}
		err = netlink.AddrAdd(link, &netlink.Addr{
			IPNet: &net.IPNet{
				IP:   net.IP(prefix.Addr().AsSlice()),
				Mask: net.CIDRMask(prefix.Bits(), prefix.Addr().BitLen()),
			},
		})
	}
	return
}

func syncDummyLink(linkName string, dryRun bool) (log []string, err error) {
	link, err := netlink.LinkByName(linkName)
	if err == nil {
		if link.Type() == "dummy" {
			core.LogDebugf("dummy link %s exists", linkName)
			return
		}
		err = fmt.Errorf("link %s exists. But has wrong type %s. It should be dummy", linkName, link.Type())
		return
	} else {
		var errNotFound netlink.LinkNotFoundError
		if !errors.As(err, &errNotFound) {
			return
		}
	}

	if !dryRun {
		newLink := &netlink.Dummy{
			LinkAttrs: netlink.NewLinkAttrs(),
		}
		newLink.Name = linkName
		link = newLink
		err = netlink.LinkAdd(link)
		if err != nil {
			return
		}
		err = linkSetIp6AddrGenMode(linkName, "none")
		if err != nil {
			return
		}
		err = netlink.LinkSetUp(link)
	}
	log = append(log, fmt.Sprintf("ip link add name %s type dummy", linkName))
	log = append(log, fmt.Sprintf("ip link set dev %s addrgenmode none", linkName))
	return
}

// linkSetIp6AddrGenMode is a workaround for a missing function in netlink
// this can be removed once this https://github.com/vishvananda/netlink/pull/869/files
// pull request gets merged
func linkSetIp6AddrGenMode(linkName, mode string) error {
	cmd := exec.Command("ip", "link", "set", linkName, "addrgenmode", mode)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
