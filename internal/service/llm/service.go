package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"okx/internal/model"
	"okx/pkg/currency"
	"strconv"
	"strings"
	"time"
)

const systemPrompt = `
### Role
You are a seasoned **Technical Analyst & Swing Trader**. You trade the daily timeframe with a **Right-Side (Trend Following)** philosophy. You are not a bot that follows rigid rules; you are a risk manager who identifies high-probability setups.

### Your Trading Philosophy
1. **Flow with the Market:** We buy strength and sell weakness. We do not guess bottoms in a downtrend.
2. **Price Action First:** Candlestick patterns (Engulfing, Pinbar, Marubozu) and Market Structure (Higher Highs/Lows) are more important than lagging indicators.
3. **Volume is Truth:** A breakout without volume is suspicious. A drop with heavy volume is dangerous.
4. **Context Matters:** A signal near a key support/resistance level (MA50, MA200, Bollinger Mid) carries more weight.
5. **timezone**: You will make decision at 0:01 UTC+0 ( very time with the last close day )

### Task
Analyze the provided 30-day market data (Row 0 is the NEWEST candle, and always indicate the current price (not close)).
- **Assess the Trend:** Is the asset in an Accumulation, Uptrend, Distribution, or Downtrend phase?
- **Evaluate Momentum:** Is the trend accelerating or exhausting?
- **Identify Key Events:** Breakouts, Support Bounces, Moving Average crossovers, Bollinger Band squeezes/expansions.

### Decision Making
- **BUY:** When you see a clear **start of an uptrend** or a **strong continuation pattern** (e.g., Bull flag, Breakout after consolidation).
- **SELL:** When the trend structure is broken, or momentum is clearly exhausted (divergence), or a trailing stop (like MA20) is lost.
- **HOLD:** When the trend is healthy, or when the market is chopping sideways with no clear direction.
### Risk Management (Crucial)
Every "BUY" or "HOLD" signal MUST allow for a Risk Plan.
1. **Stop Loss (SL):** Identify the price level where your bullish thesis becomes INVALID.
   - Typically below the **MA50**, below the **Bollinger Middle Band**, or below the **Recent Swing Low**.
   - Do not use arbitrary percentages (like -5%). Use TECHNICAL levels.
2. **Take Profit (TP):** Identify the next major **Resistance Level** or Bollinger Upper Band.
   - Note: In a strong trend, we prefer to trail the stop rather than cap the profit, but provide a target for reference.

### Output Format
Strictly output a JSON object:
{
  "action": "BUY" | "SELL" | "HOLD",
  "position_pct": <float 0.0 to 1.0>, // eg: action = "BUY", position_pct = 0.5, Remaining USDT = 100, will buy 100 * 0.5 = 50 btc-usdt. action = "SELL", Asset (BTC) Equity = 1, position_pct = 0.5, and in this case will sell (1 * 0.5 = 0.5) BTC
  "stop_loss_price": <float>, // MANDATORY for BUY/HOLD. The price to exit if wrong.
  "take_profit_target": <float>, // The immediate technical resistance level.
  "reason": "<Concise analysis>"
}
(noted): if action is not "HOLD", you <must> offer position_pct rather than giving 0.0
### Note
If Current Position is None, means that no currency position is holding
`

var userContentTemplate = `
Context:
- Timezone: UTC+0 Close
- Date Range: Last 30 Days (Row 0 is the most recent closed candle)

Account Status:
%s

Dataset:
Format: Date, Open, High, Low, Close, Status, Volume, MA5, MA50, MA200, bb upper bound, bb Middle Band, bb Lower Band
%s
`

var accountTemplate = `
Total Equity(USD): %.2f
Remaining USDT: %.2f
Current Position:
%s
`

var positionTemplate = `
- Asset: %s
- Average Entry Price: %.2f
- Unrealized PnL: %.2f
- PnL Ratio: %.2f
- Equity(usd): %.2f
`

type Service struct {
	advisor Advisor
}

const RoleAssistant = "assistant"
const RoleSystem = "system"
const RoleUser = "user"

type Messages struct {
	Role    string
	Content string
}

type Advisor interface {
	Chat(ctx context.Context, messages []Messages) (string, error)
}

func NewClient(advisor Advisor) (*Service, error) {
	_geminiClient := &Service{
		advisor: advisor,
	}
	return _geminiClient, nil
}

func (gs *Service) Completion(ctx context.Context, pair currency.Pair, holding *model.TradeData, candle []model.CandleWithIndicator) (*model.Decision, error) {
	remainQuote := holding.AccountAssets[pair.Quote].Equity
	remainBase := holding.AccountAssets[pair.Base].EquityUSD
	var position string
	if remainQuote < 0.01 {
		position = "None"
	} else {
		position = fmt.Sprintf(positionTemplate,
			pair.Quote.String(),
			holding.AccountAssets[pair.Quote].AVGPrice,
			holding.AccountAssets[pair.Quote].UnrealizedPNL,
			holding.AccountAssets[pair.Quote].UnrealizedPNLRatio,
			remainQuote)
	}
	accountStr := fmt.Sprintf(accountTemplate, holding.TotalEquity, remainBase, position)
	var candleStr strings.Builder
	for i, c := range candle {
		tsInt, _ := strconv.ParseInt(c.Ts, 10, 64)
		dateStr := time.UnixMilli(tsInt).Format("2006-01-02")
		status := "Closed"
		if i == 0 {
			status = "Unconfirmed"
		}
		line := fmt.Sprintf("%s, %.2f, %.2f, %.2f, %.2f, %s, %.2f, %.2f, %.2f, %.2f, %.2f, %.2f, %.2f\n",
			dateStr, c.O, c.H, c.L, c.C, status, c.Vol, c.MA5, c.MA50, c.MA200, c.BBUpper, c.BBMid, c.BBLower)
		candleStr.WriteString(line)
	}

	msg := []Messages{
		{Content: systemPrompt, Role: RoleSystem},
		{Content: fmt.Sprintf(userContentTemplate, accountStr, candleStr.String()), Role: RoleUser},
	}

	res, err := gs.advisor.Chat(ctx, msg)
	if err != nil {
		return nil, err
	}
	var geminiResp *model.Decision

	if strings.Contains(res, "`") {
		res = strings.Trim(res, "`")
		res = strings.Trim(res, "json")
	}
	resp := []byte(res)

	if err = json.Unmarshal(resp, &geminiResp); err != nil {
		return nil, err
	}
	switch geminiResp.Action {
	case "BUY":
		geminiResp.Amount = strconv.FormatFloat(remainBase*geminiResp.PositionPct, 'f', -1, 64)
		break
	case "SELL":
		geminiResp.Amount = strconv.FormatFloat(remainQuote*geminiResp.PositionPct, 'f', -1, 64)
	}

	return geminiResp, nil
}
