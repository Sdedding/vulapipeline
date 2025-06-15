package util

import (
	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// ChownLikeDirIfRoot changes the ownership of the file at path to match the directory it is in if the current user
// is root.
func ChownLikeDirIfRoot(path string) error {
	if os.Geteuid() != 0 {
		return nil
	}
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		if os.IsNotExist(err) {
			realPath = path
		} else {
			return fmt.Errorf("resolving real path: %w", err)
		}
	}

	dir := filepath.Dir(realPath)
	dirInfo, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("getting directory info: %w", err)
	}
	dirStat := dirInfo.Sys().(*syscall.Stat_t)
	uid := int(dirStat.Uid)
	gid := int(dirStat.Gid)

	return os.Chown(path, uid, gid)
}

// WriteFileAtomically writes a file atomically. It first writes the data to a temporary file and then renames it to
// the target file. Renaming is an atomic operation. This ensures that the target file is never partially written.
func WriteFileAtomically(path string, data []byte, perm os.FileMode) error {
	if !strings.HasPrefix(filepath.Clean(path), core.VulaOrganizeLibDir) {
		return fmt.Errorf("has invalid base path: %s", path)
	}
	file, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err := file.Close()
			if err != nil {
				return
			}
			err = os.Remove(file.Name())
			if err != nil {
				return
			}
		}
	}()
	current, err := os.Stat(path)
	if err == nil && current.Mode().IsRegular() {
		if err := os.Rename(path, path+".previous"); err != nil {
			return err
		}
	}
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err = file.Chmod(perm); err != nil {
		return err
	}
	if err = file.Sync(); err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

// EnsureDir ensures the parent directory of the given path exists.
func EnsureDir(path string, perm os.FileMode) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, perm)
}

/*
filePath = filepath.Join(basePath,filepath.Clean(filePath))
if !strings.HasPrefix(filePath, basePath) {
  return nil, fmt.Errorf("invalid path")
}
f, err := os.Open(filePath)
*/

// EnsureFile ensures the file exists (creates if missing) and updates its timestamps.
func EnsureFile(path string, perm os.FileMode) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if !strings.HasPrefix(filepath.Clean(path), core.VulaOrganizeLibDir) {
			return fmt.Errorf("has invalid base path: %s", path)
		}
		//#nosec G304 (please look at the check before)
		f, err := os.OpenFile(path, os.O_CREATE, perm)
		if err != nil {
			return err
		}
		return f.Close()
	} else if err == nil {
		now := time.Now()
		return os.Chtimes(path, now, now)
	} else {
		return err
	}
}
