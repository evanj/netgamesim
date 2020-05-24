package intersect

import (
	"fmt"
	"log"
	"math"
)

const verbose = false

func debugf(msg string, args ...interface{}) {
	if verbose {
		log.Printf(msg, args...)
	}
}

// Point represents a single point.
type Point struct {
	X float64
	Y float64
}

func (p Point) String() string {
	return fmt.Sprintf("(%f,%f)", p.X, p.Y)
}

// PointBox returns true if p is contained in the AABB with center and diameter.
func PointBox(p Point, center Point, diameter float64) bool {
	// compute bounding box
	x0 := center.X - diameter/2
	y0 := center.Y - diameter/2
	x1 := center.X + diameter/2
	y1 := center.Y + diameter/2

	// hit if point is in the box
	return (x0 <= p.X && p.X <= x1) &&
		(y0 <= p.Y && p.Y <= y1)
}

// PathBox returns true if path segment p0 -> p1 intersects the AABB with center and diameter.
func PathBox(p0 Point, p1 Point, center Point, diameter float64) bool {
	// uses the "slab intersection" algorithm: using the linear interpolation form of the line
	// x = x0 + t(x1 - x0)
	// y = y0 + t(y1 - y0)
	// then compute tMin / tMax after comparing against all 4 edges
	// http://www.pbr-book.org/3ed-2018/Shapes/Basic_Shape_Interface.html#RayndashBoundsIntersections

	// compute bounding box
	x0 := center.X - diameter/2
	y0 := center.Y - diameter/2
	x1 := center.X + diameter/2
	y1 := center.Y + diameter/2

	xVec := p1.X - p0.X
	yVec := p1.Y - p0.Y

	// intersection with top line y0
	t0 := (y0 - p0.Y) / yVec
	// intersection with bottom line y1
	t1 := (y1 - p0.Y) / yVec
	if t0 > t1 {
		t1, t0 = t0, t1
	}

	tMin := t0
	tMax := t1

	// intersection with left line x0
	t0 = (x0 - p0.X) / xVec
	// intersection with right line x1
	t1 = (x1 - p0.X) / xVec
	if t0 > t1 {
		t1, t0 = t0, t1
	}

	// intersect tMin/tMax with t0/t1
	tMin = math.Max(tMin, t0)
	tMax = math.Min(tMax, t1)

	if tMin > tMax {
		// no intersection of parameters: this means the ray does not intersect the box
		return false
	}

	// intersect tMin/tMax with [0, 1]
	tMin = math.Max(tMin, 0)
	tMax = math.Min(tMax, 1)

	if tMin > tMax {
		// no intersection of parameters in the range [0, 1]
		return false
	}

	debugf("%s->%s intersects box (%f,%f)x(%f,%f); t=[%f,%f] point min = (%f,%f)",
		p0, p1, x0, y0, x1, x1,
		tMin, tMax, p0.X+xVec*tMin, p0.Y+yVec*tMin)
	return true
}
