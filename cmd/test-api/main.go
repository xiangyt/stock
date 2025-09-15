package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

func main() {
	// 构建股东户数API URL
	baseURL := "https://datacenter-web.eastmoney.com/api/data/v1/get"
	params := url.Values{}
	params.Set("callback", fmt.Sprintf("jQuery1123014159649525581786_%d", time.Now().UnixMilli()))
	params.Set("sortColumns", "END_DATE")
	params.Set("sortTypes", "-1")
	params.Set("pageSize", "5") // 只获取5条记录用于测试
	params.Set("pageNumber", "1")
	params.Set("reportName", "RPT_HOLDERNUM_DET")
	params.Set("columns", "SECURITY_CODE,SECURITY_NAME_ABBR,CHANGE_SHARES,CHANGE_REASON,END_DATE,INTERVAL_CHRATE,AVG_MARKET_CAP,AVG_HOLD_NUM,TOTAL_MARKET_CAP,TOTAL_A_SHARES,HOLD_NOTICE_DATE,HOLDER_NUM,PRE_HOLDER_NUM,HOLDER_NUM_CHANGE,HOLDER_NUM_RATIO,END_DATE,PRE_END_DATE")
	params.Set("quoteColumns", "f2,f3")
	params.Set("quoteType", "0")
	params.Set("filter", "(SECURITY_CODE=\"001208\")")
	params.Set("source", "WEB")
	params.Set("client", "WEB")

	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	fmt.Printf("请求URL: %s\n\n", requestURL)

	// 发送HTTP请求
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		log.Fatalf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", "https://data.eastmoney.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("HTTP状态码: %d", resp.StatusCode)
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("读取响应失败: %v", err)
	}

	fmt.Printf("原始响应 (前500字符):\n%s\n\n", string(body)[:min(500, len(body))])

	// 解析JSONP响应
	jsonpPattern := regexp.MustCompile(`jQuery\d+_\d+\((.*)\)`)
	matches := jsonpPattern.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		log.Fatalf("无法解析JSONP响应")
	}

	jsonData := matches[1]
	fmt.Printf("提取的JSON数据 (前1000字符):\n%s\n", jsonData[:min(1000, len(jsonData))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
