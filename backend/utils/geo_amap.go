// utils/geo_amap.go
package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var amapKey = "e3064e9e20ff62d8ebb59d24d634c179" // 高德地图 Web API Key（与前端保持一致）

// 备用密钥
var amapBackupKey = "4ba8ba0b6cc65d2f3258e44bb196a8c5"

type GeoCodeResp struct {
	Status   string `json:"status"`
	Info     string `json:"info"`
	Geocodes []struct {
		Location string `json:"location"` // "lng,lat"
	} `json:"geocodes"`
}

func GeoCode(address string) (lng, lat float64, err error) {
	// 先尝试主密钥
	lng, lat, err = geocodeWithKey(address, amapKey, "主密钥")
	if err != nil {
		fmt.Printf("⚠️ 主密钥解析失败: %v，尝试备用密钥\n", err)
		// 尝试备用密钥
		lng, lat, err = geocodeWithKey(address, amapBackupKey, "备用密钥")
	}
	return
}

// 使用指定密钥进行地理编码
func geocodeWithKey(address string, apiKey, keyName string) (lng, lat float64, err error) {
	api := "https://restapi.amap.com/v3/geocode/geo"
	q := url.Values{}
	q.Set("key", apiKey)
	q.Set("address", address)
	q.Set("output", "json")

	u := fmt.Sprintf("%s?%s", api, q.Encode())

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(u)
	if err != nil {
		fmt.Printf("❌ [%s] HTTP请求失败: %v\n", keyName, err)
		return 0, 0, err
	}
	defer resp.Body.Close()

	var r GeoCodeResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		fmt.Printf("❌ [%s] JSON解析失败: %v\n", keyName, err)
		return 0, 0, err
	}

	// 详细的错误日志
	fmt.Printf("📍 [%s] Geocode响应: status=%s, info=%s, count=%d\n", keyName, r.Status, r.Info, len(r.Geocodes))

	if r.Status != "1" || len(r.Geocodes) == 0 {
		fmt.Printf("❌ [%s] 解析失败: %s\n", keyName, r.Info)
		return 0, 0, fmt.Errorf("geocode failed: %s", r.Info)
	}

	fmt.Sscanf(r.Geocodes[0].Location, "%f,%f", &lng, &lat)
	// 输出高精度经纬度
	fmt.Printf("✅ [%s] 解析成功: lng=%.8f, lat=%.8f\n", keyName, lng, lat)
	return
}
