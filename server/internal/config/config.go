// Package config 提供 Eino Career Agent 的配置加载功能
// 支持从 config.yaml 文件和环境变量加载配置，环境变量优先级高于 yaml 文件
package config

import (
	"fmt"
	"reflect"
	"time"

	"github.com/spf13/viper"
)

// Config 是全局配置结构体，包含所有配置项
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`    // 服务器配置
	Database DatabaseConfig `mapstructure:"database"`  // 数据库配置
	JWT      JWTConfig      `mapstructure:"jwt"`       // JWT 认证配置
	Upload   UploadConfig   `mapstructure:"upload"`    // 文件上传配置
	Agent    AgentConfig    `mapstructure:"agent"`     // Agent 编排配置
	SSE      SSEConfig      `mapstructure:"sse"`       // SSE 流式推送配置
	OpenAI   OpenAIConfig   `mapstructure:"-"`         // OpenAI 兼容 LLM 配置（仅从环境变量读取）
}

// ServerConfig 服务器相关配置
type ServerConfig struct {
	Port int    `mapstructure:"port"` // 服务监听端口
	Mode string `mapstructure:"mode"` // 运行模式: debug / release / test
}

// DatabaseConfig 数据库相关配置
type DatabaseConfig struct {
	Type string `mapstructure:"type"` // 数据库类型，当前仅支持 sqlite
	Path string `mapstructure:"path"` // SQLite 数据库文件路径
}

// JWTConfig JWT 认证相关配置
type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`     // JWT 签名密钥
	Expiration time.Duration `mapstructure:"expiration"` // Token 过期时间
}

// UploadConfig 文件上传相关配置
type UploadConfig struct {
	MaxSize      int64    `mapstructure:"max_size"`       // 文件最大大小（字节）
	AllowedTypes []string `mapstructure:"allowed_types"`  // 允许的文件类型扩展名
	StoragePath  string   `mapstructure:"storage_path"`   // 文件存储路径
}

// AgentConfig Agent 编排相关配置
type AgentConfig struct {
	ReportTimeout         int `mapstructure:"report_timeout"`          // 报告生成超时（秒）
	SectionTimeout        int `mapstructure:"section_timeout"`         // 单章节生成超时（秒）
	MaxConcurrentSections int `mapstructure:"max_concurrent_sections"` // 报告章节最大并行数
	MaxHistoryMessages    int `mapstructure:"max_history_messages"`    // 单次对话加载的最大历史消息数
}

// SSEConfig SSE 流式推送相关配置
type SSEConfig struct {
	HeartbeatInterval int `mapstructure:"heartbeat_interval"` // 心跳间隔（秒）
}

// OpenAIConfig OpenAI 兼容 LLM 配置，仅从环境变量读取
type OpenAIConfig struct {
	APIKey  string // OpenAI API Key，从环境变量 OPENAI_API_KEY 读取
	BaseURL string // OpenAI 兼容 API Base URL，从环境变量 OPENAI_BASE_URL 读取
	Model   string // 使用的模型名称，从环境变量 OPENAI_MODEL 读取
}

// globalConfig 是全局配置单例
var globalConfig *Config

// Load 加载配置文件和环境变量
// configPath 为配置文件路径，默认为 ./configs/config.yaml
// 环境变量 OPENAI_API_KEY、OPENAI_BASE_URL、OPENAI_MODEL 会覆盖配置文件中的 LLM 设置
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 设置配置文件路径和格式
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 设置默认值
	setDefaults(v)

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 手动绑定环境变量（不使用 AutomaticEnv，避免 viper 的环境变量查找干扰值读取）
	v.BindEnv("OPENAI_API_KEY")
	v.BindEnv("OPENAI_BASE_URL")
	v.BindEnv("OPENAI_MODEL")

	// 将配置解析到结构体
	// 使用自定义 DecodeHook 处理 YAML 整数（float64）到 Go int 的类型转换
	cfg := &Config{}
	if err := v.Unmarshal(cfg, viper.DecodeHook(mapstructureDecodeHook)); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 手动补充解析 viper Unmarshal 可能遗漏的整数字段
	// 这是因为 viper 的 Unmarshal 对嵌套结构的 int 解析有时不稳定
	cfg.Agent.ReportTimeout = v.GetInt("agent.report_timeout")
	cfg.Agent.SectionTimeout = v.GetInt("agent.section_timeout")
	cfg.Agent.MaxConcurrentSections = v.GetInt("agent.max_concurrent_sections")
	cfg.Agent.MaxHistoryMessages = v.GetInt("agent.max_history_messages")
	cfg.Server.Port = v.GetInt("server.port")
	cfg.Upload.MaxSize = v.GetInt64("upload.max_size")
	cfg.SSE.HeartbeatInterval = v.GetInt("sse.heartbeat_interval")

	// 从环境变量读取 OpenAI 配置
	cfg.OpenAI = OpenAIConfig{
		APIKey:  v.GetString("OPENAI_API_KEY"),
		BaseURL: v.GetString("OPENAI_BASE_URL"),
		Model:   v.GetString("OPENAI_MODEL"),
	}

	// 验证必要配置
	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	globalConfig = cfg
	return cfg, nil
}

// Get 获取全局配置单例
// 必须在 Load 之后调用，否则返回 nil
func Get() *Config {
	return globalConfig
}

// setDefaults 设置配置默认值
func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 8081)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("database.type", "sqlite")
	v.SetDefault("database.path", "./data/eino_career.db")
	v.SetDefault("jwt.secret", "eino-career-agent-secret-change-in-production")
	v.SetDefault("jwt.expiration", "72h")
	v.SetDefault("upload.max_size", 10485760)
	v.SetDefault("upload.allowed_types", []string{"pdf", "docx"})
	v.SetDefault("upload.storage_path", "./data/uploads")
	v.SetDefault("agent.report_timeout", 360)
	v.SetDefault("agent.section_timeout", 120)
	v.SetDefault("agent.max_concurrent_sections", 4)
	v.SetDefault("agent.max_history_messages", 50)
	v.SetDefault("sse.heartbeat_interval", 15)
}

// mapstructureDecodeHook 是自定义的类型转换钩子
// 用于处理 viper 从 YAML 读取值时的类型转换问题：
// 1. YAML 中的整数被 viper 解析为 float64，需要转为目标 int 类型
// 2. YAML 中的字符串（如 "72h"）需要转为 time.Duration 类型
func mapstructureDecodeHook(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
	// 处理 float64 → int 的类型转换
	if from.Kind() == reflect.Float64 {
		if to.Kind() == reflect.Int {
			floatVal, ok := data.(float64)
			if ok {
				return int(floatVal), nil
			}
		}
		if to.Kind() == reflect.Int64 {
			floatVal, ok := data.(float64)
			if ok {
				return int64(floatVal), nil
			}
		}
	}

	// 处理 string → time.Duration 的类型转换
	if from.Kind() == reflect.String && to == reflect.TypeOf(time.Duration(0)) {
		strVal, ok := data.(string)
		if ok {
			d, err := time.ParseDuration(strVal)
			if err != nil {
				return nil, fmt.Errorf("无法解析 duration 值 '%s': %w", strVal, err)
			}
			return d, nil
		}
	}

	return data, nil
}

// validate 验证配置的合法性
func validate(cfg *Config) error {
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("无效的服务端口: %d", cfg.Server.Port)
	}
	if cfg.Server.Mode != "debug" && cfg.Server.Mode != "release" && cfg.Server.Mode != "test" {
		return fmt.Errorf("无效的运行模式: %s，仅支持 debug/release/test", cfg.Server.Mode)
	}
	if cfg.Database.Path == "" {
		return fmt.Errorf("数据库路径不能为空")
	}
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("JWT 密钥不能为空")
	}
	if cfg.Agent.ReportTimeout <= 0 {
		return fmt.Errorf("报告生成超时必须大于0")
	}
	if cfg.Agent.MaxConcurrentSections <= 0 {
		return fmt.Errorf("章节最大并行数必须大于0")
	}
	return nil
}
