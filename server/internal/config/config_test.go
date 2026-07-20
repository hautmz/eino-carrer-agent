package config

import (
	"os"
	"testing"
)

// TestLoad 测试配置加载功能
func TestLoad(t *testing.T) {
	cfg, err := Load("../../configs/config.yaml")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	if cfg.Server.Port != 8081 {
		t.Errorf("服务端口期望 8081，实际 %d", cfg.Server.Port)
	}
	if cfg.Server.Mode != "debug" {
		t.Errorf("运行模式期望 debug，实际 %s", cfg.Server.Mode)
	}
	if cfg.Database.Type != "sqlite" {
		t.Errorf("数据库类型期望 sqlite，实际 %s", cfg.Database.Type)
	}
	if cfg.Agent.ReportTimeout != 360 {
		t.Errorf("报告超时期望 360，实际 %d", cfg.Agent.ReportTimeout)
	}
	if cfg.Agent.MaxConcurrentSections != 4 {
		t.Errorf("最大并行数期望 4，实际 %d", cfg.Agent.MaxConcurrentSections)
	}

	t.Logf("配置加载成功: %+v", cfg)
}

// TestEnvOverride 测试环境变量覆盖配置
func TestEnvOverride(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-api-key")
	os.Setenv("OPENAI_BASE_URL", "https://test.api.com/v1")
	os.Setenv("OPENAI_MODEL", "gpt-4o-test")
	defer func() {
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("OPENAI_BASE_URL")
		os.Unsetenv("OPENAI_MODEL")
	}()

	cfg, err := Load("../../configs/config.yaml")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	if cfg.OpenAI.APIKey != "test-api-key" {
		t.Errorf("OPENAI_API_KEY 期望 test-api-key，实际 %s", cfg.OpenAI.APIKey)
	}
	if cfg.OpenAI.BaseURL != "https://test.api.com/v1" {
		t.Errorf("OPENAI_BASE_URL 期望 https://test.api.com/v1，实际 %s", cfg.OpenAI.BaseURL)
	}
	if cfg.OpenAI.Model != "gpt-4o-test" {
		t.Errorf("OPENAI_MODEL 期望 gpt-4o-test，实际 %s", cfg.OpenAI.Model)
	}
}
