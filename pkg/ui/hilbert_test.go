package ui

import (
	"testing"
)

func TestHilbert(t *testing.T) {
	// Simple test cases for order 2 (n=2)
	// d=0 -> (0,0)
	// d=1 -> (0,1)
	// d=2 -> (1,1)
	// d=3 -> (1,0)
	tests := []struct {
		n, d int
		want Point
	}{
		{2, 0, Point{0, 0}},
		{2, 1, Point{0, 1}},
		{2, 2, Point{1, 1}},
		{2, 3, Point{1, 0}},
	}

	for _, tt := range tests {
		got := Hilbert(tt.n, tt.d)
		if got != tt.want {
			t.Errorf("Hilbert(%d, %d) = %v; want %v", tt.n, tt.d, got, tt.want)
		}
	}
}

func TestGenerateCurve(t *testing.T) {
	order := 1
	curve1, err := GenerateCurve(order)
	if err != nil {
		t.Fatalf("GenerateCurve(%d) error: %v", order, err)
	}
	if len(curve1) != 4 {
		t.Errorf("GenerateCurve(%d) returned %d points; want 4", order, len(curve1))
	}

	order2 := 2
	curve2, err := GenerateCurve(order2)
	if err != nil {
		t.Fatalf("GenerateCurve(%d) error: %v", order, err)
	}
	if len(curve2) != 16 {
		t.Errorf("GenerateCurve(%d) returned %d points; want 16", order2, len(curve2))
	}
}

func TestHilbertVulnerabilities(t *testing.T) {
	t.Run("LargeOrderError", func(t *testing.T) {
		// order = 32 should now return an error
		_, err := GenerateCurve(32)
		if err == nil {
			t.Errorf("GenerateCurve(32) should have returned an error")
		}
	})

	t.Run("NegativeD", func(t *testing.T) {
		// d < 0 should return (0,0)
		p := Hilbert(2, -1)
		if p != (Point{0, 0}) {
			t.Errorf("Hilbert(2, -1) should be {0,0}")
		}
	})

	t.Run("NonPowerOfTwoN", func(t *testing.T) {
		// n not a power of 2 should return (0,0)
		p := Hilbert(3, 1)
		if p != (Point{0, 0}) {
			t.Errorf("Hilbert(3, 1) should be {0,0}")
		}
	})
}
