package backtest

import (
	"okx/internal/model"
	"okx/pkg/currency"
	"testing"
)

func TestEngine_Run(t *testing.T) {
	strategy := NewSimpleMAStrategy(5, 20)
	engine := NewEngine(strategy, 10000, 0.001)

	// 创建模拟K线数据
	candles := make([]model.Candlestick, 100)
	for i := 0; i < 100; i++ {
		price := 100.0 + float64(i)*0.5
		candles[i] = model.Candlestick{
			O:   price - 1,
			H:   price + 1,
			L:   price - 2,
			C:   price,
			Vol: 1000,
		}
	}

	pair := currency.NewPair(currency.USDT, currency.BTC)
	result, err := engine.Run(candles, pair)
	if err != nil {
		t.Fatalf("Engine.Run failed: %v", err)
	}

	if result.InitialCapital != 10000 {
		t.Errorf("Initial capital should be 10000, got %f", result.InitialCapital)
	}

	if result.FinalCapital <= 0 {
		t.Error("Final capital should be positive")
	}
}

func TestSimpleMAStrategy_Analyze(t *testing.T) {
	strategy := NewSimpleMAStrategy(5, 10)

	// 创建上涨趋势数据
	candles := make([]model.Candlestick, 15)
	for i := 0; i < 15; i++ {
		candles[i] = model.Candlestick{
			C: float64(100 + i*2),
		}
	}

	decision := strategy.Analyze(candles, 0)
	if decision == nil {
		t.Fatal("Decision should not be nil")
	}

	// 在上涨趋势中，短期MA应该高于长期MA，应该买入
	if decision.Action != "BUY" && decision.Action != "HOLD" {
		t.Logf("Action: %s, Reason: %s", decision.Action, decision.Reason)
	}
}

func TestResult_Report(t *testing.T) {
	result := &Result{
		InitialCapital:   10000,
		FinalCapital:     12000,
		TotalReturn:      0.2,
		AnnualizedReturn: 0.3,
		MaxDrawdown:      0.1,
		TotalTrades:      10,
		WinningTrades:    6,
		LosingTrades:     4,
		WinRate:          0.6,
		SharpeRatio:      1.5,
	}

	report := result.Report()
	if report == "" {
		t.Error("Report should not be empty")
	}

	// 检查报告包含关键信息
	if len(report) < 100 {
		t.Error("Report seems too short")
	}
}
