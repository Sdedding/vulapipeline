package vulazeroconf

import (
	"sync"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"codeberg.org/vula/vula/contrib/vula-go/internal/schema"
	"github.com/grandcat/zeroconf"
)

type Publish struct {
	mux       sync.Mutex
	zeroconfs map[string]*zeroconf.Server
}

var _ core.PublishListen = &Publish{}

func NewPublish() *Publish {
	return &Publish{
		zeroconfs: map[string]*zeroconf.Server{},
	}
}

func (p *Publish) Listen(m map[string]string) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	p.shutdownAll()

	for ipAddr, d := range m {
		err := p.publishDescriptor(ipAddr, d)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Publish) publishDescriptor(ipAddr string, descriptor string) error {
	server := p.zeroconfs[ipAddr]
	if server != nil {
		server.Shutdown()
	}

	s, err := schema.ParseDescriptorString(descriptor)
	if err != nil {
		return err
	}

	text := serializeDescriptorZeroconf(&s)
	server, err = zeroconf.Register(s.Hostname, zeroconfServiceType, zeroconfDomain, int(s.Port), text, nil)
	if err != nil {
		return err
	}

	p.zeroconfs[ipAddr] = server
	return nil
}

func (p *Publish) shutdownAll() {
	for _, server := range p.zeroconfs {
		server.Shutdown()
	}

	clear(p.zeroconfs)
}

func (p *Publish) Close() error {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.shutdownAll()
	return nil
}
