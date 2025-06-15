package vulazeroconf

import (
	"context"
	"sync"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"codeberg.org/vula/vula/contrib/vula-go/internal/schema"
	"github.com/grandcat/zeroconf"
)

type Discover struct {
	mux               sync.Mutex
	processDescriptor core.OrganizeProcessDescriptor
	done              chan struct{}
}

var _ core.DiscoverListen = &Discover{}

func NewDiscover(p core.OrganizeProcessDescriptor) (d *Discover, err error) {

	d = &Discover{
		processDescriptor: p,
	}

	return
}

func (d *Discover) restart() <-chan struct{} {
	d.mux.Lock()
	defer d.mux.Unlock()
	if d.done != nil {
		close(d.done)
	}

	d.done = make(chan struct{})
	return d.done
}

func (d *Discover) Listen(addrs []string, ourWgPK string) error {
	done := d.restart()

	resolver, err := zeroconf.NewResolver()
	if err != nil {
		return err
	}

	// Browse will close entries when ctx is cancelled
	entries := make(chan *zeroconf.ServiceEntry, 16)
	ctx, cancel := context.WithCancel(context.Background())
	err = resolver.Browse(ctx, zeroconfServiceType, zeroconfDomain, entries)
	if err != nil {
		cancel()
		return err
	}

	core.LogInfo("Discover daemon starts discovering")
	go func() {
		defer cancel()
		for {
			select {
			case <-done:
				core.LogDebugf("stop discovering")
				return
			case entry := <-entries:
				if entry == nil {
					core.LogWarn("can not process zeroconf nil entry")
					continue
				}

				s, err := parseDescriptorZeroconf(entry.Text)
				if err != nil {
					core.LogInfof("dropping descriptor: %v", err)
					continue
				}
				descriptorString := schema.SerializeDescriptorString(s)
				core.LogDebugf("discovered descriptor: %s", descriptorString)
				_, err = d.processDescriptor.ProcessDescriptorString(descriptorString)
				if err != nil {
					core.LogWarn(err)
				}
				core.LogDebug("processed...")
			}
		}
	}()

	return nil
}

func (d *Discover) Close() error {
	d.done <- struct{}{}
	return nil
}
