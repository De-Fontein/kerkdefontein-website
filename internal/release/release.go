// Package release builds each site version in its own directory and switches to it with one rename,
// so visitors never see a half-written site and a failed build changes nothing.
package release

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func Publish(root, name string, keep int, build func(dir string) error) (string, error) {
	releases := filepath.Join(root, "releases")
	dir := filepath.Join(releases, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create release dir: %w", err)
	}
	if err := build(dir); err != nil {
		os.RemoveAll(dir)
		return "", fmt.Errorf("build release %s: %w", name, err)
	}
	if err := switchCurrent(root, filepath.Join("releases", name)); err != nil {
		os.RemoveAll(dir)
		return "", err
	}
	return dir, prune(releases, keep)
}

// switchCurrent relies on rename(2) replacing the old symlink atomically.
func switchCurrent(root, target string) error {
	tmp := filepath.Join(root, "current.tmp")
	os.Remove(tmp)
	if err := os.Symlink(target, tmp); err != nil {
		return fmt.Errorf("create symlink: %w", err)
	}
	if err := os.Rename(tmp, filepath.Join(root, "current")); err != nil {
		return fmt.Errorf("switch current: %w", err)
	}
	return nil
}

func prune(releases string, keep int) error {
	entries, err := os.ReadDir(releases)
	if err != nil {
		return fmt.Errorf("list releases: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	for len(names) > keep {
		if err := os.RemoveAll(filepath.Join(releases, names[0])); err != nil {
			return fmt.Errorf("remove old release: %w", err)
		}
		names = names[1:]
	}
	return nil
}
