package nominatim

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type Result struct {
	DisplayName string
	Lat         float64
	Lon         float64
}

type rawItem struct {
	DisplayName string `json:"display_name"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
}

func SearchOne(ctx context.Context, hc *http.Client, query, userAgent string) (Result, error) {
	u := "https://nominatim.openstreetmap.org/search"
	q := url.Values{}
	q.Set("q", query)
	q.Set("format", "jsonv2")
	q.Set("limit", "1")
	urlStr := u + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := hc.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return Result{}, fmt.Errorf("nominatim: status %s", resp.Status)
	}

	var items []rawItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return Result{}, err
	}
	if len(items) == 0 {
		return Result{}, errors.New("nominatim: no results")
	}

	lat, err := strconv.ParseFloat(items[0].Lat, 64)
	if err != nil {
		return Result{}, err
	}
	lon, err := strconv.ParseFloat(items[0].Lon, 64)
	if err != nil {
		return Result{}, err
	}

	return Result{DisplayName: items[0].DisplayName, Lat: lat, Lon: lon}, nil
}
