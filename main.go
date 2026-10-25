package main

import (
	"bytes"
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
	return float64(d.MeasureString(text)) / 64.0
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
	maxSize := float64(maxHeight) * 0.8
	if maxSize < 4 {
		maxSize = 4
	}
	best := -1.0
	for size := 4.0; size <= maxSize; size++ {
		face, err := loadFace(fontInput, size)
		if err != nil {
			return nil, err
		}
		fits := measureString(face, text) <= float64(maxWidth)
		face.Close()
		if fits {
			best = size
		} else {
			break
		}
	}
	if best < 0 {
		best = 4
	}
	return loadFace(fontInput, best)
}

func renderTemplate(tpl, version, buildNumber string) string {
	replacer := strings.NewReplacer(
		"%version%", version,
		"%build%", buildNumber,
		"\n", " ",
	)
	return replacer.Replace(tpl)
}

func stampIcon(opts stampOptions, path string) error {
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

	bannerRatio := opts.bannerHeightRatio
	if bannerRatio <= 0 || bannerRatio > 1 {
		bannerRatio = 0.3
	}
	bannerH := int(math.Floor(float64(height) * bannerRatio))
	if bannerH < 1 {
		bannerH = 1
	}

	br, bg, bb, ba, err := parseColor(opts.bgColor)
	if err != nil {
		return err
	}
	dc.SetRGBA255(br, bg, bb, ba)
	var bannerY float64
	switch opts.position {
	case "north":
		bannerY = 0
	case "center":
		bannerY = float64(height-bannerH) / 2
	default:
		bannerY = float64(height - bannerH)
	}
	dc.DrawRectangle(0, bannerY, float64(width), float64(bannerH))
	dc.Fill()

	text := opts.text
	if text == "" {
		text = "- " + opts.version + "(" + opts.buildNumber + ")"
	}
	face, err := fitFace(opts.font, text, int(float64(width)*0.9), bannerH)
	if err != nil {
		return err
	}
	defer face.Close()

	fr, fgC, fb, fa, err := parseColor(opts.fgColor)
	if err != nil {
		return err
	}
	dc.SetRGBA255(fr, fgC, fb, fa)
	dc.SetFontFace(face)

	// gg.DrawString takes the BASELINE y coordinate, so center the text
	// using the font's actual ascent/descent instead of treating the
	// measured height as the top of the text block.
	metrics := face.Metrics()
	ascent := float64(metrics.Ascent) / 64
	descent := float64(metrics.Descent) / 64
	glyphH := ascent + descent

	tw, _ := dc.MeasureString(text)
	tx := (float64(width) - tw) / 2
	// Center the baseline so that ascent/descent are balanced in the banner.
	ty := bannerY + (float64(bannerH)-glyphH)/2 + ascent
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

func isIconDir(path string) bool {
	base := filepath.Base(path)
	switch {
	case strings.HasSuffix(base, ".appiconset"):
		return true
	case strings.HasPrefix(base, "mipmap-") || strings.HasPrefix(base, "drawable-"):
		return true
	case strings.HasPrefix(base, "AppIcon") && strings.HasSuffix(base, ".iconset"):
		return true
	}
	return false
}

func findIconDirs(root string) []string {
	var dirs []string
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		if isIconDir(path) {
			dirs = append(dirs, path)
			return filepath.SkipDir
		}
		if strings.HasSuffix(path, ".git") || strings.HasSuffix(path, "node_modules") || strings.HasSuffix(path, "Pods") {
			return filepath.SkipDir
		}
		return nil
	})
	return dirs
}

func collectPNGs(paths []string) []string {
	var files []string
	for _, dir := range paths {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() && strings.EqualFold(filepath.Ext(e.Name()), ".png") {
				files = append(files, filepath.Join(dir, e.Name()))
			}
		}
	}
	return files
}

type stampOptions struct {
	version           string
	buildNumber       string
	text              string
	position          string
	bannerHeightRatio float64
	fgColor           string
	bgColor           string
	font              string
}

func writeEnvManifest(path string, envs map[string]string) error {
	var buf bytes.Buffer
	for k, v := range envs {
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(v)
		buf.WriteByte('\n')
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func main() {
	iconPath := os.Getenv("stamp_path_to_icons")
	version := os.Getenv("stamp_version")
	buildNumber := os.Getenv("stamp_build_number")
	textTemplate := os.Getenv("stamp_text")
	position := os.Getenv("stamp_position")
	bannerHeightRatio := 0.3
	if v := os.Getenv("stamp_banner_height_ratio"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			bannerHeightRatio = f
		} else {
			log.Printf("Could not parse stamp_banner_height_ratio %q, using 0.3.", v)
		}
	}
	fgColor := os.Getenv("foreground_color")
	bgColor := os.Getenv("background_color")
	fontInput := os.Getenv("stamp_font")

	opts := stampOptions{
		version:           version,
		buildNumber:       buildNumber,
		text:              renderTemplate(textTemplate, version, buildNumber),
		position:          position,
		bannerHeightRatio: bannerHeightRatio,
		fgColor:           fgColor,
		bgColor:           bgColor,
		font:              fontInput,
	}

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
	log.Println("Stamp text:", opts.text)
	log.Println("Stamp position:", position)
	log.Println("Foreground color is:", fgColor)
	log.Println("Background color is:", bgColor)
	if fontInput != "" {
		log.Println("Using font:", fontInput)
	} else {
		log.Println("No font specified; using the step's bundled font.")
	}

	var searchDirs []string
	if iconPath == "" {
		root := os.Getenv("BITRISE_SOURCE_DIR")
		if root == "" {
			if cwd, err := os.Getwd(); err == nil {
				root = cwd
			}
		}
		log.Println("No icon directory given; searching for icon directories under:", root)
		searchDirs = findIconDirs(root)
		if len(searchDirs) == 0 {
			log.Println("No icon directories found (looked for *.appiconset, mipmap-*, drawable-*, *.iconset) under:", root)
			log.Println("Set the stamp_path_to_icons input to the directory containing your icons.")
			os.Exit(1)
		}
		for _, d := range searchDirs {
			log.Println("Found icon directory:", d)
		}
	} else {
		log.Println("Finding icons from directory:", iconPath)
		searchDirs = []string{iconPath}
	}

	filePaths := collectPNGs(searchDirs)

	if len(filePaths) == 0 {
		log.Println("No .png icons found in:", searchDirs)
		os.Exit(1)
	}

	log.Println(filePaths)

	for _, f := range filePaths {
		if err := stampIcon(opts, f); err != nil {
			log.Printf("Failed to stamp %s: %v", f, err)
			os.Exit(1)
		}
	}

	log.Printf("Stamped %d icons.", len(filePaths))

	if envmanPath := os.Getenv("BITRISE_ENV_MANIFEST_PATH"); envmanPath != "" {
		manifest := map[string]string{
			"stamped_icon_count": strconv.Itoa(len(filePaths)),
			"stamped_icon_paths": strings.Join(filePaths, "\n"),
		}
		if err := writeEnvManifest(envmanPath, manifest); err != nil {
			log.Printf("Could not write Bitrise outputs: %v", err)
		}
		log.Println("Outputs: stamped_icon_count, stamped_icon_paths")
	}

	os.Exit(0)
}
