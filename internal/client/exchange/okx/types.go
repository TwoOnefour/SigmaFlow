package okx

import "sigmaflow/pkg"

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
		Ccy           string          `json:"ccy"`
		OpenAvgPx     pkg.TextFloat64 `json:"openAvgPx"`
		SpotUpl       pkg.TextFloat64 `json:"spotUpl"`
		SpotUplRatio  pkg.TextFloat64 `json:"spotUplRatio"`
		EQusd         pkg.TextFloat64 `json:"equsd"`
		EQ            pkg.TextFloat64 `json:"eq"`
		TotalPNLRatio pkg.TextFloat64 `json:"totalPnlRatio"`
		TotalPNL      pkg.TextFloat64 `json:"totalPnl"`
	} `json:"details"`
	TotalEq pkg.TextFloat64 `json:"totalEq,string"`
}

type BalanceResponse struct {
	Code string        `json:"code"`
	Data []BalanceData `json:"data"`
	Msg  string        `json:"msg"`
}
