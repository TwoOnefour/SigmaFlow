package trade

import (
	"context"
	"okx/internal/logger"
	"okx/internal/model"
	"okx/internal/notify"
	"okx/internal/risk"
	"okx/pkg/currency"
)

// TradingTask 交易任务
type TradingTask struct {
	service     *Service
	pair        currency.Pair
	riskManager *risk.Manager
	notifier    *notify.Manager
}

// NewTradingTask 创建交易任务
func NewTradingTask(service *Service, pair currency.Pair, riskMgr *risk.Manager, notifier *notify.Manager) *TradingTask {
	return &TradingTask{
		service:     service,
		pair:        pair,
		riskManager: riskMgr,
		notifier:    notifier,
	}
}

// Name 任务名称
func (t *TradingTask) Name() string {
	return "TradingTask-" + t.pair.String()
}

// Run 执行交易任务
func (t *TradingTask) Run(ctx context.Context) error {
	log := logger.Default().WithField("task", t.Name())

	log.Info("Starting trading analysis...")

	// 1. 获取K线数据
	candles, err := t.service.GetCandle(t.pair)
	if err != nil {
		log.Error("Failed to get candles: %v", err)
		if t.notifier != nil {
			_ = t.notifier.SendErrorAlert(ctx, err)
		}
		return err
	}

	log.Info("Got %d candles", len(candles))

	// 2. 获取账户余额
	balance, err := t.service.GetBalance(t.pair.Base, t.pair.Quote)
	if err != nil {
		log.Error("Failed to get balance: %v", err)
		if t.notifier != nil {
			_ = t.notifier.SendErrorAlert(ctx, err)
		}
		return err
	}

	log.WithFields(map[string]interface{}{
		"total_equity": balance.TotalEquity,
	}).Info("Got account balance")

	// 设置初始资产用于风险控制
	if t.riskManager != nil {
		t.riskManager.SetInitialEquity(balance.TotalEquity)
	}

	// 3. 分析市场并获取交易决策
	decision, err := t.service.AnalyzeMarket(t.pair, balance, candles)
	if err != nil {
		log.Error("Failed to analyze market: %v", err)
		if t.notifier != nil {
			_ = t.notifier.SendErrorAlert(ctx, err)
		}
		return err
	}

	log.WithFields(map[string]interface{}{
		"action":       decision.Action,
		"position_pct": decision.PositionPct,
		"reason":       decision.Reason,
	}).Info("Got trading decision")

	// 4. 风险控制验证
	if t.riskManager != nil {
		if err := t.riskManager.ValidateDecision(decision, balance); err != nil {
			log.Warn("Risk validation failed: %v", err)
			// 调整决策
			decision = t.riskManager.AdjustDecision(decision)
		}
	}

	// 5. 发送交易通知
	if t.notifier != nil {
		_ = t.notifier.SendTradeAlert(ctx, decision)
	}

	// 6. 执行交易
	if decision.Action != "HOLD" {
		log.Info("Executing order...")
		if err := t.service.Order(*decision); err != nil {
			log.Error("Failed to execute order: %v", err)
			if t.notifier != nil {
				_ = t.notifier.SendErrorAlert(ctx, err)
			}
			return err
		}

		log.Info("Order executed successfully")

		// 记录交易
		if t.riskManager != nil {
			t.riskManager.RecordTrade(
				decision.Action,
				decision.PositionPct,
				candles[0].C, // 当前价格
				0,            // PnL在后续更新
				balance.TotalEquity,
			)
		}
	} else {
		log.Info("Decision is HOLD, no action taken")
	}

	return nil
}

// ExecuteOrder 带风险控制的下单
func (s *Service) ExecuteOrder(ctx context.Context, decision model.Decision, riskMgr *risk.Manager, balance *model.TradeData) error {
	log := logger.Default().WithField("action", decision.Action)

	// 风险验证
	if riskMgr != nil {
		if err := riskMgr.ValidateDecision(&decision, balance); err != nil {
			log.Warn("Risk validation warning: %v", err)
			// 继续执行但记录警告
		}
	}

	// 执行订单
	if err := s.Order(decision); err != nil {
		return err
	}

	log.Info("Order executed successfully")
	return nil
}
