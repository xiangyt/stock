package collector

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// UserAgentGenerator 用户代理生成器
type UserAgentGenerator struct {
	rand *rand.Rand
}

// NewUserAgentGenerator 创建新的用户代理生成器
func NewUserAgentGenerator() *UserAgentGenerator {
	return &UserAgentGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateUserAgent 生成随机的User-Agent
func (g *UserAgentGenerator) GenerateUserAgent() string {
	browsers := []string{
		"Chrome",
		"Firefox",
		"Safari",
		"Edge",
	}

	chromeVersions := []string{
		"120.0.0.0", "121.0.0.0", "122.0.0.0", "123.0.0.0", "124.0.0.0",
		"125.0.0.0", "126.0.0.0", "127.0.0.0", "128.0.0.0", "129.0.0.0",
		"130.0.0.0", "131.0.0.0", "132.0.0.0", "133.0.0.0", "134.0.0.0",
		"135.0.0.0", "136.0.0.0", "137.0.0.0", "138.0.0.0", "139.0.0.0",
	}

	firefoxVersions := []string{
		"118.0", "119.0", "120.0", "121.0", "122.0", "123.0", "124.0", "125.0",
		"126.0", "127.0", "128.0", "129.0", "130.0", "131.0", "132.0",
	}

	safariVersions := []string{
		"537.36", "605.1.15", "604.1.38", "605.1.15", "537.36",
	}

	operatingSystems := []struct {
		name    string
		version string
	}{
		{"Windows NT 10.0; Win64; x64", "10.0"},
		{"Windows NT 11.0; Win64; x64", "11.0"},
		{"Macintosh; Intel Mac OS X 10_15_7", "10.15.7"},
		{"Macintosh; Intel Mac OS X 11_7_10", "11.7.10"},
		{"Macintosh; Intel Mac OS X 12_7_4", "12.7.4"},
		{"Macintosh; Intel Mac OS X 13_6_6", "13.6.6"},
		{"Macintosh; Intel Mac OS X 14_4_1", "14.4.1"},
		{"X11; Linux x86_64", ""},
		{"X11; Ubuntu; Linux x86_64", ""},
	}

	browser := browsers[g.rand.Intn(len(browsers))]
	os := operatingSystems[g.rand.Intn(len(operatingSystems))]

	switch browser {
	case "Chrome":
		version := chromeVersions[g.rand.Intn(len(chromeVersions))]
		webkitVersion := safariVersions[g.rand.Intn(len(safariVersions))]
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/%s (KHTML, like Gecko) Chrome/%s Safari/%s",
			os.name, webkitVersion, version, webkitVersion)

	case "Firefox":
		version := firefoxVersions[g.rand.Intn(len(firefoxVersions))]
		return fmt.Sprintf("Mozilla/5.0 (%s; rv:%s) Gecko/20100101 Firefox/%s",
			os.name, version, version)

	case "Safari":
		if !strings.Contains(os.name, "Macintosh") {
			// Safari主要在macOS上，如果不是Mac系统，改用Chrome
			version := chromeVersions[g.rand.Intn(len(chromeVersions))]
			webkitVersion := safariVersions[g.rand.Intn(len(safariVersions))]
			return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/%s (KHTML, like Gecko) Chrome/%s Safari/%s",
				os.name, webkitVersion, version, webkitVersion)
		}
		webkitVersion := safariVersions[g.rand.Intn(len(safariVersions))]
		safariVersion := fmt.Sprintf("17.%d", g.rand.Intn(5)+1)
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/%s (KHTML, like Gecko) Version/%s Safari/%s",
			os.name, webkitVersion, safariVersion, webkitVersion)

	case "Edge":
		version := chromeVersions[g.rand.Intn(len(chromeVersions))]
		webkitVersion := safariVersions[g.rand.Intn(len(safariVersions))]
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/%s (KHTML, like Gecko) Chrome/%s Safari/%s Edg/%s",
			os.name, webkitVersion, version, webkitVersion, version)

	default:
		// 默认返回Chrome
		version := chromeVersions[g.rand.Intn(len(chromeVersions))]
		webkitVersion := safariVersions[g.rand.Intn(len(safariVersions))]
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/%s (KHTML, like Gecko) Chrome/%s Safari/%s",
			os.name, webkitVersion, version, webkitVersion)
	}
}

// GenerateSecChUa 生成对应的sec-ch-ua头
func (g *UserAgentGenerator) GenerateSecChUa(userAgent string) string {
	if strings.Contains(userAgent, "Chrome") && !strings.Contains(userAgent, "Edg") {
		// Chrome浏览器
		chromeVersion := g.extractChromeVersion(userAgent)
		return fmt.Sprintf(`"Not;A=Brand";v="99", "Google Chrome";v="%s", "Chromium";v="%s"`,
			chromeVersion, chromeVersion)
	} else if strings.Contains(userAgent, "Firefox") {
		// Firefox浏览器
		return `"Not;A=Brand";v="99", "Firefox";v="130"`
	} else if strings.Contains(userAgent, "Safari") && !strings.Contains(userAgent, "Chrome") {
		// Safari浏览器
		return `"Not;A=Brand";v="99", "Safari";v="17"`
	} else if strings.Contains(userAgent, "Edg") {
		// Edge浏览器
		edgeVersion := g.extractChromeVersion(userAgent)
		return fmt.Sprintf(`"Not;A=Brand";v="99", "Microsoft Edge";v="%s", "Chromium";v="%s"`,
			edgeVersion, edgeVersion)
	}

	// 默认返回Chrome格式
	return `"Not;A=Brand";v="99", "Google Chrome";v="139", "Chromium";v="139"`
}

// extractChromeVersion 从User-Agent中提取Chrome版本号
func (g *UserAgentGenerator) extractChromeVersion(userAgent string) string {
	// 匹配Chrome/版本号
	if idx := strings.Index(userAgent, "Chrome/"); idx != -1 {
		start := idx + 7 // "Chrome/"的长度
		end := start
		for end < len(userAgent) && userAgent[end] != ' ' && userAgent[end] != '.' {
			end++
		}
		if end > start {
			// 只返回主版本号
			version := userAgent[start:end]
			if dotIdx := strings.Index(version, "."); dotIdx != -1 {
				return version[:dotIdx]
			}
			return version
		}
	}
	return "139" // 默认版本
}
