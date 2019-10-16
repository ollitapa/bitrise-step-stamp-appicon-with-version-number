package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

func main() {
	iconPath := os.Getenv("stamp_path_to_icons")
	version := os.Getenv("stamp_version")
	buildNumber := os.Getenv("stamp_build_number")
	fgColor := os.Getenv("foreground_color")
	bgColor := os.Getenv("background_color")

	fmt.Println("Version number to stamp:", version)
	fmt.Println("Build number to stamp:", buildNumber)
	fmt.Println("Foreground color is:", fgColor)
	fmt.Println("Background color is:", bgColor)

	fmt.Println("Finding icons from directory:", iconPath)

	files, err := ioutil.ReadDir(iconPath)
	if err != nil {
		fmt.Println("Could not read directory!")
		os.Exit(1)
	}

	filePaths := make([]string, 0)
	for _, f := range files {
		if path.Ext(f.Name()) == ".png" {
			filePaths = append(filePaths, path.Join(iconPath, f.Name()))
		}
	}

	fmt.Println(filePaths)

	for _, f := range filePaths {
		dimsOut, err := exec.Command("identify", "-format", "%w,%h", f).Output()
		if err != nil {
			fmt.Println("ImageMagick failed!")
			os.Exit(1)
		}

		dims := strings.Split(string(dimsOut), ",")

		width, err1 := strconv.Atoi(dims[0])
		height, err2 := strconv.Atoi(dims[1])
		if err1 != nil && err2 != nil {
			os.Exit(1)
		}

		bannerH := int(math.Floor(float64(height) * 0.3))
		bannerDims := strconv.Itoa(width) + "x" + strconv.Itoa(bannerH)

		bannerCaption := "- " + version + "(" + buildNumber + ")" + " -"

		imgOutString, error := exec.Command("convert",
			"-background", bgColor,
			"-fill", fgColor,
			"-gravity", "center",
			"-size", bannerDims,
			"caption:"+bannerCaption,
			f, "+swap",
			"-gravity", "south",
			"-composite", f).CombinedOutput()
		if error != nil {
			fmt.Println(string(imgOutString))
			fmt.Println(error)
			fmt.Println("ImageMagick failed!")
			os.Exit(1)
		}

	}

	//
	// --- Exit codes:
	// The exit code of your Step is very important. If you return
	//  with a 0 exit code `bitrise` will register your Step as "successful".
	// Any non zero exit code will be registered as "failed" by `bitrise`.
	os.Exit(0)
}
