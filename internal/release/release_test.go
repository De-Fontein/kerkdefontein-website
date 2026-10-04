package release

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func writeIndex(text string) func(string) error {
	return func(dir string) error { return os.WriteFile(filepath.Join(dir, "index.html"), []byte(text), 0o644) }
}

func readCurrent(t *testing.T, root string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "current", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestPublish_SwitchesCurrentToNewRelease(t *testing.T) {
	root := t.TempDir()
	if _, err := Publish(root, "001", 5, writeIndex("one")); err != nil {
		t.Fatal(err)
	}
	if _, err := Publish(root, "002", 5, writeIndex("two")); err != nil {
		t.Fatal(err)
	}
	if got := readCurrent(t, root); got != "two" {
		t.Errorf("current = %q, want two", got)
	}
	target, err := os.Readlink(filepath.Join(root, "current"))
	if err != nil || target != filepath.Join("releases", "002") {
		t.Errorf("symlink target = %q, %v; want relative releases/002", target, err)
	}
}

func TestPublish_FailedBuildKeepsCurrentAndRemovesPartialDir(t *testing.T) {
	root := t.TempDir()
	if _, err := Publish(root, "001", 5, writeIndex("good")); err != nil {
		t.Fatal(err)
	}
	_, err := Publish(root, "002", 5, func(dir string) error {
		os.WriteFile(filepath.Join(dir, "index.html"), []byte("half"), 0o644)
		return errors.New("render failed")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if got := readCurrent(t, root); got != "good" {
		t.Errorf("current = %q, want good", got)
	}
	if _, err := os.Stat(filepath.Join(root, "releases", "002")); !errors.Is(err, os.ErrNotExist) {
		t.Error("partial release dir was not removed")
	}
}

func TestPublish_KeepsOnlyNewestReleases(t *testing.T) {
	root := t.TempDir()
	for i := 1; i <= 7; i++ {
		if _, err := Publish(root, fmt.Sprintf("%03d", i), 5, writeIndex(fmt.Sprint(i))); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(filepath.Join(root, "releases"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 5 || entries[0].Name() != "003" {
		t.Errorf("releases = %v, want 003..007", entries)
	}
}

// Two runs in the same second (timer + make deploy) must not write into, or delete, the live release.
func TestPublish_ExistingNameFailsWithoutTouchingCurrent(t *testing.T) {
	root := t.TempDir()
	if _, err := Publish(root, "001", 5, writeIndex("live")); err != nil {
		t.Fatal(err)
	}
	_, err := Publish(root, "001", 5, func(string) error { return errors.New("render failed") })
	if err == nil {
		t.Fatal("expected an error for a release name that already exists")
	}
	if got := readCurrent(t, root); got != "live" {
		t.Errorf("current = %q, want live", got)
	}
}

// If the clock jumps back, the newest release sorts first; pruning must still keep what current points to.
func TestPublish_PruneNeverRemovesCurrent(t *testing.T) {
	root := t.TempDir()
	if _, err := Publish(root, "005", 1, writeIndex("new")); err != nil {
		t.Fatal(err)
	}
	if _, err := Publish(root, "001", 1, writeIndex("after clock jump")); err != nil {
		t.Fatal(err)
	}
	if got := readCurrent(t, root); got != "after clock jump" {
		t.Errorf("current = %q", got)
	}
}
