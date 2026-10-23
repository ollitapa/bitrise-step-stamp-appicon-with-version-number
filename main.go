package main

import (
	"embed"
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fogleman/gg"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/opentype"
)

func measureString(face font.Face, text string) float64 {
	d := &font.Drawer{Face: face}
	return float64(d.MeasureString(text))
}

//go:embed assets
var assetsFS embed.FS

const fallbackFontPath = "assets/Go-Bold.ttf"

func loadFallbackFont() ([]byte, error) {
	data, err := assetsFS.ReadFile(fallbackFontPath)
	if err != nil {
		data = gobold.TTF
	}
	return data, nil
}

func loadFace(fontInput string, size float64) (font.Face, error) {
	var fontData []byte
	if fontInput != "" {
		data, err := os.ReadFile(fontInput)
		if err != nil {
			return nil, fmt.Errorf("could not read font file %s: %w", fontInput, err)
		}
		fontData = data
	} else {
		data, err := loadFallbackFont()
		if err != nil {
			return nil, fmt.Errorf("could not load fallback font: %w", err)
		}
		fontData = data
	}

	parsed, err := opentype.Parse(fontData)
	if err != nil {
		return nil, fmt.Errorf("could not parse font: %w", err)
	}
	return opentype.NewFace(parsed, &opentype.FaceOptions{
		Size: size,
		DPI:  72,
	})
}

func parseHexColor(value string) (int, int, int, int, bool) {
	if !strings.HasPrefix(value, "#") {
		return 0, 0, 0, 0, false
	}
	hex := strings.TrimPrefix(value, "#")
	switch len(hex) {
	case 3:
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2], 'f', 'f'})
	case 4:
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2], hex[3], hex[3]})
	case 6:
		hex += "ff"
	case 8:
	default:
		return 0, 0, 0, 0, false
	}
	v, err := strconv.ParseUint(hex, 16, 64)
	if err != nil {
		return 0, 0, 0, 0, false
	}
	return int(v >> 24 & 0xff), int(v >> 16 & 0xff), int(v >> 8 & 0xff), int(v & 0xff), true
}

func parseColor(value string) (int, int, int, int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, 0, 0, 128, nil
	}
	if r, g, b, a, ok := parseHexColor(strings.ToLower(value)); ok {
		return r, g, b, a, nil
	}
	named := map[string][4]int{
		"white": {255, 255, 255, 255}, "black": {0, 0, 0, 255},
		"red": {255, 0, 0, 255}, "green": {0, 128, 0, 255},
		"blue": {0, 0, 255, 255}, "yellow": {255, 255, 0, 255},
		"cyan": {0, 255, 255, 255}, "magenta": {255, 0, 255, 255},
		"gray": {128, 128, 128, 255}, "grey": {128, 128, 128, 255},
		"orange": {255, 165, 0, 255}, "transparent": {0, 0, 0, 0},
	}
	if rgba, ok := named[value]; ok {
		return rgba[0], rgba[1], rgba[2], rgba[3], nil
	}
	return 0, 0, 0, 0, fmt.Errorf("could not parse color %q: use #RGB, #RRGGBB, #RRGGBBAA or a common color name", value)
}

func fitFace(fontInput, text string, maxWidth, maxHeight int) (font.Face, error) {
	low, high := 4.0, float64(maxHeight)
	bestSize := low
	for high-low > 0.5 {
		size := math.Floor((low + high) / 2)
		face, err := loadFace(fontInput, size)
		if err != nil {
			return nil, err
		}
		w := measureString(face, text)
		face.Close()
		if w <= float64(maxWidth) && size <= float64(maxHeight)*0.8 {
			bestSize = size
			low = size
		} else {
			high = size
		}
	}
	return loadFace(fontInput, bestSize)
}

func stampIcon(path, version, buildNumber, fgColor, bgColor, fontInput string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open icon %s: %w", path, err)
	}
	defer f.Close()

	imgData, _, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("could not decode icon %s: %w", path, err)
	}
	bounds := imgData.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	dc := gg.NewContext(width, height)
	dc.DrawImage(imgData, 0, 0)

	bannerH := int(math.Floor(float64(height) * 0.3))
	if bannerH < 1 {
		bannerH = 1
	}

	br, bg, bb, ba, err := parseColor(bgColor)
	if err != nil {
		return err
	}
	dc.SetRGBA255(br, bg, bb, ba)
	dc.DrawRectangle(0, float64(height-bannerH), float64(width), float64(bannerH))
	dc.Fill()

	text := "- " + version + "(" + buildNumber + ")"
	face, err := fitFace(fontInput, text, int(float64(width)*0.9), bannerH)
	if err != nil {
		return err
	}
	defer face.Close()

	fr, fgC, fb, fa, err := parseColor(fgColor)
	if err != nil {
		return err
	}
	dc.SetRGBA255(fr, fgC, fb, fa)
	dc.SetFontFace(face)

	tw, th := dc.MeasureString(text)
	tx := (float64(width) - tw) / 2
	ty := float64(height) - float64(bannerH) + (float64(bannerH)-th)/2
	dc.DrawString(text, tx, ty)

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not write icon %s: %w", path, err)
	}
	defer out.Close()
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	if err := encoder.Encode(out, dc.Image()); err != nil {
		return fmt.Errorf("could not encode icon %s: %w", path, err)
	}
	return nil
}

func main() {
	iconPath := os.Getenv("stamp_path_to_icons")
	version := os.Getenv("stamp_version")
	buildNumber := os.Getenv("stamp_build_number")
	fgColor := os.Getenv("foreground_color")
	bgColor := os.Getenv("background_color")
	fontInput := os.Getenv("stamp_font")

	if version == "" {
		log.Println("Version number (stamp_version) is required.")
		os.Exit(1)
	}
	if buildNumber == "" {
		log.Println("Build number (stamp_build_number) is required.")
		os.Exit(1)
	}

	log.Println("Version number to stamp:", version)
	log.Println("Build number to stamp:", buildNumber)
	log.Println("Foreground color is:", fgColor)
	log.Println("Background color is:", bgColor)
	if fontInput != "" {
		log.Println("Using font:", fontInput)
	} else {
		log.Println("No font specified; using the step's bundled font.")
	}

	log.Println("Finding icons from directory:", iconPath)

	files, err := os.ReadDir(iconPath)
	if err != nil {
		log.Printf("Could not read directory: %v", err)
		os.Exit(1)
	}

	filePaths := make([]string, 0)
	for _, f := range files {
		if strings.EqualFold(filepath.Ext(f.Name()), ".png") {
			filePaths = append(filePaths, filepath.Join(iconPath, f.Name()))
		}
	}

	if len(filePaths) == 0 {
		log.Println("No .png icons found in directory:", iconPath)
		os.Exit(1)
	}

	log.Println(filePaths)

	for _, f := range filePaths {
		if err := stampIcon(f, version, buildNumber, fgColor, bgColor, fontInput); err != nil {
			log.Printf("Failed to stamp %s: %v", f, err)
			os.Exit(1)
		}
	}

	log.Printf("Stamped %d icons.", len(filePaths))
	os.Exit(0)
}
