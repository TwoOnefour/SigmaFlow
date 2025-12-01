package config

import (
	"os"
	"strconv"
	"time"
)

// Config 应用配置
type Config struct {
	Exchange   ExchangeConfig
	LLM        LLMConfig
	Trading    TradingConfig
	Scheduler  SchedulerConfig
	Logging    LoggingConfig
	Notify     NotifyConfig
	RiskLimit  RiskLimitConfig
}

// ExchangeConfig 交易所配置
type ExchangeConfig struct {
	APIKey     string
	APISecret  string
	Passphrase string
	Simulate   bool
}

// LLMConfig AI模型配置
type LLMConfig struct {
	Provider string // gemini, openai, etc.
	APIKey   string
	Model    string
}

// TradingConfig 交易配置
type TradingConfig struct {
	BaseCurrency  string // 基础货币 (USDT)
	QuoteCurrency string // 交易货币 (BTC)
	MinOrderSize  float64
	MaxOrderSize  float64
}

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	Enabled  bool
	Interval time.Duration
	TimeZone string
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string // debug, info, warn, error
	Format string // json, text
	Output string // stdout, file path
}

// NotifyConfig 通知配置
type NotifyConfig struct {
	Enabled      bool
	TelegramBot  string
	TelegramChat string
}

// RiskLimitConfig 风险控制配置
type RiskLimitConfig struct {
	MaxPositionPct   float64 // 最大仓位比例
	MaxDailyLossPct  float64 // 每日最大亏损比例
	StopLossEnabled  bool
	TakeProfitEnabled bool
}

// LoadFromEnv 从环境变量加载配置
func LoadFromEnv() *Config {
	simulate := os.Getenv("OKX_SIMULATE") == "1"
	okxPrefix := "OKX_"
	if simulate {
		okxPrefix = "OKX_SIMULATE_"
	}

	interval, err := time.ParseDuration(os.Getenv("SCHEDULER_INTERVAL"))
	if err != nil {
		interval = 24 * time.Hour // 默认每天执行一次
	}

	return &Config{
		Exchange: ExchangeConfig{
			APIKey:     os.Getenv(okxPrefix + "API_KEY"),
			APISecret:  os.Getenv(okxPrefix + "API_SECRET"),
			Passphrase: os.Getenv(okxPrefix + "API_PASSPHRASE"),
			Simulate:   simulate,
		},
		LLM: LLMConfig{
			Provider: getEnvWithDefault("LLM_PROVIDER", "gemini"),
			APIKey:   os.Getenv("GEMINI_API_KEY"),
			Model:    getEnvWithDefault("LLM_MODEL", "gemini-2.5-pro"),
		},
		Trading: TradingConfig{
			BaseCurrency:  getEnvWithDefault("TRADING_BASE_CURRENCY", "USDT"),
			QuoteCurrency: getEnvWithDefault("TRADING_QUOTE_CURRENCY", "BTC"),
			MinOrderSize:  getEnvFloat("TRADING_MIN_ORDER_SIZE", 10.0),
			MaxOrderSize:  getEnvFloat("TRADING_MAX_ORDER_SIZE", 10000.0),
		},
		Scheduler: SchedulerConfig{
			Enabled:  os.Getenv("SCHEDULER_ENABLED") == "1",
			Interval: interval,
			TimeZone: getEnvWithDefault("SCHEDULER_TIMEZONE", "UTC"),
		},
		Logging: LoggingConfig{
			Level:  getEnvWithDefault("LOG_LEVEL", "info"),
			Format: getEnvWithDefault("LOG_FORMAT", "json"),
			Output: getEnvWithDefault("LOG_OUTPUT", "stdout"),
		},
		Notify: NotifyConfig{
			Enabled:      os.Getenv("NOTIFY_ENABLED") == "1",
			TelegramBot:  os.Getenv("TELEGRAM_BOT_TOKEN"),
			TelegramChat: os.Getenv("TELEGRAM_CHAT_ID"),
		},
		RiskLimit: RiskLimitConfig{
			MaxPositionPct:    getEnvFloat("RISK_MAX_POSITION_PCT", 0.5),
			MaxDailyLossPct:   getEnvFloat("RISK_MAX_DAILY_LOSS_PCT", 0.05),
			StopLossEnabled:   os.Getenv("RISK_STOP_LOSS_ENABLED") != "0",
			TakeProfitEnabled: os.Getenv("RISK_TAKE_PROFIT_ENABLED") != "0",
		},
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}
