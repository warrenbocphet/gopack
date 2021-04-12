package packing

import (
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"math"
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
	// Every time an image is added, the partitions will be sorted by size
	// in ascending order (best fit).
	sort.Sort(BySize(p.partitions))
	i := 0
	for true {
		// Iterate through all images found in the folder
		// to find fit partition
		if i == len(p.metas) {
			break
		}

		success, err := p.openAndAddImage(p.metas[i].path)
		must(err)

		if success {
			// Delete this meta
			p.metas = append(p.metas[:i], p.metas[i+1:]...)
		} else {
			i++
		}
	}

	// Second pass, try to fit the rest of the partition with photo with ~ same ratio
	// By now, most "good" partition is already used. We need to sort the partition in
	// descending order (based on size).
	sort.Sort(sort.Reverse(BySize(p.partitions)))
	i = 0
	for true {
		// We will keep trying to add stuffs into the main image until
		// we ran out of images to add, or we ran out partitions
		if len(p.partitions) == 0 || len(p.metas) == 0 {
			break
		}

		// Since we have the ability to resize images, it is almost guarantee that any images can fit in any partition.
		// The harder part is to know what is the best image to fit in a partition.
		// The metric I used to measure is: [ratioDiff = image.ratio - partition.ratio]
		// Smaller ratioDiff the better
		// One problem with this metric is partition that are way too small for an image might still
		// be chosen due to ratioDiff still very small.
		bestMetaInd := 0
		bestRatioDiff := 1000.0
		for j := range p.metas {
			ratioDiff := math.Abs(float64(p.metas[j].Ratio() - p.partitions[i].Ratio()))
			if ratioDiff < bestRatioDiff {
				bestRatioDiff = ratioDiff
				bestMetaInd = j
			}
		}

		meta := p.metas[bestMetaInd]
		img, err := openImage(meta.path)
		must(err)

		newImg := resize.Thumbnail(p.partitions[i].Width(), p.partitions[i].Height(), img, resize.Bilinear)
		rgbaImg := toRGBA(newImg)

		p.addImageToPartition(rgbaImg, i, true, true)
		p.metas = append(p.metas[:bestMetaInd], p.metas[bestMetaInd+1:]...)
	}
}

// Add the image to selected partition and sort by size
func (p *Packer) addImageToPartition(newImg *image.RGBA, partitionInd int, isDescending, autoFit bool) {
	var img image.Image

	if autoFit {
		img = resize.Thumbnail(p.partitions[partitionInd].Width(), p.partitions[partitionInd].Height(), newImg, resize.Bilinear)
		newImg = toRGBA(img)
	}

	if int(p.partitions[partitionInd].Width())-newImg.Bounds().Dx() < 10 && int(p.partitions[partitionInd].Height())-newImg.Bounds().Dy() < 10 {
		widthRatio := float32(p.partitions[partitionInd].Width()) / float32(newImg.Bounds().Dx())
		heightRatio := float32(p.partitions[partitionInd].Height()) / float32(newImg.Bounds().Dy())
		if widthRatio < heightRatio {
			img = resize.Resize(0, p.partitions[partitionInd].Height(), newImg, resize.Bilinear)
			newImg = toRGBA(img)
		} else {
			img = resize.Resize(p.partitions[partitionInd].Width(), 0, newImg, resize.Bilinear)
			newImg = toRGBA(img)
		}
	} else if int(p.partitions[partitionInd].Width())-newImg.Bounds().Dx() < 10 {
		img = resize.Resize(p.partitions[partitionInd].Width(), 0, newImg, resize.Bilinear)
		newImg = toRGBA(img)
	} else if int(p.partitions[partitionInd].Width())-newImg.Bounds().Dx() < 10 {
		newImg = toRGBA(img)
	}

	var minWidth, minHeight int
	if newImg.Bounds().Dx() < int(p.partitions[partitionInd].Width()) {
		minWidth = newImg.Bounds().Dx()
	} else {
		minWidth = int(p.partitions[partitionInd].Width())
	}

	if newImg.Bounds().Dy() < int(p.partitions[partitionInd].Height()) {
		minHeight = newImg.Bounds().Dy()
	} else {
		minHeight = int(p.partitions[partitionInd].Height())
	}

	// Add the image to Packer's image (code from https://blog.golang.org/image-draw)
	pivotPoint := p.partitions[partitionInd].P1()
	dp := image.Point{X: int(pivotPoint.X()), Y: int(pivotPoint.Y())}

	r := image.Rectangle{Min: dp, Max: dp.Add(image.Point{X: minWidth, Y: minHeight})}
	draw.Draw(p.img, r, newImg, newImg.Bounds().Min, draw.Src)

	// Keeping track of the partition by notifying the used are of this partition to itself
	newPartition1, newPartition2 := p.partitions[partitionInd].AddRectangle(uint(minWidth), uint(minHeight), false)

	// Delete current partition and add two new partitions from the split
	p.partitions = append(p.partitions[:partitionInd], p.partitions[partitionInd+1:]...)

	if newPartition1.IsValid() {
		p.partitions = append(p.partitions, newPartition1)
	}
	if newPartition2.IsValid() {
		p.partitions = append(p.partitions, newPartition2)
	}

	// Sort the partition
	if isDescending {
		sort.Sort(sort.Reverse(BySize(p.partitions)))
	} else {
		sort.Sort(BySize(p.partitions))
	}

	p.numberOfImg++
}

// A high level call to add another image to the packer.
// This will need to iterate through all partition to find appropriate partition to fit the data in
func (p *Packer) addImage(newImg image.RGBA) bool {
	// Step 1: Check for available partition
	width, height := newImg.Bounds().Dx(), newImg.Bounds().Dy()
	for i := range p.partitions {
		if p.partitions[i].BigEnough(uint(width), uint(height)) {
			p.Debug("Found a suitable partition")
			for j := range p.partitions {
				p.Debug("Rectangle " + strconv.Itoa(j) + ": " + fmt.Sprint(p.partitions[j].P1().X()) + ", " + fmt.Sprint(p.partitions[j].P1().Y()) + ", " + fmt.Sprint(p.partitions[j].P2().X()) + ", " + fmt.Sprint(p.partitions[j].P2().Y()))
			}

			p.addImageToPartition(&newImg, i, false, false)
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

	rgbaImg := toRGBA(img)

	return rgbaImg, err
}

func (p *Packer) openAndAddImage(path string) (bool, error) {
	p.Debug("Finding suitable partition for " + path)

	rgbaImg, err := openImage(path)
	must(err)

	success := p.addImage(*rgbaImg)

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

func toRGBA(img image.Image) *image.RGBA {
	rgbaImg := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
	draw.Draw(rgbaImg, rgbaImg.Bounds(), img, img.Bounds().Min, draw.Src)

	return rgbaImg
}
