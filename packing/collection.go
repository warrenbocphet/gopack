package packing

import (
	"image"
	"os"
	"sort"
)

type Metas []Meta

func (a Metas) Len() int {
	return len(a)
}

func (a Metas) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a Metas) Less(i, j int) bool {
	return a[i].width*a[i].height < a[j].width*a[i].height
}

func GenerateMetas(paths []string) []Meta {
	metas := make([]Meta, len(paths), len(paths))
	for i := range paths {
		f, err := os.Open(paths[i])
		must(err)

		img, _, err := image.Decode(f)
		must(err)

		metas[i] = Meta{
			path:   paths[i],
			width:  img.Bounds().Dx(),
			height: img.Bounds().Dy(),
			ratio:  float32(img.Bounds().Dx()) / float32(img.Bounds().Dy()),
		}

		err = f.Close()
		must(err)
	}
	sort.Sort(Metas(metas))

	return metas
}

type Meta struct {
	path   string
	width  int
	height int
	ratio  float32
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
