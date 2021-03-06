package packing

type Point struct {
	x int
	y int
}

func (p Point) X() int {
	return p.x
}

func (p Point) Y() int {
	return p.y
}

// Partition represent the single cell that an image can be inserted in.
type Partition struct {
	p1    Point
	p2    Point
	ratio float32
}

func CreatePartition(p1, p2 Point) Partition {
	if p2.Y()-p1.Y() == 0 {
		return Partition{
			p1:    p1,
			p2:    p2,
			ratio: 0,
		}
	}

	return Partition{
		p1:    p1,
		p2:    p2,
		ratio: float32(p2.X()-p1.X()) / float32(p2.Y()-p1.Y()),
	}
}

// A high level function that split up a partition based on the image added in the partition
func (p Partition) AddRectangle(width, height int, hMajor bool) (Partition, Partition) {
	return p.split(Point{p.p1.X() + width, p.p1.Y() + height}, hMajor)
}

func (p Partition) split(point Point, hMajor bool) (Partition, Partition) {
	var partition1, partition2 Partition

	if hMajor {
		// Case 1: horizontal major
		p1, p2 := Point{point.X(), p.p1.Y()}, Point{p.p2.X(), point.Y()}
		pA, pB := Point{p.p1.X(), point.Y()}, p.p2

		partition1 = CreatePartition(p1, p2)
		partition2 = CreatePartition(pA, pB)

	} else {
		// Case 2: vertical major
		p1, p2 := Point{point.X(), p.p1.Y()}, p.p2
		pA, pB := Point{p.p1.X(), point.Y()}, Point{point.X(), p.p2.Y()}

		partition1 = CreatePartition(p1, p2)
		partition2 = CreatePartition(pA, pB)
	}

	return partition1, partition2
}

func (p Partition) BigEnough(width, height int) bool {
	dw := p.p2.X() - p.p1.X()
	dh := p.p2.Y() - p.p1.Y()

	if width <= dw && height <= dh {
		return true
	}

	return false
}

func (p Partition) Size() int {
	return (p.p2.X() - p.p1.X()) * (p.p2.Y() - p.p1.Y())
}

func (p Partition) Width() int {
	return p.p2.X() - p.p1.X()
}

func (p Partition) Height() int {
	return p.p2.Y() - p.p1.Y()
}

func (p Partition) Ratio() float32 {
	return p.ratio
}

func (p Partition) P1() Point {
	return p.p1
}

func (p Partition) P2() Point {
	return p.p2
}

func (p Partition) IsValid() bool {
	return p.Size() > 0 && p.Width() > tolWidth && p.Height() > tolHeight
}
