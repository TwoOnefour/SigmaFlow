package risk

import (
	"errors"
	"okx/internal/config"
	"okx/internal/model"
	"sync"
	"time"
)

// Manager 风险控制管理器
type Manager struct {
	mu           sync.RWMutex
	config       config.RiskLimitConfig
	dailyPnL     float64
	dailyStart   time.Time
	trades       []TradeRecord
	initialEquity float64
}

// TradeRecord 交易记录
type TradeRecord struct {
	Timestamp   time.Time
	Action      string  // BUY, SELL
	Amount      float64
	Price       float64
	PnL         float64
	TotalEquity float64
}

var (
	ErrMaxPositionExceeded = errors.New("maximum position limit exceeded")
	ErrMaxDailyLossExceeded = errors.New("maximum daily loss limit exceeded")
	ErrInvalidDecision      = errors.New("invalid trading decision")
)

// NewManager 创建风险管理器
func NewManager(cfg config.RiskLimitConfig) *Manager {
	return &Manager{
		config:     cfg,
		dailyStart: time.Now().UTC().Truncate(24 * time.Hour),
		trades:     make([]TradeRecord, 0),
	}
}

// SetInitialEquity 设置初始资产
func (m *Manager) SetInitialEquity(equity float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.initialEquity = equity
}

// ValidateDecision 验证交易决策是否符合风险控制
func (m *Manager) ValidateDecision(decision *model.Decision, tradeData *model.TradeData) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 重置每日统计
	m.resetDailyIfNeeded()

	if decision == nil {
		return ErrInvalidDecision
	}

	// 检查最大仓位限制
	if decision.PositionPct > m.config.MaxPositionPct {
		return ErrMaxPositionExceeded
	}

	// 检查每日最大亏损
	if m.initialEquity > 0 {
		currentLossRatio := -m.dailyPnL / m.initialEquity
		if currentLossRatio >= m.config.MaxDailyLossPct {
			return ErrMaxDailyLossExceeded
		}
	}

	return nil
}

// AdjustDecision 根据风险控制调整交易决策
func (m *Manager) AdjustDecision(decision *model.Decision) *model.Decision {
	if decision == nil {
		return decision
	}

	// 限制仓位比例
	if decision.PositionPct > m.config.MaxPositionPct {
		decision.PositionPct = m.config.MaxPositionPct
	}

	return decision
}

// RecordTrade 记录交易
func (m *Manager) RecordTrade(action string, amount, price, pnl, totalEquity float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.resetDailyIfNeeded()

	record := TradeRecord{
		Timestamp:   time.Now().UTC(),
		Action:      action,
		Amount:      amount,
		Price:       price,
		PnL:         pnl,
		TotalEquity: totalEquity,
	}

	m.trades = append(m.trades, record)
	m.dailyPnL += pnl
}

// GetDailyPnL 获取当日盈亏
func (m *Manager) GetDailyPnL() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.dailyPnL
}

// GetTradeHistory 获取交易历史
func (m *Manager) GetTradeHistory(limit int) []TradeRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.trades) {
		limit = len(m.trades)
	}

	// 返回最近的交易记录
	start := len(m.trades) - limit
	if start < 0 {
		start = 0
	}

	result := make([]TradeRecord, limit)
	copy(result, m.trades[start:])
	return result
}

// ShouldStopLoss 检查是否应该止损
func (m *Manager) ShouldStopLoss(currentPrice, stopLossPrice float64) bool {
	if !m.config.StopLossEnabled {
		return false
	}
	return currentPrice <= stopLossPrice
}

// ShouldTakeProfit 检查是否应该止盈
func (m *Manager) ShouldTakeProfit(currentPrice, takeProfitPrice float64) bool {
	if !m.config.TakeProfitEnabled {
		return false
	}
	return currentPrice >= takeProfitPrice
}

// CalculatePositionSize 根据风险计算仓位大小
// riskPerTrade: 每笔交易愿意承担的风险比例 (如0.02表示2%)
// entryPrice: 入场价格
// stopLossPrice: 止损价格
// totalEquity: 总资产
func (m *Manager) CalculatePositionSize(riskPerTrade, entryPrice, stopLossPrice, totalEquity float64) float64 {
	if entryPrice <= 0 || stopLossPrice <= 0 || totalEquity <= 0 {
		return 0
	}

	// 每笔交易的风险金额
	riskAmount := totalEquity * riskPerTrade

	// 每单位的风险 (入场价到止损价的差距)
	riskPerUnit := entryPrice - stopLossPrice
	if riskPerUnit <= 0 {
		return 0
	}

	// 仓位大小 (以基础货币计)
	positionSize := riskAmount / riskPerUnit

	// 转换为金额
	positionValue := positionSize * entryPrice

	// 限制最大仓位
	maxPosition := totalEquity * m.config.MaxPositionPct
	if positionValue > maxPosition {
		positionValue = maxPosition
	}

	return positionValue
}

func (m *Manager) resetDailyIfNeeded() {
	now := time.Now().UTC().Truncate(24 * time.Hour)
	if now.After(m.dailyStart) {
		m.dailyPnL = 0
		m.dailyStart = now
	}
}
