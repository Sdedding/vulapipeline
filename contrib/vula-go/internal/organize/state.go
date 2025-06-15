package organize

import (
	"os"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
)

func loadState(r core.OrganizeStateRepository) (s *core.OrganizeState, err error) {
	s, err = r.LoadOrganizeState()
	if err != nil {
		if err == os.ErrNotExist {
			core.LogDebug("Created new OrganizeState")
			return defaultOrganizeState, nil
		}
		return
	}
	return
}
