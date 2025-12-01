package util

import (
	"math"
	"okx/internal/model"
)

func CalculateBollingerBands(closes []float64, period int, k float64) []model.BollingerResult {
	length := len(closes)
	results := make([]model.BollingerResult, length)

	for i := 0; i < length; i++ {
		if i < period-1 {
			results[i] = model.BollingerResult{}
			continue
		}

		window := closes[i-period+1 : i+1]

		sum := 0.0
		for _, price := range window {
			sum += price
		}
		ma := sum / float64(period)

		varianceSum := 0.0
		for _, price := range window {
			varianceSum += math.Pow(price-ma, 2)
		}

		variance := varianceSum / float64(period)
		stdDev := math.Sqrt(variance)

		upper := ma + (k * stdDev)
		lower := ma - (k * stdDev)

		results[i] = model.BollingerResult{
			Middle: ma,
			Upper:  upper,
			Lower:  lower,
		}
	}

	return results
}

func CalculateMA(data []float64, period int) []float64 {
	if len(data) < period {
		return nil
	}

	sma := make([]float64, len(data)-period+1)
	var sum float64

	for i := 0; i < period; i++ {
		sum += data[i]
	}
	sma[0] = sum / float64(period)
	for i := period; i < len(data); i++ {
		sum -= data[i-period]
		sum += data[i]
		sma[i-period+1] = sum / float64(period)
	}

	return sma
}

// CalculateEMA 计算指数移动平均线
func CalculateEMA(data []float64, period int) []float64 {
	if len(data) < period {
		return nil
	}

	ema := make([]float64, len(data))
	multiplier := 2.0 / float64(period+1)

	// 第一个EMA值使用SMA
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += data[i]
	}
	ema[period-1] = sum / float64(period)

	// 计算后续EMA
	for i := period; i < len(data); i++ {
		ema[i] = (data[i]-ema[i-1])*multiplier + ema[i-1]
	}

	return ema
}

// RSIResult RSI计算结果
type RSIResult struct {
	RSI float64
}

// CalculateRSI 计算相对强弱指数
// period: 通常为14
func CalculateRSI(closes []float64, period int) []float64 {
	if len(closes) < period+1 {
		return nil
	}

	rsi := make([]float64, len(closes))

	// 计算价格变化
	gains := make([]float64, len(closes))
	losses := make([]float64, len(closes))

	for i := 1; i < len(closes); i++ {
		change := closes[i] - closes[i-1]
		if change > 0 {
			gains[i] = change
		} else {
			losses[i] = -change
		}
	}

	// 计算初始平均涨跌
	var avgGain, avgLoss float64
	for i := 1; i <= period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// 计算第一个RSI
	if avgLoss == 0 {
		rsi[period] = 100
	} else {
		rs := avgGain / avgLoss
		rsi[period] = 100 - (100 / (1 + rs))
	}

	// 计算后续RSI (使用平滑方法)
	for i := period + 1; i < len(closes); i++ {
		avgGain = (avgGain*float64(period-1) + gains[i]) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + losses[i]) / float64(period)

		if avgLoss == 0 {
			rsi[i] = 100
		} else {
			rs := avgGain / avgLoss
			rsi[i] = 100 - (100 / (1 + rs))
		}
	}

	return rsi
}

// MACDResult MACD计算结果
type MACDResult struct {
	MACD      float64 // MACD线 (快线 - 慢线)
	Signal    float64 // 信号线 (MACD的EMA)
	Histogram float64 // 柱状图 (MACD - Signal)
}

// CalculateMACD 计算MACD指标
// fastPeriod: 快速EMA周期，通常为12
// slowPeriod: 慢速EMA周期，通常为26
// signalPeriod: 信号线周期，通常为9
func CalculateMACD(closes []float64, fastPeriod, slowPeriod, signalPeriod int) []MACDResult {
	if len(closes) < slowPeriod {
		return nil
	}

	results := make([]MACDResult, len(closes))

	fastEMA := CalculateEMA(closes, fastPeriod)
	slowEMA := CalculateEMA(closes, slowPeriod)

	// 计算MACD线
	macdLine := make([]float64, len(closes))
	for i := slowPeriod - 1; i < len(closes); i++ {
		macdLine[i] = fastEMA[i] - slowEMA[i]
	}

	// 计算信号线 (MACD的EMA)
	macdForSignal := macdLine[slowPeriod-1:]
	signalLine := CalculateEMA(macdForSignal, signalPeriod)

	// 填充结果
	signalStartIdx := slowPeriod - 1 + signalPeriod - 1
	for i := signalStartIdx; i < len(closes); i++ {
		signalIdx := i - (slowPeriod - 1)
		results[i] = MACDResult{
			MACD:      macdLine[i],
			Signal:    signalLine[signalIdx],
			Histogram: macdLine[i] - signalLine[signalIdx],
		}
	}

	return results
}

// CalculateATR 计算平均真实波幅
// high, low, close: 对应的价格数据
// period: 通常为14
func CalculateATR(high, low, close []float64, period int) []float64 {
	if len(high) < period+1 || len(high) != len(low) || len(high) != len(close) {
		return nil
	}

	atr := make([]float64, len(high))

	// 计算真实波幅 (True Range)
	tr := make([]float64, len(high))
	tr[0] = high[0] - low[0] // 第一个TR

	for i := 1; i < len(high); i++ {
		hl := high[i] - low[i]
		hc := math.Abs(high[i] - close[i-1])
		lc := math.Abs(low[i] - close[i-1])
		tr[i] = math.Max(hl, math.Max(hc, lc))
	}

	// 计算初始ATR (SMA)
	var sum float64
	for i := 0; i < period; i++ {
		sum += tr[i]
	}
	atr[period-1] = sum / float64(period)

	// 计算后续ATR (使用Wilder's平滑方法)
	for i := period; i < len(high); i++ {
		atr[i] = (atr[i-1]*float64(period-1) + tr[i]) / float64(period)
	}

	return atr
}

// StochResult 随机指标结果
type StochResult struct {
	K float64 // %K线
	D float64 // %D线 (K的SMA)
}

// CalculateStochastic 计算随机指标
// kPeriod: %K周期，通常为14
// dPeriod: %D周期，通常为3
func CalculateStochastic(high, low, close []float64, kPeriod, dPeriod int) []StochResult {
	if len(high) < kPeriod || len(high) != len(low) || len(high) != len(close) {
		return nil
	}

	results := make([]StochResult, len(high))
	kValues := make([]float64, len(high))

	// 计算%K
	for i := kPeriod - 1; i < len(high); i++ {
		// 找周期内最高和最低
		highestHigh := high[i-kPeriod+1]
		lowestLow := low[i-kPeriod+1]
		for j := i - kPeriod + 2; j <= i; j++ {
			if high[j] > highestHigh {
				highestHigh = high[j]
			}
			if low[j] < lowestLow {
				lowestLow = low[j]
			}
		}

		// %K = (C - LL) / (HH - LL) * 100
		if highestHigh == lowestLow {
			kValues[i] = 50
		} else {
			kValues[i] = (close[i] - lowestLow) / (highestHigh - lowestLow) * 100
		}
	}

	// 计算%D (K的SMA)
	startIdx := kPeriod - 1 + dPeriod - 1
	for i := startIdx; i < len(high); i++ {
		var sum float64
		for j := i - dPeriod + 1; j <= i; j++ {
			sum += kValues[j]
		}
		results[i] = StochResult{
			K: kValues[i],
			D: sum / float64(dPeriod),
		}
	}

	return results
}
