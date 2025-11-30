package model

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
	Ts      string  `json:"ts"`  // 时间戳
	O       float64 `json:"o"`   // Open (开盘价)
	H       float64 `json:"h"`   // High (最高价)
	L       float64 `json:"l"`   // Low (最低价)
	C       float64 `json:"c"`   // Close (收盘价)
	Vol     float64 `json:"vol"` // 当天交易数量
	Confirm string  `json:"confirm"`
}

type BollingerResult struct {
	Upper  float64
	Middle float64
	Lower  float64
}

type CandleWithIndicator struct {
	Candlestick
	TrendIndicators
}

type HoldingData struct {
	AvgPx       string `json:"avgPx"`       // 开仓均价
	Upl         string `json:"upl"`         // 未实现收益
	NotionalUsd string `json:"notionalUsd"` // 持仓总价值 (USD)
	InstId      string `json:"instId"`      // 交易对 BTC-USDT
}

type HoldingResponse struct {
	Code string        `json:"code"`
	Data []HoldingData `json:"data"`
	Msg  string        `json:"msg"`
}

type BalanceData struct {
	Details []struct {
		Ccy          string `json:"ccy"`
		OpenAvgPx    string `json:"openAvgPx"`
		SpotUpl      string `json:"spotUpl"`
		SpotUplRatio string `json:"spotUplRatio"`
		EQusd        string `json:"equsd"`
		EQ           string `json:"eq"`
	} `json:"details"`
	TotalEq string `json:"totalEq"`
}

type BalanceResponse struct {
	Code string        `json:"code"`
	Data []BalanceData `json:"data"`
	Msg  string        `json:"msg"`
}
