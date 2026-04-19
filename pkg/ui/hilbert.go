package ui

import (
	"errors"
)

// Point represents a 2D coordinate.
type Point struct {
	X, Y int
}

// Hilbert converts a 1D index 'd' to a 2D coordinate within an 'n'x'n' square.
// 'n' must be a power of 2. If n is not a power of 2 or d is negative, it returns (0,0).
func Hilbert(n, d int) Point {
	if n <= 0 || (n&(n-1)) != 0 || d < 0 {
		return Point{0, 0}
	}
	p := Point{X: 0, Y: 0}
	t := d
	for s := 1; s < n; s *= 2 {
		rx := 1 & (t / 2)
		ry := 1 & (t ^ rx)
		rot(s, &p.X, &p.Y, rx, ry)
		p.X += s * rx
		p.Y += s * ry
		t /= 4
	}
	return p
}

// rot rotates and flips the quadrant appropriately.
func rot(n int, x, y *int, rx, ry int) {
	if ry == 0 {
		if rx == 1 {
			*x = n - 1 - *x
			*y = n - 1 - *y
		}
		// Swap x and y
		*x, *y = *y, *x
	}
}

// GenerateCurve returns a slice of Points representing a Hilbert Curve of order 'order'.
// The total number of points is 2^(2*order).
// Returns an error if order is negative or too large (over 15) to prevent OOM or overflows.
func GenerateCurve(order int) ([]Point, error) {
	if order < 0 {
		return nil, errors.New("order must be non-negative")
	}
	if order > 15 {
		return nil, errors.New("order too large (max 15)")
	}
	n := 1 << order
	total := n * n
	curve := make([]Point, total)
	for i := 0; i < total; i++ {
		curve[i] = Hilbert(n, i)
	}
	return curve, nil
}
