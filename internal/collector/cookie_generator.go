package collector

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// CookieGenerator Cookie生成器
type CookieGenerator struct {
	rand *rand.Rand
}

// NewCookieGenerator 创建新的Cookie生成器
func NewCookieGenerator() *CookieGenerator {
	return &CookieGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateCookie 生成随机的Cookie
func (g *CookieGenerator) GenerateCookie() string {
	cookies := []string{
		g.generateQgqpBId(),
		g.generateStNvi(),
		g.generateNid(),
		g.generateNidCreateTime(),
		g.generateGvi(),
		g.generateGviCreateTime(),
		g.generateStSi(),
		"fullscreengg=1",
		"fullscreengg2=1",
		g.generateWebsitepoptgApiTime(),
		"st_asi=delete",
		"wsc_checkuser_ok=1",
		g.generateStPvi(),
		g.generateStSp(),
		g.generateStInirUrl(),
		g.generateStSn(),
		g.generateStPsi(),
		"wsc_checkuser_ok=1", // 重复的cookie，保持原有格式
	}

	return strings.Join(cookies, "; ")
}

// generateQgqpBId 生成qgqp_b_id
func (g *CookieGenerator) generateQgqpBId() string {
	return fmt.Sprintf("qgqp_b_id=%s", g.generateRandomHex(32))
}

// generateStNvi 生成st_nvi
func (g *CookieGenerator) generateStNvi() string {
	return fmt.Sprintf("st_nvi=%s", g.generateRandomAlphaNum(25))
}

// generateNid 生成nid
func (g *CookieGenerator) generateNid() string {
	return fmt.Sprintf("nid=%s", g.generateRandomHex(32))
}

// generateNidCreateTime 生成nid_create_time
func (g *CookieGenerator) generateNidCreateTime() string {
	// 生成最近30天内的时间戳（毫秒）
	now := time.Now()
	past := now.AddDate(0, 0, -30)
	timestamp := past.Unix() + g.rand.Int63n(now.Unix()-past.Unix())
	return fmt.Sprintf("nid_create_time=%d", timestamp*1000+int64(g.rand.Intn(1000)))
}

// generateGvi 生成gvi
func (g *CookieGenerator) generateGvi() string {
	return fmt.Sprintf("gvi=%s", g.generateRandomAlphaNum(26))
}

// generateGviCreateTime 生成gvi_create_time
func (g *CookieGenerator) generateGviCreateTime() string {
	// 生成最近30天内的时间戳（毫秒）
	now := time.Now()
	past := now.AddDate(0, 0, -30)
	timestamp := past.Unix() + g.rand.Int63n(now.Unix()-past.Unix())
	return fmt.Sprintf("gvi_create_time=%d", timestamp*1000+int64(g.rand.Intn(1000)))
}

// generateStSi 生成st_si
func (g *CookieGenerator) generateStSi() string {
	return fmt.Sprintf("st_si=%d", g.rand.Int63n(99999999999999)+10000000000000)
}

// generateWebsitepoptgApiTime 生成websitepoptg_api_time
func (g *CookieGenerator) generateWebsitepoptgApiTime() string {
	// 生成最近7天内的时间戳（毫秒）
	now := time.Now()
	past := now.AddDate(0, 0, -7)
	timestamp := past.Unix() + g.rand.Int63n(now.Unix()-past.Unix())
	return fmt.Sprintf("websitepoptg_api_time=%d", timestamp*1000+int64(g.rand.Intn(1000)))
}

// generateStPvi 生成st_pvi
func (g *CookieGenerator) generateStPvi() string {
	return fmt.Sprintf("st_pvi=%d", g.rand.Int63n(99999999999999)+10000000000000)
}

// generateStSp 生成st_sp
func (g *CookieGenerator) generateStSp() string {
	// 生成最近7天内的日期时间
	now := time.Now()
	days := g.rand.Intn(7)
	date := now.AddDate(0, 0, -days)

	// 随机生成时分秒
	hour := g.rand.Intn(24)
	minute := g.rand.Intn(60)
	second := g.rand.Intn(60)

	dateTime := fmt.Sprintf("%04d-%02d-%02d%%20%02d%%3A%02d%%3A%02d",
		date.Year(), date.Month(), date.Day(), hour, minute, second)

	return fmt.Sprintf("st_sp=%s", dateTime)
}

// generateStInirUrl 生成st_inirUrl
func (g *CookieGenerator) generateStInirUrl() string {
	urls := []string{
		"https%3A%2F%2Fdata.eastmoney.com%2Fgphg%2F",
		"https%3A%2F%2Fdata.eastmoney.com%2Fxjllb%2F",
		"https%3A%2F%2Fdata.eastmoney.com%2Fhsgtcg%2F",
		"https%3A%2F%2Fdata.eastmoney.com%2Fbbsj%2F",
		"https%3A%2F%2Fdata.eastmoney.com%2Fzjlx%2F",
	}

	url := urls[g.rand.Intn(len(urls))]
	return fmt.Sprintf("st_inirUrl=%s", url)
}

// generateStSn 生成st_sn
func (g *CookieGenerator) generateStSn() string {
	return fmt.Sprintf("st_sn=%d", g.rand.Intn(500)+50)
}

// generateStPsi 生成st_psi
func (g *CookieGenerator) generateStPsi() string {
	// 生成今天的日期时间戳格式
	now := time.Now()

	// 随机生成今天的时分秒
	hour := g.rand.Intn(24)
	minute := g.rand.Intn(60)
	second := g.rand.Intn(60)
	millisecond := g.rand.Intn(1000)

	dateStr := fmt.Sprintf("%04d%02d%02d%02d%02d%02d%03d",
		now.Year(), now.Month(), now.Day(), hour, minute, second, millisecond)

	// 生成随机的后缀
	suffix := g.rand.Int63n(9999999999999) + 1000000000000

	// 生成随机的最后部分
	lastPart := g.rand.Int63n(9999999999) + 1000000000

	return fmt.Sprintf("st_psi=%s-%d-%d", dateStr, suffix, lastPart)
}

// generateRandomHex 生成指定长度的随机十六进制字符串
func (g *CookieGenerator) generateRandomHex(length int) string {
	const hexChars = "0123456789abcdef"
	result := make([]byte, length)
	for i := range result {
		result[i] = hexChars[g.rand.Intn(len(hexChars))]
	}
	return string(result)
}

// generateRandomAlphaNum 生成指定长度的随机字母数字字符串
func (g *CookieGenerator) generateRandomAlphaNum(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[g.rand.Intn(len(chars))]
	}
	return string(result)
}
