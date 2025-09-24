package collector

import (
	"context"
	"testing"
	"time"

	"stock/internal/logger"
)

// TestEastMoneyCollector_RateLimit 测试限流功能
func TestEastMoneyCollector_RateLimit(t *testing.T) {
	// 创建测试用的 logger
	logger := logger.NewLogger(logger.LogConfig{
		Level:  "debug",
		Format: "text",
	})

	// 创建东财采集器，设置较低的限流速率便于测试
	collector := newEastMoneyCollector(logger)
	collector.SetRateLimit(2) // 每秒2个请求

	// 测试连续请求的时间间隔
	start := time.Now()

	// 发送3个请求
	for i := 0; i < 3; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := collector.makeRequestWithContext(ctx, "https://httpbin.org/delay/0", "")
		cancel()

		if err != nil {
			t.Logf("Request %d failed (expected for rate limiting): %v", i+1, err)
		} else {
			t.Logf("Request %d completed at %v", i+1, time.Since(start))
		}
	}

	elapsed := time.Since(start)
	t.Logf("Total time for 3 requests: %v", elapsed)

	// 验证限流是否生效（3个请求在2 RPS的限制下应该至少需要1秒）
	if elapsed < 1*time.Second {
		t.Errorf("Rate limiting may not be working properly. Expected at least 1s, got %v", elapsed)
	}
}

// TestEastMoneyCollector_SetRateLimit 测试动态设置限流速率
func TestEastMoneyCollector_SetRateLimit(t *testing.T) {
	logger := logger.NewLogger(logger.LogConfig{
		Level:  "debug",
		Format: "text",
	})

	collector := newEastMoneyCollector(logger)

	// 测试初始限流速率
	initialRate := collector.GetRateLimit()
	if initialRate != 10 {
		t.Errorf("Expected initial rate limit to be 10, got %d", initialRate)
	}

	// 测试设置新的限流速率
	newRate := 5
	collector.SetRateLimit(newRate)

	if collector.GetRateLimit() != newRate {
		t.Errorf("Expected rate limit to be %d, got %d", newRate, collector.GetRateLimit())
	}

	// 测试设置无效值（<=0）
	collector.SetRateLimit(0)
	if collector.GetRateLimit() != 1 {
		t.Errorf("Expected rate limit to be 1 when set to 0, got %d", collector.GetRateLimit())
	}

	collector.SetRateLimit(-5)
	if collector.GetRateLimit() != 1 {
		t.Errorf("Expected rate limit to be 1 when set to -5, got %d", collector.GetRateLimit())
	}
}

// TestEastMoneyCollector_GetRateLimitStats 测试获取限流统计信息
func TestEastMoneyCollector_GetRateLimitStats(t *testing.T) {
	logger := logger.NewLogger(logger.LogConfig{
		Level:  "debug",
		Format: "text",
	})

	collector := newEastMoneyCollector(logger)
	collector.SetRateLimit(5)

	stats := collector.GetRateLimitStats()

	// 验证统计信息包含必要的字段
	expectedFields := []string{"rate_limit", "current_limit", "burst_size", "tokens"}
	for _, field := range expectedFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("Expected stats to contain field '%s'", field)
		}
	}

	// 验证具体值
	if stats["rate_limit"] != 5 {
		t.Errorf("Expected rate_limit to be 5, got %v", stats["rate_limit"])
	}

	if stats["current_limit"] != float64(5) {
		t.Errorf("Expected current_limit to be 5.0, got %v", stats["current_limit"])
	}

	if stats["burst_size"] != 10 { // 应该是 rate_limit * 2
		t.Errorf("Expected burst_size to be 10, got %v", stats["burst_size"])
	}

	t.Logf("Rate limit stats: %+v", stats)
}

// BenchmarkEastMoneyCollector_RateLimit 基准测试限流性能
func BenchmarkEastMoneyCollector_RateLimit(b *testing.B) {
	logger := logger.NewLogger(logger.LogConfig{
		Level:  "error", // 减少日志输出
		Format: "text",
	})

	collector := newEastMoneyCollector(logger)
	collector.SetRateLimit(100) // 设置较高的限流速率

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			// 使用一个快速响应的测试URL
			collector.makeRequestWithContext(ctx, "https://httpbin.org/status/200", "")
			cancel()
		}
	})
}
