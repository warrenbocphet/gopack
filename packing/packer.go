package packing

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"os"
	"sort"
	"strconv"
)

type BySize []Partition

func (a BySize) Len() int {
	return len(a)
}

func (a BySize) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a BySize) Less(i, j int) bool {
	return a[i].Size() < a[j].Size()
}

type Packer struct {
	partitions  []Partition
	metas       []Meta
	img         *image.RGBA
	numberOfImg int
	debug       *os.File
}

func CreatePacker(width, height uint) Packer {
	var partitions []Partition
	partitions = append(partitions, CreatePartition(Point{0, 0}, Point{width, height}))
	debugFie, _ := os.Create("debug.txt")
	_ = debugFie.Truncate(0)

	return Packer{
		partitions: partitions,
		img: image.NewRGBA(image.Rectangle{
			Min: image.Point{X: 0, Y: 0},
			Max: image.Point{X: int(width), Y: int(height)},
		}),
		numberOfImg: 0,
		debug:       debugFie,
	}
}

func (p *Packer) GenerateMetas(paths []string) {
	p.metas = GenerateMetas(paths)
}

func (p *Packer) Pack() {
	// First pass, try to push full size img to main canvas
	i := 0
	for true {
		if i == len(p.metas) {
			break
		}

		success, err := p.OpenAndAddImage(p.metas[i].path)
		must(err)

		if success {
			// Delete this meta
			p.metas = append(p.metas[:i], p.metas[i+1:]...)
		} else {
			i++
		}
	}

	// Second pass, try to fit the rest of the partition with photo with ~ same ratio
	//i = 0
	//for true {
	//	if len(p.partitions) == 0 || len(p.metas) == 0 {
	//		break
	//	}
	//
	//	bestMetaInd := 0
	//	bestRatioDiff := 1000.0
	//	for j := range p.metas {
	//		ratioDiff := math.Abs(float64(p.metas[j].Ratio() - p.partitions[i].Ratio()))
	//		if ratioDiff < bestRatioDiff {
	//			bestRatioDiff = ratioDiff
	//			bestMetaInd = j
	//		}
	//
	//		meta := p.metas[bestMetaInd]
	//		img, err := openImage(meta.path)
	//		must(err)
	//
	//		newImg := resize.Thumbnail(p.partitions[i].Width(), p.partitions[i].Height(), img, resize.Bilinear)
	//		rgbaImg := image.NewRGBA(image.Rect(0, 0, newImg.Bounds().Dx(), newImg.Bounds().Dy()))
	//		draw.Draw(rgbaImg, rgbaImg.Bounds(), newImg, newImg.Bounds().Min, draw.Src)
	//
	//	}
	//
	//}
}

func (p *Packer) AddImage(newImg image.RGBA) bool {
	// Step 1: Check for available partition
	width, height := newImg.Bounds().Dx(), newImg.Bounds().Dy()
	for i := range p.partitions {
		if p.partitions[i].BigEnough(uint(width), uint(height)) {
			p.Debug("Found a suitable partition")
			for j := range p.partitions {
				p.Debug("Rectangle " + strconv.Itoa(j) + ": " + fmt.Sprint(p.partitions[j].P1().X()) + ", " + fmt.Sprint(p.partitions[j].P1().Y()) + ", " + fmt.Sprint(p.partitions[j].P2().X()) + ", " + fmt.Sprint(p.partitions[j].P2().Y()))
			}

			pivotPoint := p.partitions[i].P1()
			dp := image.Point{X: int(pivotPoint.X()), Y: int(pivotPoint.Y())}
			// Add the image to Packer's image (code from https://blog.golang.org/image-draw)
			r := image.Rectangle{Min: dp, Max: dp.Add(newImg.Bounds().Size())}
			draw.Draw(p.img, r, &newImg, newImg.Bounds().Min, draw.Src)
			newPartition1, newPartition2 := p.partitions[i].AddRectangle(uint(newImg.Bounds().Dx()), uint(newImg.Bounds().Dy()), false)

			// Delete current partition and add two new partitions from the split
			p.partitions = append(p.partitions[:i], p.partitions[i+1:]...)

			if newPartition1.Size() > 0 {
				p.partitions = append(p.partitions, newPartition1)
			}
			if newPartition2.Size() > 0 {
				p.partitions = append(p.partitions, newPartition2)
			}

			// Sort the partition
			sort.Sort(BySize(p.partitions))

			p.numberOfImg++
			return true
		}
	}

	p.Debug("Cannot find suitable partition")
	return false
}

func openImage(path string) (*image.RGBA, error) {
	f, err := os.Open(path)
	if err != nil {
		return &image.RGBA{}, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return &image.RGBA{}, err
	}

	rgbaImg := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
	draw.Draw(rgbaImg, rgbaImg.Bounds(), img, img.Bounds().Min, draw.Src)

	return rgbaImg, err
}

func (p *Packer) OpenAndAddImage(path string) (bool, error) {
	p.Debug("Finding suitable partition for " + path)

	rgbaImg, err := openImage(path)
	must(err)

	success := p.AddImage(*rgbaImg)

	return success, nil
}

func (p Packer) ToFile(path string) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = png.Encode(f, p.img)
	if err != nil {
		panic(err)
	}
}

func (p Packer) Debug(msg string) {
	p.debug.WriteString(msg + "\n")
}

func (p Packer) Clean() {
	p.debug.Close()
}
