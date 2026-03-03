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
		fmt.Fprintln(out, "                          키 발급: https://www.opinet.co.kr/user/custapi/custApiInfo.do")
		fmt.Fprintln(out, "  NOMINATIM_USER_AGENT      (조건부) --query 사용 시 필요. OSM Nominatim 정책상 contact 포함 권장(미설정 시 403 가능)")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "NOTES")
		fmt.Fprintln(out, "  - Opinet aroundAll.do 제한: radius <= 5000m")
		fmt.Fprintln(out, "  - Opinet aroundAll.do 좌표: x,y 는 KATEC (WGS84 위경도 아님)")
		fmt.Fprintln(out, "  - 코드 목록 출력: nearby-gas-prices --list-codes")
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
		query     = flag.String("query", "", "장소명(예: 소사역)")
		lat       = flag.Float64("lat", 0, "위도 (WGS84)")
		lon       = flag.Float64("lon", 0, "경도 (WGS84)")
		radius    = flag.Int("radius", 5000, "검색 반경(m). 오피넷 aroundAll.do 는 최대 5000")
		prodcd    = flag.String("prodcd", "B027", "유종 코드. 예: B027=휘발유, D047=경유, K015=LPG(부탄). 여러 개는 콤마로(B027,D047), 또는 all")
		brand     = flag.String("brand", "", "상표(브랜드) 필터. 예: SOL,SKE,GSC,HDO,RTE... (콤마로 여러 개)")
		withAvg   = flag.Bool("with-avg", false, "전국 평균가(avgAllPrice) 대비 차이도 함께 표시")
		detail    = flag.Bool("detail", false, "상위 1개 주유소에 대해 상세정보(detailById)도 함께 출력")
		listCodes = flag.Bool("list-codes", false, "유종/브랜드 코드 목록 출력 후 종료")
		sortBy    = flag.Int("sort", 1, "정렬: 1=가격순, 2=거리순")
		top       = flag.Int("top", 5, "상위 N개 출력")
		jsonOut   = flag.Bool("json", false, "JSON으로 출력")
	)
	flag.Parse()

	if *listCodes {
		printCodes()
		return
	}

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
		fatalf("OPINET_KEY가 필요합니다. (환경변수 OPINET_KEY 또는 설정 파일 %s 의 opinet_key)\n키 발급 안내: https://www.opinet.co.kr/user/custapi/custApiInfo.do", cfgPath)
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

	prodList := parseProdList(*prodcd)
	brandSet := parseCSVSet(*brand)

	var avgByProd map[string]float64
	if *withAvg {
		avgByProd, err = opinet.AvgAllPrice(ctx, hc, apiKey)
		if err != nil {
			fatalErr(err)
		}
	}

	type outBlock struct {
		ProdCD   string           `json:"prodcd"`
		Stations []opinet.Station `json:"stations"`
	}
	blocks := make([]outBlock, 0, len(prodList))

	for _, prod := range prodList {
		items, err := opinet.AroundAll(ctx, hc, opinet.AroundAllParams{
			Code:   apiKey,
			X:      x,
			Y:      y,
			Radius: *radius,
			ProdCD: prod,
			Sort:   *sortBy,
		})
		if err != nil {
			fatalErr(err)
		}

		if len(brandSet) > 0 {
			filtered := items[:0]
			for _, it := range items {
				if brandSet[strings.ToUpper(strings.TrimSpace(it.Brand))] {
					filtered = append(filtered, it)
				}
			}
			items = filtered
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

		blocks = append(blocks, outBlock{ProdCD: prod, Stations: items})
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(blocks)
		return
	}

	fmt.Printf("기준 좌표(WGS84): %.6f, %.6f\n", wgsLat, wgsLon)
	fmt.Printf("검색 반경: %dm (오피넷 aroundAll.do 최대 5000m)\n\n", *radius)

	printedAny := false
	for _, b := range blocks {
		if len(prodList) > 1 {
			fmt.Printf("[%s]\n\n", b.ProdCD)
		}
		if len(b.Stations) == 0 {
			fmt.Println("결과가 없습니다.")
			fmt.Println()
			continue
		}
		printedAny = true
		avg := avgByProd[b.ProdCD]

		for i, it := range b.Stations {
			fmt.Printf("%d) %s\n", i+1, it.Name)
			fmt.Printf("   가격: %d원\n", it.Price)
			fmt.Printf("   거리: %.0fm\n", it.DistanceM)
			if *withAvg && avg > 0 {
				diff := float64(it.Price) - avg
				fmt.Printf("   전국평균(%.2f) 대비: %+0.0f원\n", avg, diff)
			}
			if it.Brand != "" {
				fmt.Printf("   브랜드: %s\n", it.Brand)
			}

			// 오피넷 응답 좌표는 KATEC이므로 역변환해서 링크 생성
			lat2, lon2 := geo.KATECToWGS84(it.X, it.Y)
			if lat2 != 0 && lon2 != 0 {
				fmt.Printf("   지도(네이버): %s\n", geo.NaverMapLink(it.Name, lat2, lon2))
			}
			if it.ID != "" {
				fmt.Printf("   ID: %s\n", it.ID)
			}
			fmt.Println()
		}

		if *detail && len(b.Stations) > 0 {
			d, err := opinet.DetailByID(ctx, hc, apiKey, b.Stations[0].ID)
			if err == nil {
				fmt.Println("상세정보")
				fmt.Printf("- 상호: %s (%s)\n", d.Name, d.Brand)
				if d.Address != "" {
					fmt.Printf("- 주소: %s\n", d.Address)
				}
				if d.Tel != "" {
					fmt.Printf("- 전화: %s\n", d.Tel)
				}
				fmt.Printf("- 부가서비스: 경정비=%v, 세차=%v, 편의점=%v, 품질인증=%v\n", d.Services.Maint, d.Services.CarWash, d.Services.CVS, d.Services.Quality)
				if len(d.Prices) > 0 {
					fmt.Println("- 유종별 가격")
					for _, p := range d.Prices {
						fmt.Printf("  - %s: %d원 (%s %s)\n", p.ProdCD, p.Price, p.TradeDT, p.TradeTM)
					}
				}
				fmt.Println()
			}
		}
	}

	if !printedAny {
		fmt.Println("결과가 없습니다.")
		return
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

func parseProdList(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{"B027"}
	}
	if strings.EqualFold(s, "all") {
		// user request focus: gasoline/diesel/LPG
		return []string{"B027", "D047", "K015"}
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, p := range parts {
		p = strings.ToUpper(strings.TrimSpace(p))
		if p == "" {
			continue
		}
		if seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	if len(out) == 0 {
		return []string{"B027"}
	}
	return out
}

func parseCSVSet(s string) map[string]bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	out := map[string]bool{}
	for _, p := range strings.Split(s, ",") {
		p = strings.ToUpper(strings.TrimSpace(p))
		if p == "" {
			continue
		}
		out[p] = true
	}
	return out
}

func printCodes() {
	fmt.Println("유종(prodcd) 코드")
	fmt.Println("- B027: 휘발유")
	fmt.Println("- D047: 경유(자동차용)")
	fmt.Println("- K015: LPG(자동차용부탄)")
	fmt.Println()

	fmt.Println("브랜드(상표) 코드 (brand)")
	fmt.Println("- SKE: SK에너지")
	fmt.Println("- GSC: GS칼텍스")
	fmt.Println("- HDO: 현대오일뱅크")
	fmt.Println("- SOL: S-OIL")
	fmt.Println("- RTE: 자영알뜰")
	fmt.Println("- RTX: 고속도로알뜰")
	fmt.Println("- NHO: 농협알뜰")
	fmt.Println("- ETC: 자가상표")
	fmt.Println("- E1G: E1")
	fmt.Println("- SKG: SK가스")
}
