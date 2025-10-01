package collector

import (
	"testing"
)

func TestGenerateWencaiToken(t *testing.T) {
	// 测试便捷函数
	token := GenerateWencaiToken()

	if token == "" {
		t.Error("GenerateWencaiToken returned empty string")
	}

	t.Logf("Generated token: %s", token)
}
