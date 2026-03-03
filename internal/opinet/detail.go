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

type StationDetail struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Brand     string  `json:"brand"`
	Address   string  `json:"address"` // prefer NEW_ADR if present
	JibunAddr string  `json:"jibun_address"`
	RoadAddr  string  `json:"road_address"`
	Tel       string  `json:"tel"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Services  struct {
		Maint   bool   `json:"maint"`
		CarWash bool   `json:"car_wash"`
		Quality bool   `json:"quality_cert"`
		CVS     bool   `json:"cvs"`
		LPGYN   string `json:"lpg_yn"` // N/Y/C
	} `json:"services"`
	Prices []struct {
		ProdCD  string `json:"prodcd"`
		Price   int    `json:"price"`
		TradeDT string `json:"trade_dt"`
		TradeTM string `json:"trade_tm"`
	} `json:"prices"`
}

func DetailByID(ctx context.Context, hc *http.Client, key string, id string) (*StationDetail, error) {
	if key == "" {
		return nil, errors.New("opinet: missing key")
	}
	if id == "" {
		return nil, errors.New("opinet: missing id")
	}

	endpoint := "https://www.opinet.co.kr/api/detailById.do"
	q := url.Values{}
	q.Set("out", "json")
	q.Set("id", id)
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
				ID    string         `json:"UNI_ID"`
				Brand string         `json:"POLL_DIV_CD"`
				Name  string         `json:"OS_NM"`
				VAddr string         `json:"VAN_ADR"`
				NAddr string         `json:"NEW_ADR"`
				Tel   string         `json:"TEL"`
				X     stringOrNumber `json:"GIS_X_COOR"`
				Y     stringOrNumber `json:"GIS_Y_COOR"`
				LPGYN string         `json:"LPG_YN"`
				Maint string         `json:"MAINT_YN"`
				Wash  string         `json:"CAR_WASH_YN"`
				Qual  string         `json:"KPETRO_YN"`
				CVS   string         `json:"CVS_YN"`
				OilP  []struct {
					ProdCD  string         `json:"PRODCD"`
					Price   stringOrNumber `json:"PRICE"`
					TradeDT string         `json:"TRADE_DT"`
					TradeTM string         `json:"TRADE_TM"`
				} `json:"OIL_PRICE"`
			} `json:"OIL"`
		} `json:"RESULT"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if len(raw.Result.Oil) == 0 {
		return nil, fmt.Errorf("opinet: no station for id %s", id)
	}

	src := raw.Result.Oil[0]
	out := &StationDetail{}
	out.ID = src.ID
	out.Name = src.Name
	out.Brand = src.Brand
	out.JibunAddr = src.VAddr
	out.RoadAddr = src.NAddr
	out.Tel = src.Tel
	out.Services.LPGYN = src.LPGYN
	out.Services.Maint = src.Maint == "Y"
	out.Services.CarWash = src.Wash == "Y"
	out.Services.Quality = src.Qual == "Y"
	out.Services.CVS = src.CVS == "Y"

	// choose primary address
	out.Address = out.RoadAddr
	if out.Address == "" {
		out.Address = out.JibunAddr
	}

	out.X, _ = strconv.ParseFloat(string(src.X), 64)
	out.Y, _ = strconv.ParseFloat(string(src.Y), 64)

	for _, p := range src.OilP {
		pi, _ := strconv.Atoi(string(p.Price))
		out.Prices = append(out.Prices, struct {
			ProdCD  string `json:"prodcd"`
			Price   int    `json:"price"`
			TradeDT string `json:"trade_dt"`
			TradeTM string `json:"trade_tm"`
		}{ProdCD: p.ProdCD, Price: pi, TradeDT: p.TradeDT, TradeTM: p.TradeTM})
	}

	return out, nil
}
