package geo

import "testing"

func TestWGS84ToKATEC_SosaStation(t *testing.T) {
	// 소사역 근처(대화에서 사용했던 값)
	lat := 37.48278
	lon := 126.79565

	x, y := WGS84ToKATEC(lon, lat)

	// proj4로 검증했던 값: x=293684.8, y=542979.5
	if diff(x, 293684.8) > 15.0 || diff(y, 542979.5) > 15.0 {
		t.Fatalf("unexpected KATEC: x=%.3f y=%.3f", x, y)
	}
}

func diff(a, b float64) float64 {
	if a > b {
		return a - b
	}
	return b - a
}
