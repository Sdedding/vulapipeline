package filerepo

import (
	"fmt"
	"os"
	"path/filepath"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"codeberg.org/vula/vula/contrib/vula-go/internal/schema"
	"codeberg.org/vula/vula/contrib/vula-go/internal/util"
	"gopkg.in/yaml.v3"
)

type OrganizeStateFile struct {
	FileName string
}

var _ core.OrganizeStateRepository = &OrganizeStateFile{}

func (o *OrganizeStateFile) LoadOrganizeState() (*core.OrganizeState, error) {
	core.LogDebugf("Loading s file from %s", o.FileName)

	data, err := os.ReadFile(o.FileName)
	if err != nil {
		core.LogInfof("Couldn't load s file: %v\n", err)
		return nil, os.ErrNotExist
	}

	s := schema.OrganizeState{}
	if err = yaml.Unmarshal(data, &s); err != nil {
		core.LogInfof("Couldn't load s file: %v", err)

		if _, statErr := os.Stat(o.FileName); statErr == nil {
			core.LogInfof(
				"Existing state file was malformed: %v",
				err,
			)
			return nil, fmt.Errorf("malformed state file")
		}

		return nil, os.ErrNotExist
	}

	if len(s.EventLog) > 0 {
		core.LogInfof("event_log contains %d entries", len(s.EventLog))
	}

	state, err := schema.ImportState(&s)
	return &state, err
}

func (o *OrganizeStateFile) SaveOrganizeState(state *core.OrganizeState) error {
	dir := filepath.Dir(o.FileName)

	if err := util.EnsureDir(dir, 0755); err != nil {
		return fmt.Errorf("creating hosts parent directory: %w", err)
	}

	if err := util.EnsureFile(o.FileName, 0600); err != nil {
		return fmt.Errorf("creating/updating hosts file: %w", err)
	}

	yamlEncoder := yaml.NewEncoder(nil)
	yamlEncoder.SetIndent(2)

	s := schema.ExportState(state)

	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshaling state to YAML: %w", err)
	}

	if err := util.WriteFileAtomically(o.FileName, data, 0600); err != nil {
		return fmt.Errorf("writing state file atomically: %w", err)
	}

	if err := util.ChownLikeDirIfRoot(o.FileName); err != nil {
		return fmt.Errorf("setting state file ownership: %w", err)
	}

	core.LogInfof("vula state file updated: %d peers", len(state.Peers))
	return nil
}
