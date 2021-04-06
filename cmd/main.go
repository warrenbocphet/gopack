package main

import (
	"fmt"
	"github.com/vitsensei/gopack/packing"
	"path/filepath"
)

func main()  {
	paths, _ := filepath.Glob("/home/atran/Desktop/inspiration/*")
	packer := packing.CreatePacker(2560, 1080)
	defer packer.Clean()

	for i := range paths {
		fmt.Println("Ading", paths[i])
		_ = packer.OpenAndAddImage(paths[i])
	}
	packer.ToFile("result.png")
	fmt.Println("Done.")
}
