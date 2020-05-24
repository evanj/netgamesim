package intersect

import "testing"

func TestPathBox(t *testing.T) {
	boxCenter := Point{10.0, 10.0}
	const boxSize = 5.0

	type path struct {
		p1 Point
		p2 Point
	}

	shouldNotIntersect := []path{
		path{Point{0, 0}, Point{5, 5}},

		// vertical
		path{Point{0, 0}, Point{0, 20}},
		// horizontal
		path{Point{0, 0}, Point{20, 0}},

		// point not in the box
		path{Point{0, 0}, Point{0, 0}},
	}
	for i, testPath := range shouldNotIntersect {
		for _, path := range []path{testPath, path{testPath.p2, testPath.p1}} {
			intersects := PathBox(path.p1, path.p2, boxCenter, boxSize)
			if intersects {
				t.Errorf("%d: %s->%s should not intersect box center=%s size=%f",
					i, path.p2, path.p1, boxCenter, boxSize)
			}
		}
	}

	shouldIntersect := []path{
		// just barely touches the corner: still counts!
		path{Point{0, 0}, Point{7.5, 7.5}},

		path{Point{10, 0}, Point{10, 20}},
		path{Point{0, 10}, Point{20, 10}},

		// contained segment
		path{Point{9, 9}, Point{11, 11}},
		// contained point
		path{Point{12, 12}, Point{12, 12}},
	}
	for i, testPath := range shouldIntersect {
		for _, path := range []path{testPath, path{testPath.p2, testPath.p1}} {
			intersects := PathBox(path.p1, path.p2, boxCenter, boxSize)
			if !intersects {
				t.Errorf("%d: %s->%s should intersect box center=%s size=%f",
					i, path.p1, path.p2, boxCenter, boxSize)
			}
		}
	}

}
