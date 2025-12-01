package model

import "github.com/twoonefour/sigmaflow/pkg/currency"

type Decision struct {
	Action          string  `json:"action"`
	PositionPct     float64 `json:"position_pct"`
	Reason          string  `json:"reason"`
	StopLossPrice   float64 `json:"stop_loss_price"`
	TakeProfitPrice float64 `json:"take_profit_price"`
	Amount          string  `json:"amount"`
}

type TrendIndicators struct {
	MA5     float64 `json:"MA5"`
	MA50    float64 `json:"MA50"`
	MA200   float64 `json:"MA200"`
	BBUpper float64 `json:"bb_upper"`
	BBMid   float64 `json:"bb_mid"`
	BBLower float64 `json:"bb_lower"`
}

type Candlestick struct {
	Ts      string  `json:"ts"`      // 时间戳
	O       float64 `json:"o"`       // Open (开盘价)
	H       float64 `json:"h"`       // High (最高价)
	L       float64 `json:"l"`       // Low (最低价)
	C       float64 `json:"c"`       // Close (收盘价)
	Vol     float64 `json:"vol"`     // 当天交易数量
	Confirm string  `json:"confirm"` // 是否确认
}

type CandleWithIndicator struct {
	Candlestick
	TrendIndicators
}

type TradeData struct {
	AccountAssets map[currency.Coin]*Asset
	TotalEquity   float64 // 该账户总价值
}

type Asset struct {
	Equity             float64       // 该货币的数量
	Currency           currency.Coin // 货币种类，如USDT
	TotalProfit        float64       //  已实现盈亏
	UnrealizedPNL      float64       // 未实现盈利/亏损
	UnrealizedPNLRatio float64       // 未实现盈利亏损比
	AVGPrice           float64       // 平均买入价格
	EquityUSD          float64       // usd表示的价值
}

type GraphicData struct {
}
