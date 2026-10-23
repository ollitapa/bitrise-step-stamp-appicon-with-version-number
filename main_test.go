package main

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestParseColor(t *testing.T) {
	tests := []struct {
		value      string
		r, g, b, a int

		wantErr bool
	}{
		{"white", 255, 255, 255, 255, false},
		{"#333", 51, 51, 51, 255, false},
		{"#0008", 0, 0, 0, 136, false},
		{"#333333", 51, 51, 51, 255, false},
		{"#00000080", 0, 0, 0, 128, false},
		{"", 0, 0, 0, 128, false},
		{"notacolor", 0, 0, 0, 0, true},
	}
	for _, tt := range tests {
		r, g, b, a, err := parseColor(tt.value)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseColor(%q): expected error", tt.value)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseColor(%q): unexpected error %v", tt.value, err)
			continue
		}
		if r != tt.r || g != tt.g || b != tt.b || a != tt.a {
			t.Errorf("parseColor(%q) = %d,%d,%d,%d; want %d,%d,%d,%d", tt.value, r, g, b, a, tt.r, tt.g, tt.b, tt.a)
		}
	}
}

func TestStampIcon(t *testing.T) {
	dir := t.TempDir()
	src, err := os.Open(filepath.Join("..", "testing", "AppIcon", "go.png"))
	if err != nil {
		t.Skipf("test fixture not available: %v", err)
	}
	imgData, _, err := image.Decode(src)
	src.Close()
	if err != nil {
		t.Fatalf("could not decode fixture: %v", err)
	}
	origBounds := imgData.Bounds()

	iconPath := filepath.Join(dir, "icon.png")
	out, err := os.Create(iconPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(out, imgData); err != nil {
		t.Fatal(err)
	}
	out.Close()

	if err := stampIcon(iconPath, "1.2", "123", "white", "#0008", ""); err != nil {
		t.Fatalf("stampIcon failed: %v", err)
	}

	f, err := os.Open(iconPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	stamped, _, err := image.Decode(f)
	if err != nil {
		t.Fatalf("could not decode stamped icon: %v", err)
	}
	b := stamped.Bounds()
	if b.Dx() != origBounds.Dx() || b.Dy() != origBounds.Dy() {
		t.Fatalf("stamped icon dimensions changed: %v; want %v", b, origBounds)
	}

	// The bottom banner must differ from the original pixels: check the
	// bottom row is darker than the original image's bottom row average.
	origAvg := averageColor(imgData, origBounds.Dy()-1)
	stampedAvg := averageColor(stamped, b.Dy()-1)
	if origAvg == stampedAvg {
		t.Errorf("bottom row unchanged after stamping; banner was not drawn")
	}
}

func averageColor(img image.Image, y int) uint32 {
	b := img.Bounds()
	var r, g, bl uint32
	n := uint32(b.Dx())
	for x := b.Min.X; x < b.Max.X; x++ {
		cr, cg, cb, _ := img.At(x, y).RGBA()
		r += cr
		g += cg
		bl += cb
	}
	return (r/n)<<16 | (g/n)<<8 | bl/n
}

func TestFitFaceFillsBanner(t *testing.T) {
	text := "- 1.2(123)"
	face, err := fitFace("", text, 525, 99)
	if err != nil {
		t.Fatalf("fitFace failed: %v", err)
	}
	defer face.Close()
	m := face.Metrics()
	if m.Height.Round() < 20 {
		t.Errorf("fitted font height = %d px; want a font that fills a meaningful part of the 99px banner (>=20px)", m.Height.Round())
	}
	if w := measureString(face, text); w > 525 {
		t.Errorf("fitted text width %.1f exceeds max width 525", w)
	}
}
