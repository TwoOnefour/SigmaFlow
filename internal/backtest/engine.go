package backtest

import (
	"fmt"
	"okx/internal/model"
	"okx/pkg/currency"
	"time"
)

// Result 回测结果
type Result struct {
	StartDate       time.Time
	EndDate         time.Time
	InitialCapital  float64
	FinalCapital    float64
	TotalReturn     float64 // 总收益率
	AnnualizedReturn float64 // 年化收益率
	MaxDrawdown     float64 // 最大回撤
	WinRate         float64 // 胜率
	TotalTrades     int
	WinningTrades   int
	LosingTrades    int
	SharpeRatio     float64 // 夏普比率
	Trades          []Trade
}

// Trade 回测中的交易记录
type Trade struct {
	Timestamp   time.Time
	Action      string
	Price       float64
	Amount      float64
	PnL         float64
	TotalEquity float64
}

// Strategy 策略接口
type Strategy interface {
	// Analyze 分析市场数据并返回交易决策
	Analyze(candles []model.Candlestick, position float64) *model.Decision
	Name() string
}

// Engine 回测引擎
type Engine struct {
	strategy       Strategy
	initialCapital float64
	commission     float64 // 手续费率
}

// NewEngine 创建回测引擎
func NewEngine(strategy Strategy, initialCapital, commission float64) *Engine {
	return &Engine{
		strategy:       strategy,
		initialCapital: initialCapital,
		commission:     commission,
	}
}

// Run 运行回测
func (e *Engine) Run(candles []model.Candlestick, pair currency.Pair) (*Result, error) {
	if len(candles) == 0 {
		return nil, fmt.Errorf("no candles provided")
	}

	result := &Result{
		InitialCapital: e.initialCapital,
		Trades:         make([]Trade, 0),
	}

	capital := e.initialCapital
	position := 0.0       // 持仓数量
	avgEntryPrice := 0.0  // 平均入场价格
	maxEquity := capital
	maxDrawdown := 0.0
	winningTrades := 0
	losingTrades := 0

	// 从有足够历史数据的位置开始
	startIdx := 30 // 需要至少30天历史数据来计算指标

	for i := startIdx; i < len(candles); i++ {
		// 获取历史数据窗口
		window := candles[:i+1]

		// 获取策略决策
		decision := e.strategy.Analyze(window, position)
		if decision == nil {
			continue
		}

		currentPrice := candles[i].C
		currentEquity := capital + position*currentPrice

		// 更新最大权益和回撤
		if currentEquity > maxEquity {
			maxEquity = currentEquity
		}
		drawdown := (maxEquity - currentEquity) / maxEquity
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}

		// 执行交易
		switch decision.Action {
		case "BUY":
			if capital > 0 {
				buyAmount := capital * decision.PositionPct
				buyAmountAfterFee := buyAmount * (1 - e.commission) // 扣除手续费
				units := buyAmountAfterFee / currentPrice
				
				// 更新平均入场价格
				if position > 0 {
					avgEntryPrice = (avgEntryPrice*position + currentPrice*units) / (position + units)
				} else {
					avgEntryPrice = currentPrice
				}
				
				position += units
				capital -= buyAmount

				result.Trades = append(result.Trades, Trade{
					Timestamp:   time.Now(), // 实际应使用蜡烛图时间
					Action:      "BUY",
					Price:       currentPrice,
					Amount:      units,
					TotalEquity: currentEquity,
				})
			}

		case "SELL":
			if position > 0 {
				sellUnits := position * decision.PositionPct
				sellValue := sellUnits * currentPrice * (1 - e.commission)
				// 计算实际盈亏：卖出价值 - 买入成本
				pnl := sellValue - (sellUnits * avgEntryPrice)
				capital += sellValue
				position -= sellUnits

				if pnl > 0 {
					winningTrades++
				} else {
					losingTrades++
				}

				result.Trades = append(result.Trades, Trade{
					Timestamp:   time.Now(),
					Action:      "SELL",
					Price:       currentPrice,
					Amount:      sellUnits,
					PnL:         pnl,
					TotalEquity: currentEquity,
				})
			}
		}
	}

	// 计算最终结果
	finalPrice := candles[len(candles)-1].C
	result.FinalCapital = capital + position*finalPrice
	result.TotalReturn = (result.FinalCapital - result.InitialCapital) / result.InitialCapital
	result.MaxDrawdown = maxDrawdown
	result.TotalTrades = len(result.Trades)
	result.WinningTrades = winningTrades
	result.LosingTrades = losingTrades

	if result.TotalTrades > 0 {
		result.WinRate = float64(winningTrades) / float64(result.TotalTrades)
	}

	// 计算年化收益率 (假设每条蜡烛图是1天)
	days := float64(len(candles) - startIdx)
	if days > 0 {
		result.AnnualizedReturn = (result.TotalReturn + 1) * (365 / days) - 1
	}

	return result, nil
}

// Report 生成回测报告
func (r *Result) Report() string {
	return fmt.Sprintf(`
=== Backtest Report ===
Strategy Performance:
  Initial Capital: $%.2f
  Final Capital:   $%.2f
  Total Return:    %.2f%%
  Annualized:      %.2f%%
  Max Drawdown:    %.2f%%

Trading Statistics:
  Total Trades:    %d
  Winning Trades:  %d
  Losing Trades:   %d
  Win Rate:        %.2f%%

Risk Metrics:
  Sharpe Ratio:    %.2f
=======================
`,
		r.InitialCapital,
		r.FinalCapital,
		r.TotalReturn*100,
		r.AnnualizedReturn*100,
		r.MaxDrawdown*100,
		r.TotalTrades,
		r.WinningTrades,
		r.LosingTrades,
		r.WinRate*100,
		r.SharpeRatio,
	)
}

// SimpleMAStrategy 简单MA策略（示例）
type SimpleMAStrategy struct {
	shortPeriod int
	longPeriod  int
}

// NewSimpleMAStrategy 创建简单MA策略
func NewSimpleMAStrategy(shortPeriod, longPeriod int) *SimpleMAStrategy {
	return &SimpleMAStrategy{
		shortPeriod: shortPeriod,
		longPeriod:  longPeriod,
	}
}

func (s *SimpleMAStrategy) Name() string {
	return fmt.Sprintf("SimpleMA(%d,%d)", s.shortPeriod, s.longPeriod)
}

func (s *SimpleMAStrategy) Analyze(candles []model.Candlestick, position float64) *model.Decision {
	if len(candles) < s.longPeriod {
		return nil
	}

	// 计算短期和长期MA
	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.C
	}

	shortMA := calculateSMA(closes[len(closes)-s.shortPeriod:], s.shortPeriod)
	longMA := calculateSMA(closes[len(closes)-s.longPeriod:], s.longPeriod)

	// 金叉买入，死叉卖出
	if shortMA > longMA && position == 0 {
		return &model.Decision{
			Action:      "BUY",
			PositionPct: 1.0,
			Reason:      "Golden cross detected",
		}
	} else if shortMA < longMA && position > 0 {
		return &model.Decision{
			Action:      "SELL",
			PositionPct: 1.0,
			Reason:      "Death cross detected",
		}
	}

	return &model.Decision{
		Action: "HOLD",
		Reason: "No signal",
	}
}

func calculateSMA(data []float64, period int) float64 {
	if len(data) < period {
		return 0
	}
	sum := 0.0
	for i := len(data) - period; i < len(data); i++ {
		sum += data[i]
	}
	return sum / float64(period)
}
