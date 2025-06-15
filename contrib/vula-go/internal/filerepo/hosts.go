package filerepo

import (
	"fmt"
	"path/filepath"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"codeberg.org/vula/vula/contrib/vula-go/internal/util"
)

type HostsFile struct {
	FileName string
}

var _ core.HostsFileRepository = &HostsFile{}

func (f *HostsFile) WriteHostsFile(entries [][2]string) error {
	var content []byte

	for i := range entries {
		content = append(content, entries[i][0]...)
		content = append(content, ' ')
		content = append(content, entries[i][1]...)
		content = append(content, '\n')
	}

	dir := filepath.Dir(f.FileName)

	// TODO: Here and in organize_repository.go.. The parent dir is actually always the same. So we should probably do this differently
	if err := util.EnsureDir(dir, 0755); err != nil {
		return fmt.Errorf("creating hosts parent directory: %w", err)
	}

	if err := util.EnsureFile(f.FileName, 0644); err != nil {
		return fmt.Errorf("creating/updating hosts file: %w", err)
	}

	if err := util.WriteFileAtomically(f.FileName, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing hosts file atomically: %w", err)
	}

	if err := util.ChownLikeDirIfRoot(dir); err != nil {
		return fmt.Errorf("setting hosts directory ownership: %w", err)
	}

	return nil
}
