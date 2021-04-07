package packing

import (
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

func CreatePacker(width, height int) Packer {
	var partitions []Partition
	partitions = append(partitions, CreatePartition(Point{0, 0}, Point{width, height}))
	debugFie, _ := os.Create("debug.txt")
	_ = debugFie.Truncate(0)

	return Packer{
		partitions: partitions,
		img: image.NewRGBA(image.Rectangle{
			Min: image.Point{X: 0, Y: 0},
			Max: image.Point{X: width, Y: height},
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
}

func (p *Packer) AddImage(newImg image.RGBA) bool {
	// Step 1: Check for available partition
	width, height := newImg.Bounds().Dx(), newImg.Bounds().Dy()
	for i := range p.partitions {
		if p.partitions[i].BigEnough(width, height) {
			p.Debug("Found a suitable partition")
			for j := range p.partitions {
				p.Debug("Rectangle " + strconv.Itoa(j) + ": " + strconv.Itoa(p.partitions[j].P1().X()) + ", " + strconv.Itoa(p.partitions[j].P1().Y()) + ", " + strconv.Itoa(p.partitions[j].P2().X()) + ", " + strconv.Itoa(p.partitions[j].P2().Y()))
			}

			pivotPoint := p.partitions[i].P1()
			dp := image.Point{X: pivotPoint.X(), Y: pivotPoint.Y()}
			// Add the image to Packer's image (code from https://blog.golang.org/image-draw)
			r := image.Rectangle{Min: dp, Max: dp.Add(newImg.Bounds().Size())}
			draw.Draw(p.img, r, &newImg, newImg.Bounds().Min, draw.Src)
			newPartition1, newPartition2 := p.partitions[i].AddRectangle(newImg.Bounds().Dx(), newImg.Bounds().Dy(), false)

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

func (p *Packer) OpenAndAddImage(path string) (bool, error) {
	p.Debug("Finding suitable partition for " + path)

	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return false, err
	}

	rgbaImg := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
	draw.Draw(rgbaImg, rgbaImg.Bounds(), img, img.Bounds().Min, draw.Src)
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
