// utils/geo_amap.go
package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var amapKey = "" // 从 env 读，例如 AMAP_WEB_KEY

type GeoCodeResp struct {
	Status   string `json:"status"`
	Geocodes []struct {
		Location string `json:"location"` // "lng,lat"
	} `json:"geocodes"`
}

func GeoCode(address string) (lng, lat float64, err error) {
	api := "https://restapi.amap.com/v3/geocode/geo"
	q := url.Values{}
	q.Set("key", amapKey)
	q.Set("address", address)

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
	if r.Status != "1" || len(r.Geocodes) == 0 {
		return 0, 0, fmt.Errorf("geocode failed")
	}

	fmt.Sscanf(r.Geocodes[0].Location, "%f,%f", &lng, &lat)
	return
}
