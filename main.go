package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	iconPath := os.Getenv("stamp_path_to_icons")
	version := os.Getenv("stamp_version")
	buildNumber := os.Getenv("stamp_build_number")
	fgColor := os.Getenv("foreground_color")
	bgColor := os.Getenv("background_color")

	log.Println("Version number to stamp:", version)
	log.Println("Build number to stamp:", buildNumber)
	log.Println("Foreground color is:", fgColor)
	log.Println("Background color is:", bgColor)

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
		dimsOut, err := exec.Command("identify", "-format", "%w,%h", f).Output()
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

		bannerCaption := "- " + version + "(" + buildNumber + ")" + " -"

		imgOut, err := exec.Command("convert",
			"-background", bgColor,
			"-fill", fgColor,
			"-gravity", "center",
			"-size", bannerDims,
			"caption:"+bannerCaption,
			f, "+swap",
			"-gravity", "south",
			"-composite", f).CombinedOutput()
		if err != nil {
			fmt.Println(string(imgOut))
			log.Printf("ImageMagick failed to stamp %s: %v", f, err)
			os.Exit(1)
		}

	}

	//
	// --- Exit codes:
	// The exit code of your Step is very important. If you return
	// with a 0 exit code `bitrise` will register your Step as "successful".
	// Any non zero exit code will be registered as "failed" by `bitrise`.
	os.Exit(0)
}
