package opinet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type AroundAllParams struct {
	Code   string
	X      float64
	Y      float64
	Radius int
	ProdCD string
	Sort   int
}

type Station struct {
	ID        string  `json:"id"`
	Brand     string  `json:"brand"`
	Name      string  `json:"name"`
	Price     int     `json:"price"`
	DistanceM float64 `json:"distance_m"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
}

type stringOrNumber string

func (s *stringOrNumber) UnmarshalJSON(b []byte) error {
	// Handle: "123", 123, null
	str := strings.TrimSpace(string(b))
	if str == "null" || str == "" {
		*s = ""
		return nil
	}
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		*s = stringOrNumber(str[1 : len(str)-1])
		return nil
	}
	*s = stringOrNumber(str)
	return nil
}

type aroundAllResp struct {
	Result struct {
		Oil []struct {
			ID    string         `json:"UNI_ID"`
			Brand string         `json:"POLL_DIV_CD"`
			OSNM  string         `json:"OS_NM"`
			Price stringOrNumber `json:"PRICE"`
			Dist  stringOrNumber `json:"DISTANCE"`
			X     stringOrNumber `json:"GIS_X_COOR"`
			Y     stringOrNumber `json:"GIS_Y_COOR"`
		} `json:"OIL"`
	} `json:"RESULT"`
}

func AroundAll(ctx context.Context, hc *http.Client, p AroundAllParams) ([]Station, error) {
	if p.Code == "" {
		return nil, errors.New("opinet: missing code")
	}

	endpoint := "https://www.opinet.co.kr/api/aroundAll.do"
	q := url.Values{}
	q.Set("code", p.Code)
	q.Set("x", fmt.Sprintf("%.3f", p.X))
	q.Set("y", fmt.Sprintf("%.3f", p.Y))
	q.Set("radius", strconv.Itoa(p.Radius))
	q.Set("prodcd", p.ProdCD)
	q.Set("sort", strconv.Itoa(p.Sort))
	q.Set("out", "json")

	urlStr := endpoint + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}

	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("opinet: status %s", resp.Status)
	}

	var raw aroundAllResp
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	out := make([]Station, 0, len(raw.Result.Oil))
	for _, it := range raw.Result.Oil {
		price, _ := strconv.Atoi(string(it.Price))
		dist, _ := strconv.ParseFloat(string(it.Dist), 64)
		x, _ := strconv.ParseFloat(string(it.X), 64)
		y, _ := strconv.ParseFloat(string(it.Y), 64)
		out = append(out, Station{ID: it.ID, Brand: it.Brand, Name: it.OSNM, Price: price, DistanceM: dist, X: x, Y: y})
	}

	return out, nil
}
