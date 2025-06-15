package core

type Discover interface {
	DiscoverListen
}

type DiscoverListen interface {
	Listen(addrs []string, ourWgPK string) error
}
