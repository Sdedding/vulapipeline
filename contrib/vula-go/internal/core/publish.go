package core

type PublishListen interface {
	Listen(newAnnouncements map[string]string) error
}

type Publish interface {
	PublishListen
}
