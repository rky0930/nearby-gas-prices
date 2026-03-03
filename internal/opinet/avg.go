package opinet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// AvgAllPrice returns the current nationwide average price per product.
//
// Official docs mention certkey; aroundAll works with code. To be resilient,
// we send both code and certkey with the same key.
func AvgAllPrice(ctx context.Context, hc *http.Client, key string) (map[string]float64, error) {
	if key == "" {
		return nil, errors.New("opinet: missing key")
	}

	endpoint := "https://www.opinet.co.kr/api/avgAllPrice.do"
	q := url.Values{}
	q.Set("out", "json")
	q.Set("code", key)
	q.Set("certkey", key)

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

	var raw struct {
		Result struct {
			Oil []struct {
				ProdCD string         `json:"PRODCD"`
				Price  stringOrNumber `json:"PRICE"`
			} `json:"OIL"`
		} `json:"RESULT"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	out := make(map[string]float64, len(raw.Result.Oil))
	for _, it := range raw.Result.Oil {
		p, err := strconv.ParseFloat(string(it.Price), 64)
		if err != nil {
			continue
		}
		out[it.ProdCD] = p
	}
	return out, nil
}
