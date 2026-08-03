package site

import (
	"math"
	"testing"
)

func TestHeartLine3DPointsFollowParametricCurve(t *testing.T) {
	points := heartLine3DPoints()
	if len(points) != 321 {
		t.Fatalf("point count = %d, want 321", len(points))
	}
	for index, want := range map[int][3]float64{
		0:   {0, 0, 5},
		80:  {16, -2.4, 4},
		160: {0, 0, -17},
		320: {0, 0, 5},
	} {
		got := points[index]
		for axis, value := range []float64{got.X, got.Y, got.Z} {
			if math.Abs(value-want[axis]) > 1e-9 {
				t.Errorf("point %d axis %d = %.12f, want %.12f", index, axis, value, want[axis])
			}
		}
	}
}
