package main

import (
	"fmt"
	"github.com/vitsensei/gopack/packing"
	"path/filepath"
)

func main() {
	paths, _ := filepath.Glob("/home/atran/Desktop/inspiration/*")
	packer := packing.CreatePacker(2560, 1080)

	packer.GenerateMetas(paths)
	packer.Pack()

	packer.ToFile("result.png")
	fmt.Println("Done.")
}
