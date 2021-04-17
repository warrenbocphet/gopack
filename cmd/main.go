package main

import (
	"flag"
	"fmt"
	"github.com/vitsensei/gopack/packing"
	"io/ioutil"
	"os"
)

var (
	folderPath string
	imagePaths []string
	outputPath string
	width, height int
)

func main() {
	Init()

	packer := packing.CreatePacker(width, height)
	packer.GenerateMetas(imagePaths)
	packer.Pack()

	packer.ToFile(outputPath)
	fmt.Println("Done.")
}

func Init() {
	flag.StringVar(&folderPath, "p", "", "Full path to folder containing images")
	flag.StringVar(&folderPath, "path", "", "Full path to folder containing images")

	flag.IntVar(&width, "w", 2560, "Width of wallpaper")
	flag.IntVar(&width, "width", 2560, "Width of wallpaper")

	flag.IntVar(&height, "h", 1080, "Height of wallpaper")
	flag.IntVar(&height, "height", 1080, "Height of wallpaper")

	flag.StringVar(&outputPath, "o", "./result.png", "Output path")
	flag.StringVar(&outputPath, "output", "./result.png", "Output path")

	flag.Parse()

	if !isDir(folderPath) {
		panic("Invalid folder path.")
	} else {
		fmt.Println("Path to use: ", folderPath)
		imagePaths = collectPaths(folderPath)
	}

	if width <= 0 {
		panic("Invalid width")
	}

	if height <= 0 {
		panic("Invalid height")
	}
}

func isDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func collectPaths(p string) []string {
	files, _ := ioutil.ReadDir(p)

	paths := make([]string, len(files), len(files))
	for i, f := range files {
		paths[i] = p + string(os.PathSeparator) + f.Name()
	}

	return paths
}