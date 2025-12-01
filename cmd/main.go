package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"okx/internal/client/exchange/okx"
	"okx/internal/client/llm/gemini"
	"okx/internal/config"
	"okx/internal/logger"
	"okx/internal/notify"
	"okx/internal/risk"
	"okx/internal/scheduler"
	"okx/internal/service/llm"
	"okx/internal/service/trade"
	"okx/pkg/currency"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		logger.Info("注意: 未找到 .env 文件，将尝试使用系统环境变量")
	}

	// 加载配置
	cfg := config.LoadFromEnv()

	// 初始化日志
	log := logger.New(cfg.Logging.Level, cfg.Logging.Format, os.Stdout)
	logger.SetDefault(log)

	log.Info("SigmaFlow 量化交易机器人启动中...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 初始化服务
	tradeService, err := initServices(ctx, cfg)
	if err != nil {
		log.Error("初始化服务失败: %v", err)
		return
	}

	// 初始化风险管理
	riskMgr := risk.NewManager(cfg.RiskLimit)

	// 初始化通知管理
	notifyMgr := notify.NewManager(cfg.Notify.Enabled)
	notifyMgr.AddNotifier(notify.NewConsoleNotifier())
	if cfg.Notify.TelegramBot != "" && cfg.Notify.TelegramChat != "" {
		notifyMgr.AddNotifier(notify.NewTelegramNotifier(cfg.Notify.TelegramBot, cfg.Notify.TelegramChat))
	}

	// 交易对
	pair := currency.NewPair(
		currency.Coin(cfg.Trading.BaseCurrency),
		currency.Coin(cfg.Trading.QuoteCurrency),
	)

	// 检查是否启用调度器
	if cfg.Scheduler.Enabled {
		log.Info("调度器模式已启用，将定期执行交易策略")
		runWithScheduler(ctx, cfg, tradeService, pair, riskMgr, notifyMgr)
	} else {
		log.Info("单次执行模式")
		runOnce(ctx, tradeService, pair, riskMgr, notifyMgr)
	}
}

// initServices 初始化所有服务
func initServices(ctx context.Context, cfg *config.Config) (*trade.Service, error) {
	// 初始化OKX客户端
	simulate := "0"
	if cfg.Exchange.Simulate {
		simulate = "1"
	}
	okxClient, err := okx.NewOkxClient(
		cfg.Exchange.Passphrase,
		cfg.Exchange.APISecret,
		cfg.Exchange.APIKey,
		simulate,
	)
	if err != nil {
		return nil, err
	}

	// 初始化Gemini客户端
	geminiClient, err := gemini.NewClient(ctx, cfg.LLM.APIKey)
	if err != nil {
		return nil, err
	}

	// 初始化LLM服务
	llmService, err := llm.NewClient(ctx, geminiClient)
	if err != nil {
		return nil, err
	}

	// 创建交易服务
	return trade.NewTradeService(okxClient, llmService), nil
}

// runOnce 执行一次交易分析
func runOnce(ctx context.Context, tradeService *trade.Service, pair currency.Pair, riskMgr *risk.Manager, notifyMgr *notify.Manager) {
	log := logger.Default()

	task := trade.NewTradingTask(tradeService, pair, riskMgr, notifyMgr)
	if err := task.Run(ctx); err != nil {
		log.Error("交易任务执行失败: %v", err)
	}
}

// runWithScheduler 使用调度器运行
func runWithScheduler(ctx context.Context, cfg *config.Config, tradeService *trade.Service, pair currency.Pair, riskMgr *risk.Manager, notifyMgr *notify.Manager) {
	log := logger.Default()

	// 创建调度器
	sched, err := scheduler.New(cfg.Scheduler.TimeZone)
	if err != nil {
		log.Error("创建调度器失败: %v", err)
		return
	}

	// 创建交易任务
	task := trade.NewTradingTask(tradeService, pair, riskMgr, notifyMgr)

	// 添加每日定时任务 (UTC 00:01 执行)
	sched.AddDailyTask(task, 0, 1)

	// 启动调度器
	sched.Start(ctx)
	log.Info("调度器已启动，等待任务执行...")

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("收到停止信号，正在关闭...")
	sched.Stop()
	log.Info("SigmaFlow 已关闭")
}
