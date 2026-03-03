package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/rky0930/nearby-gas-prices/internal/config"
	"github.com/rky0930/nearby-gas-prices/internal/geo"
	"github.com/rky0930/nearby-gas-prices/internal/nominatim"
	"github.com/rky0930/nearby-gas-prices/internal/opinet"
)

func main() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintln(out, "nearby-gas-prices - 주변 주유소 가격/최저가 조회 (Opinet + Nominatim)")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "USAGE")
		fmt.Fprintln(out, "  nearby-gas-prices --query \"소사역\" [options]")
		fmt.Fprintln(out, "  nearby-gas-prices --lat 37.48278 --lon 126.79565 [options]")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "CONFIG")
		fmt.Fprintln(out, "  (우선순위) 환경변수 > 설정 파일(~/.config/nearby-gas-prices/config.toml)")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "ENV")
		fmt.Fprintln(out, "  OPINET_KEY                (필수) 오피넷 무료 API KEY. 요청 파라미터 code 로 전달됨")
		fmt.Fprintln(out, "  NOMINATIM_USER_AGENT      (조건부) --query 사용 시 필요. OSM Nominatim 정책상 contact 포함 권장(미설정 시 403 가능)")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "NOTES")
		fmt.Fprintln(out, "  - Opinet aroundAll.do 제한: radius <= 5000m")
		fmt.Fprintln(out, "  - Opinet aroundAll.do 좌표: x,y 는 KATEC (WGS84 위경도 아님)")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "EXAMPLES")
		fmt.Fprintln(out, "  export OPINET_KEY=\"<YOUR_KEY>\"")
		fmt.Fprintln(out, "  export NOMINATIM_USER_AGENT=\"nearby-gas-prices/0.1 (contact: you@example.com)\"")
		fmt.Fprintln(out, "  nearby-gas-prices --query \"소사역\" --top 5")
		fmt.Fprintln(out, "  nearby-gas-prices --lat 37.48278 --lon 126.79565 --top 5 --json")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "OPTIONS")
		flag.PrintDefaults()
	}

	var (
		query   = flag.String("query", "", "장소명(예: 소사역)")
		lat     = flag.Float64("lat", 0, "위도 (WGS84)")
		lon     = flag.Float64("lon", 0, "경도 (WGS84)")
		radius  = flag.Int("radius", 5000, "검색 반경(m). 오피넷 aroundAll.do 는 최대 5000")
		prodcd  = flag.String("prodcd", "B027", "유종 코드 (예: B027=휘발유)")
		sortBy  = flag.Int("sort", 1, "정렬: 1=가격순, 2=거리순")
		top     = flag.Int("top", 5, "상위 N개 출력")
		jsonOut = flag.Bool("json", false, "JSON으로 출력")
	)
	flag.Parse()

	if *radius > 5000 {
		fatalf("radius는 오피넷 API 제한으로 최대 5000m 입니다 (입력값: %d)", *radius)
	}

	cfg, cfgPath, err := config.Load()
	if err != nil {
		fatalErr(err)
	}

	apiKey := strings.TrimSpace(os.Getenv("OPINET_KEY"))
	if apiKey == "" {
		apiKey = strings.TrimSpace(cfg.OpinetKey)
	}
	if apiKey == "" {
		fatalf("OPINET_KEY가 필요합니다. (환경변수 OPINET_KEY 또는 설정 파일 %s 의 opinet_key)", cfgPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	hc := &http.Client{Timeout: 15 * time.Second}

	var wgsLat, wgsLon float64
	if *query != "" {
		ua := strings.TrimSpace(os.Getenv("NOMINATIM_USER_AGENT"))
		if ua == "" {
			ua = strings.TrimSpace(cfg.NominatimUserAgent)
		}
		if ua == "" {
			fatalf("--query 사용 시 NOMINATIM_USER_AGENT가 필요합니다 (환경변수 또는 설정 파일 %s 의 nominatim_user_agent)", cfgPath)
		}
		res, err := nominatim.SearchOne(ctx, hc, *query, ua)
		if err != nil {
			fatalErr(err)
		}
		wgsLat, wgsLon = res.Lat, res.Lon
	} else {
		if *lat == 0 || *lon == 0 {
			fatalf("--query 또는 --lat/--lon 중 하나를 입력하세요")
		}
		wgsLat, wgsLon = *lat, *lon
	}

	x, y := geo.WGS84ToKATEC(wgsLon, wgsLat)

	items, err := opinet.AroundAll(ctx, hc, opinet.AroundAllParams{
		Code:   apiKey,
		X:      x,
		Y:      y,
		Radius: *radius,
		ProdCD: *prodcd,
		Sort:   *sortBy,
	})
	if err != nil {
		fatalErr(err)
	}

	// 정렬은 sort 파라미터가 하긴 하지만, 안전하게 한 번 더.
	switch *sortBy {
	case 2:
		sort.Slice(items, func(i, j int) bool { return items[i].DistanceM < items[j].DistanceM })
	default:
		sort.Slice(items, func(i, j int) bool { return items[i].Price < items[j].Price })
	}

	if *top > 0 && len(items) > *top {
		items = items[:*top]
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(items)
		return
	}

	if len(items) == 0 {
		fmt.Println("결과가 없습니다.")
		return
	}

	fmt.Printf("기준 좌표(WGS84): %.6f, %.6f\n", wgsLat, wgsLon)
	fmt.Printf("검색 반경: %dm (오피넷 aroundAll.do 최대 5000m)\n\n", *radius)

	for i, it := range items {
		fmt.Printf("%d) %s\n", i+1, it.Name)
		fmt.Printf("   가격: %d원\n", it.Price)
		fmt.Printf("   거리: %.0fm\n", it.DistanceM)

		// 오피넷 응답 좌표는 KATEC이므로 역변환해서 링크 생성
		lat2, lon2 := geo.KATECToWGS84(it.X, it.Y)
		if lat2 != 0 && lon2 != 0 {
			fmt.Printf("   지도(네이버): %s\n", geo.NaverMapLink(it.Name, lat2, lon2))
		}
		fmt.Println()
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(2)
}

func fatalErr(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
