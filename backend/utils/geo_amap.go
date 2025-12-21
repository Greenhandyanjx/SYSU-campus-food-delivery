// utils/geo_amap.go
package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var amapKey = "e3064e9e20ff62d8ebb59d24d634c179" // 高德地图 Web API Key

type GeoCodeResp struct {
	Status   string `json:"status"`
	Info     string `json:"info"`
	Geocodes []struct {
		Location string `json:"location"` // "lng,lat"
	} `json:"geocodes"`
}

func GeoCode(address string) (lng, lat float64, err error) {
	api := "https://restapi.amap.com/v3/geocode/geo"
	q := url.Values{}
	q.Set("key", amapKey)
	q.Set("address", address)
	q.Set("output", "json")

	u := fmt.Sprintf("%s?%s", api, q.Encode())

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(u)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	var r GeoCodeResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return 0, 0, err
	}

	// 详细的错误日志
	fmt.Printf("Geocode response: status=%s, info=%s, count=%d\n", r.Status, r.Info, len(r.Geocodes))

	if r.Status != "1" || len(r.Geocodes) == 0 {
		return 0, 0, fmt.Errorf("geocode failed: %s", r.Info)
	}

	fmt.Sscanf(r.Geocodes[0].Location, "%f,%f", &lng, &lat)
	// 输出高精度经纬度
	fmt.Printf("高德地图API返回: lng=%.8f, lat=%.8f\n", lng, lat)
	return
}
