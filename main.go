package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func findMagickBinary() string {
	if _, err := exec.LookPath("magick"); err == nil {
		return "magick"
	}
	return "convert"
}

func findIdentifyBinary() (string, []string) {
	if _, err := exec.LookPath("identify"); err == nil {
		return "identify", nil
	}
	return "magick", []string{"identify"}
}

func findFont(fontInput string) string {
	if fontInput != "" {
		return fontInput
	}
	candidates := []string{}
	if runtime.GOOS == "darwin" {
		candidates = []string{
			"/System/Library/Fonts/Helvetica.ttc",
			"/System/Library/Fonts/Supplemental/Arial.ttf",
			"/System/Library/Fonts/Supplemental/Georgia.ttf",
			"/System/Library/Fonts/Supplemental/Times New Roman.ttf",
			"/Library/Fonts/Arial.ttf",
		}
	} else {
		candidates = []string{
			"/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",
			"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
			"/usr/share/fonts/truetype/liberation/LiberationSans-Bold.ttf",
			"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
			"/usr/share/fonts/truetype/freefont/FreeSansBold.ttf",
			"/usr/share/fonts/truetype/freefont/FreeSans.ttf",
			"/usr/local/share/fonts/DejaVuSans.ttf",
			"/usr/share/fonts/dejavu/DejaVuSans.ttf",
		}
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func main() {
	iconPath := os.Getenv("stamp_path_to_icons")
	version := os.Getenv("stamp_version")
	buildNumber := os.Getenv("stamp_build_number")
	fgColor := os.Getenv("foreground_color")
	bgColor := os.Getenv("background_color")
	fontInput := os.Getenv("stamp_font")

	log.Println("Version number to stamp:", version)
	log.Println("Build number to stamp:", buildNumber)
	log.Println("Foreground color is:", fgColor)
	log.Println("Background color is:", bgColor)

	font := findFont(fontInput)
	if font == "" {
		log.Println("No font specified or detected; ImageMagick will use its default font if available.")
	} else {
		log.Println("Using font:", font)
	}

	magick := findMagickBinary()

	log.Println("Finding icons from directory:", iconPath)

	files, err := os.ReadDir(iconPath)
	if err != nil {
		log.Printf("Could not read directory: %v", err)
		os.Exit(1)
	}

	filePaths := make([]string, 0)
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".png" {
			filePaths = append(filePaths, filepath.Join(iconPath, f.Name()))
		}
	}

	if len(filePaths) == 0 {
		log.Println("No .png icons found in directory:", iconPath)
		os.Exit(1)
	}

	log.Println(filePaths)

	for _, f := range filePaths {
		identifyBin, identifyPrefix := findIdentifyBinary()
		dimsOut, err := exec.Command(identifyBin, append(identifyPrefix, "-format", "%w,%h", f)...).Output()
		if err != nil {
			log.Printf("ImageMagick failed to read dimensions of %s: %v", f, err)
			os.Exit(1)
		}

		dims := strings.Split(string(dimsOut), ",")

		width, err1 := strconv.Atoi(dims[0])
		height, err2 := strconv.Atoi(dims[1])
		if err1 != nil || err2 != nil {
			log.Printf("Could not parse image dimensions %q for %s", string(dimsOut), f)
			os.Exit(1)
		}

		bannerH := int(math.Floor(float64(height) * 0.3))
		bannerDims := strconv.Itoa(width) + "x" + strconv.Itoa(bannerH)

		bannerCaption := "- " + version + "(" + buildNumber + ") -"

		args := make([]string, 0)
		if font != "" {
			args = append(args, "-font", font)
		}
		args = append(args,
			"-background", bgColor,
			"-fill", fgColor,
			"-gravity", "center",
			"-size", bannerDims,
			"caption:"+bannerCaption,
			f, "+swap",
			"-gravity", "south",
			"-composite", f)

		imgOut, err := exec.Command(magick, args...).CombinedOutput()
		if err != nil {
			fmt.Println(string(imgOut))
			log.Printf("ImageMagick failed to stamp %s: %v", f, err)
			os.Exit(1)
		}

	}

	os.Exit(0)
}
