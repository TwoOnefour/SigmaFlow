package trade

import (
	"context"
	"math"
	"okx/internal/model"
	"okx/internal/service/llm"
	"okx/internal/util"
	"resty.dev/v3"
	"strings"
)

type Service struct {
	market Market
	llm    *llm.Service
}

type Market interface {
	DoRestyRequest(req *resty.Request, method, path string, body ...interface{}) error
	GetRestClient() *resty.Client
	GetCandle(period int) ([]model.Candlestick, error)
	GetHolding(instId string) ([]model.HoldingData, error)
	GetBalance(ccy ...string) ([]model.BalanceData, error)
	Order(instId, side, sz string) error
}

type Trade interface {
	// Order bussiness
	Order()
	// Query statics
	Query()
}

func NewTradeService(c Market, llmService *llm.Service) *Service {
	return &Service{
		market: c,
		llm:    llmService,
	}
}

func (o *Service) GetCandle() ([]model.CandleWithIndicator, error) {
	const needCount = 231
	const outputCount = 30
	c, err := o.market.GetCandle(needCount)
	if err != nil {
		return nil, err
	}

	m := make([]float64, len(c))
	for i, v := range c {
		m[len(c)-1-i] = v.C
	}
	bbResults := util.CalculateBollingerBands(m, 20, 2.0)
	ma5 := util.CalculateMA(m, 5)
	ma50 := util.CalculateMA(m, 50)
	ma200 := util.CalculateMA(m, 200)

	candles := make([]model.CandleWithIndicator, outputCount)

	for i := 0; i < outputCount; i++ {
		originalCandle := c[i]
		valMA5 := ma5[len(ma5)-1-i]
		valMA50 := ma50[len(ma50)-1-i]
		valMA200 := ma200[len(ma200)-1-i]
		BBUpper := bbResults[len(bbResults)-1-i].Upper
		BBMid := bbResults[len(bbResults)-1-i].Middle
		BBLower := bbResults[len(bbResults)-1-i].Lower

		valMA5 = math.Floor(valMA5*10) / 10
		valMA50 = math.Floor(valMA50*10) / 10
		valMA200 = math.Floor(valMA50*10) / 10
		BBUpper = math.Floor(BBUpper*10) / 10
		BBMid = math.Floor(BBUpper*10) / 10
		BBLower = math.Floor(BBUpper*10) / 10
		candles[i] = model.CandleWithIndicator{
			TrendIndicators: model.TrendIndicators{
				MA5:     valMA5,
				MA50:    valMA50,
				MA200:   valMA200,
				BBUpper: BBUpper,
				BBMid:   BBMid,
				BBLower: BBLower,
			},
			Candlestick: originalCandle,
		}
	}

	return candles, nil
}

func (o *Service) GetBalance(instId ...string) ([]model.BalanceData, error) {
	if len(instId) == 0 {
		return o.market.GetBalance()
	}
	return o.market.GetBalance(instId...)
}

func (o *Service) Order(decision model.Decision) error {
	if decision.Action == "HOLD" {
		return nil
	}

	return o.market.Order("BTC-USDT", strings.ToLower(decision.Action), decision.Amount)
}

func (o *Service) AnalyzeMarket(ctx context.Context, holding []model.BalanceData, candle []model.CandleWithIndicator) (*model.Decision, error) {
	return o.llm.Completion(holding, candle)
}
