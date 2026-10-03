// Package images turns uploaded flyers into WebP via the distro's vips and poppler tools, so the Go
// binary carries no image codecs and the codecs get security updates from apt.
package images

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	ThumbWidth = 480
	LargeWidth = 1600
)

type Variant struct {
	File   string
	Width  int
	Height int
}

type Pair struct {
	Thumb Variant
	Large Variant
}

func Convert(ctx context.Context, src, mimeType, outDir, base string) (Pair, error) {
	input := src
	if mimeType == "application/pdf" {
		pagePrefix := filepath.Join(outDir, base+"-page")
		if err := run(ctx, "pdftoppm", "-f", "1", "-l", "1", "-r", "150", "-png", "-singlefile", src, pagePrefix); err != nil {
			return Pair{}, fmt.Errorf("render first pdf page: %w", err)
		}
		input = pagePrefix + ".png"
	}
	thumb, err := resize(ctx, input, outDir, base, ThumbWidth)
	if err != nil {
		return Pair{}, err
	}
	large, err := resize(ctx, input, outDir, base, LargeWidth)
	if err != nil {
		return Pair{}, err
	}
	return Pair{Thumb: thumb, Large: large}, nil
}

// resize fits the image to width; the huge --height leaves height unconstrained and --size down prevents upscaling.
func resize(ctx context.Context, input, outDir, base string, width int) (Variant, error) {
	name := fmt.Sprintf("%s-%d.webp", base, width)
	out := filepath.Join(outDir, name)
	if err := run(ctx, "vips", "thumbnail", input, out+"[Q=80,strip]", strconv.Itoa(width), "--height", "100000", "--size", "down"); err != nil {
		return Variant{}, fmt.Errorf("resize to %d: %w", width, err)
	}
	w, err := header(ctx, out, "width")
	if err != nil {
		return Variant{}, err
	}
	h, err := header(ctx, out, "height")
	if err != nil {
		return Variant{}, err
	}
	return Variant{File: name, Width: w, Height: h}, nil
}

func header(ctx context.Context, file, field string) (int, error) {
	out, err := exec.CommandContext(ctx, "vipsheader", "-f", field, file).Output()
	if err != nil {
		return 0, fmt.Errorf("read %s of %s: %w", field, file, err)
	}
	return strconv.Atoi(strings.TrimSpace(string(out)))
}

func run(ctx context.Context, name string, args ...string) error {
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(stderr.String()))
	}
	return nil
}
