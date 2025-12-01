package risk

import (
	"okx/internal/config"
	"okx/internal/model"
	"okx/pkg/currency"
	"testing"
)

func TestManager_ValidateDecision(t *testing.T) {
	cfg := config.RiskLimitConfig{
		MaxPositionPct:    0.5,
		MaxDailyLossPct:   0.05,
		StopLossEnabled:   true,
		TakeProfitEnabled: true,
	}

	mgr := NewManager(cfg)
	mgr.SetInitialEquity(10000)

	tradeData := &model.TradeData{
		TotalEquity: 10000,
		AccountAssets: map[currency.Coin]*model.Asset{
			currency.USDT: {Equity: 5000},
			currency.BTC:  {Equity: 0.05},
		},
	}

	// 测试正常决策
	decision := &model.Decision{
		Action:      "BUY",
		PositionPct: 0.3,
	}

	err := mgr.ValidateDecision(decision, tradeData)
	if err != nil {
		t.Errorf("Should not return error for valid decision: %v", err)
	}

	// 测试超过最大仓位
	decision.PositionPct = 0.8
	err = mgr.ValidateDecision(decision, tradeData)
	if err != ErrMaxPositionExceeded {
		t.Errorf("Should return ErrMaxPositionExceeded, got: %v", err)
	}
}

func TestManager_AdjustDecision(t *testing.T) {
	cfg := config.RiskLimitConfig{
		MaxPositionPct: 0.5,
	}

	mgr := NewManager(cfg)

	decision := &model.Decision{
		Action:      "BUY",
		PositionPct: 0.8,
	}

	adjusted := mgr.AdjustDecision(decision)
	if adjusted.PositionPct != 0.5 {
		t.Errorf("Position should be adjusted to 0.5, got %f", adjusted.PositionPct)
	}
}

func TestManager_RecordTrade(t *testing.T) {
	cfg := config.RiskLimitConfig{}
	mgr := NewManager(cfg)

	mgr.RecordTrade("BUY", 100, 50000, 0, 10000)
	mgr.RecordTrade("SELL", 50, 52000, 100, 10100)

	history := mgr.GetTradeHistory(10)
	if len(history) != 2 {
		t.Errorf("Expected 2 trades, got %d", len(history))
	}

	dailyPnL := mgr.GetDailyPnL()
	if dailyPnL != 100 {
		t.Errorf("Expected daily PnL of 100, got %f", dailyPnL)
	}
}

func TestManager_ShouldStopLoss(t *testing.T) {
	cfg := config.RiskLimitConfig{
		StopLossEnabled: true,
	}
	mgr := NewManager(cfg)

	if !mgr.ShouldStopLoss(48000, 49000) {
		t.Error("Should trigger stop loss when price below stop loss")
	}

	if mgr.ShouldStopLoss(50000, 49000) {
		t.Error("Should not trigger stop loss when price above stop loss")
	}
}

func TestManager_ShouldTakeProfit(t *testing.T) {
	cfg := config.RiskLimitConfig{
		TakeProfitEnabled: true,
	}
	mgr := NewManager(cfg)

	if !mgr.ShouldTakeProfit(55000, 54000) {
		t.Error("Should trigger take profit when price above target")
	}

	if mgr.ShouldTakeProfit(53000, 54000) {
		t.Error("Should not trigger take profit when price below target")
	}
}

func TestManager_CalculatePositionSize(t *testing.T) {
	cfg := config.RiskLimitConfig{
		MaxPositionPct: 0.5,
	}
	mgr := NewManager(cfg)

	// 2%风险，入场50000，止损48000
	posSize := mgr.CalculatePositionSize(0.02, 50000, 48000, 10000)

	// 风险金额 = 10000 * 0.02 = 200
	// 每单位风险 = 50000 - 48000 = 2000
	// 仓位大小 = 200 / 2000 * 50000 = 5000
	// 但最大仓位 = 10000 * 0.5 = 5000
	if posSize > 5000 {
		t.Errorf("Position size should not exceed max position, got %f", posSize)
	}
}
