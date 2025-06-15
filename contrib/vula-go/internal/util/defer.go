package util

import (
	"io"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
)

func CloseLog(c io.Closer) {
	err := c.Close()
	if err != nil {
		core.LogWarnf("error closing resource: %v", err)
	}
}
