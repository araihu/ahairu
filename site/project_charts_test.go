package site

import (
	"math"
	"testing"
)

func TestHeartSurface3DPointsFollowOrderedParametricMesh(t *testing.T) {
	points := heartSurface3DPoints()
	if len(points) != heartSurface3DRows*heartSurface3DColumns {
		t.Fatalf("point count = %d, want %d", len(points), heartSurface3DRows*heartSurface3DColumns)
	}
	for index, want := range map[int][3]float64{
		0:                             {0, 5.2, -1.5},
		32:                            {0, 0, 5.16},
		12*heartSurface3DColumns + 32: {13.5, 0, 3.680000033378601},
		24*heartSurface3DColumns + 32: {0, 0, -15.64},
	} {
		got := points[index]
		for axis, value := range []float64{got.X, got.Y, got.Z} {
			if math.Abs(value-want[axis]) > 1e-8 {
				t.Errorf("point %d axis %d = %.12f, want %.12f", index, axis, value, want[axis])
			}
		}
	}

	firstRow := points[:heartSurface3DColumns]
	lastRow := points[(heartSurface3DRows-1)*heartSurface3DColumns:]
	for index := range firstRow {
		if math.Abs(firstRow[index].X-lastRow[index].X) > 1e-9 || math.Abs(firstRow[index].Y-lastRow[index].Y) > 1e-9 || math.Abs(firstRow[index].Z-lastRow[index].Z) > 1e-9 {
			t.Fatalf("mesh seam differs at column %d", index)
		}
	}
}
