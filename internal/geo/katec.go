package geo

import (
	"fmt"
	"net/url"

	"github.com/wroge/wgs84"
)

// KATEC projection parameters used widely in Korea.
// NOTE: Opinet's aroundAll.do expects x,y in KATEC.
//
// Proj4-style params (reference commonly used in community examples):
// +proj=tmerc +lat_0=38 +lon_0=128 +k=0.9999 +x_0=400000 +y_0=600000 +ellps=bessel +units=m
// plus a Helmert transform (towgs84): -115.80,474.99,674.11,1.16,-2.31,-1.63,6.43
var katecCRS = func() wgs84.ProjectedReferenceSystem {
	b := wgs84.Bessel{}
	// Helmert parameters: tx,ty,tz, rx,ry,rz, ds(ppm)
	d := wgs84.Helmert(b.A(), b.Fi(), -115.80, 474.99, 674.11, 1.16, -2.31, -1.63, 6.43)
	return d.TransverseMercator(128, 38, 0.9999, 400000, 600000)
}()

// WGS84ToKATEC converts WGS84 lon/lat to KATEC x/y (meters).
func WGS84ToKATEC(lon, lat float64) (x, y float64) {
	x, y, _ = wgs84.LonLat().To(katecCRS)(lon, lat, 0)
	return x, y
}

// KATECToWGS84 converts KATEC x/y (meters) back to WGS84 lat/lon.
func KATECToWGS84(x, y float64) (lat, lon float64) {
	lon, lat, _ = katecCRS.To(wgs84.LonLat())(x, y, 0)
	return lat, lon
}

func NaverMapLink(name string, lat, lon float64) string {
	q := url.QueryEscape(name)
	// Keep it simple: use search with center.
	return fmt.Sprintf("https://map.naver.com/v5/search/%s?c=%.6f,%.6f,15,0,0,0,dh", q, lon, lat)
}
