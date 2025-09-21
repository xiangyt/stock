package collector

import (
	"strings"
	"testing"
	"time"

	"stock/internal/config"
	"stock/internal/logger"
)

func TestUserAgentGenerator(t *testing.T) {
	gen := NewUserAgentGenerator()

	// 测试生成多个不同的User-Agent
	userAgents := make(map[string]bool)
	for i := 0; i < 10; i++ {
		ua := gen.GenerateUserAgent()
		if ua == "" {
			t.Error("Generated empty User-Agent")
		}

		// 检查基本格式
		if !strings.Contains(ua, "Mozilla") {
			t.Errorf("User-Agent should contain 'Mozilla': %s", ua)
		}

		userAgents[ua] = true
		t.Logf("Generated UA %d: %s", i+1, ua)
	}

	// 检查是否生成了不同的User-Agent
	if len(userAgents) < 3 {
		t.Errorf("Expected at least 3 different User-Agents, got %d", len(userAgents))
	}
}

func TestCookieGenerator(t *testing.T) {
	gen := NewCookieGenerator()

	// 测试生成多个不同的Cookie
	cookies := make(map[string]bool)
	for i := 0; i < 5; i++ {
		cookie := gen.GenerateCookie()
		if cookie == "" {
			t.Error("Generated empty Cookie")
		}

		// 检查必要的cookie字段
		requiredFields := []string{
			"qgqp_b_id=",
			"st_nvi=",
			"nid=",
			"wsc_checkuser_ok=1",
		}

		for _, field := range requiredFields {
			if !strings.Contains(cookie, field) {
				t.Errorf("Cookie should contain '%s': %s", field, cookie)
			}
		}

		cookies[cookie] = true
		t.Logf("Generated Cookie %d length: %d", i+1, len(cookie))
	}

	// 检查是否生成了不同的Cookie
	if len(cookies) < 3 {
		t.Errorf("Expected at least 3 different Cookies, got %d", len(cookies))
	}
}

func TestSecChUaGeneration(t *testing.T) {
	gen := NewUserAgentGenerator()

	testCases := []struct {
		userAgent string
		expected  string
	}{
		{
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			expected:  "Chrome",
		},
		{
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
			expected:  "Safari",
		},
		{
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
			expected:  "Firefox",
		},
	}

	for _, tc := range testCases {
		secChUa := gen.GenerateSecChUa(tc.userAgent)
		if !strings.Contains(secChUa, tc.expected) {
			t.Errorf("sec-ch-ua should contain '%s' for UA: %s, got: %s", tc.expected, tc.userAgent, secChUa)
		}
		t.Logf("UA: %s\nsec-ch-ua: %s\n", tc.userAgent, secChUa)
	}
}

func TestEastMoneyCollectorRandomHeaders(t *testing.T) {
	// 创建测试logger
	testLogger := logger.NewLogger(config.LogConfig{
		Level:  "debug",
		Format: "text",
	})

	// 创建采集器
	collector := newEastMoneyCollector(testLogger)

	// 测试初始User-Agent和Cookie
	initialUA := collector.GetCurrentUserAgent()
	initialCookie := collector.GetCurrentCookie()

	if initialUA == "" {
		t.Error("Initial User-Agent should not be empty")
	}
	if initialCookie == "" {
		t.Error("Initial Cookie should not be empty")
	}

	t.Logf("Initial UA: %s", initialUA)
	t.Logf("Initial Cookie length: %d", len(initialCookie))

	// 测试强制更新
	collector.ForceUpdateUserAgentAndCookie()

	newUA := collector.GetCurrentUserAgent()
	newCookie := collector.GetCurrentCookie()

	// 检查是否更新了（大概率会不同）
	if newUA == initialUA && newCookie == initialCookie {
		t.Log("Warning: New headers are same as initial (low probability but possible)")
	}

	t.Logf("New UA: %s", newUA)
	t.Logf("New Cookie length: %d", len(newCookie))

	// 测试多次更新生成不同的值
	uniqueUAs := make(map[string]bool)
	uniqueCookies := make(map[string]bool)

	for i := 0; i < 5; i++ {
		collector.ForceUpdateUserAgentAndCookie()
		uniqueUAs[collector.GetCurrentUserAgent()] = true
		uniqueCookies[collector.GetCurrentCookie()] = true
	}

	if len(uniqueUAs) < 2 {
		t.Log("Warning: Generated less than 2 unique User-Agents (low probability but possible)")
	}
	if len(uniqueCookies) < 2 {
		t.Log("Warning: Generated less than 2 unique Cookies (low probability but possible)")
	}

	t.Logf("Generated %d unique User-Agents and %d unique Cookies", len(uniqueUAs), len(uniqueCookies))
}

func TestRandomHeadersIntegration(t *testing.T) {
	// 创建测试logger
	testLogger := logger.NewLogger(config.LogConfig{
		Level:  "info",
		Format: "text",
	})

	// 创建采集器
	collector := newEastMoneyCollector(testLogger)

	// 模拟时间流逝，测试自动更新
	collector.lastUpdateTime = time.Now().Add(-15 * time.Minute) // 15分钟前

	initialUA := collector.GetCurrentUserAgent()

	// 这里我们不能真正发送请求，但可以测试更新逻辑
	// 在实际的makeRequestWithContext中会检查时间并自动更新

	// 手动触发更新检查
	if time.Since(collector.lastUpdateTime) > 10*time.Minute {
		collector.updateUserAgentAndCookie()
	}

	newUA := collector.GetCurrentUserAgent()

	// 验证更新时间被重置
	if time.Since(collector.lastUpdateTime) > time.Minute {
		t.Error("Last update time should be recent after update")
	}

	t.Logf("Auto-update test - Initial: %s, New: %s", initialUA[:50]+"...", newUA[:50]+"...")
}
