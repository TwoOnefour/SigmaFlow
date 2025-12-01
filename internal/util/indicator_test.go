package util

import (
	"math"
	"testing"
)

func TestCalculateEMA(t *testing.T) {
	data := []float64{100, 102, 104, 103, 105, 107, 106, 108, 110, 109}
	period := 5

	ema := CalculateEMA(data, period)

	if ema == nil {
		t.Fatal("EMA should not be nil")
	}

	// EMA应该从period-1开始有值
	if ema[period-1] == 0 {
		t.Error("First EMA value should not be 0")
	}

	// 后续EMA应该有值
	if ema[len(data)-1] == 0 {
		t.Error("Last EMA value should not be 0")
	}
}

func TestCalculateRSI(t *testing.T) {
	// 创建一个上涨趋势的数据
	data := make([]float64, 20)
	for i := 0; i < 20; i++ {
		data[i] = float64(100 + i)
	}

	rsi := CalculateRSI(data, 14)

	if rsi == nil {
		t.Fatal("RSI should not be nil")
	}

	// 在持续上涨趋势中，RSI应该接近100
	lastRSI := rsi[len(rsi)-1]
	if lastRSI < 70 {
		t.Errorf("RSI in uptrend should be high, got %f", lastRSI)
	}
}

func TestCalculateRSI_Downtrend(t *testing.T) {
	// 创建一个下跌趋势的数据
	data := make([]float64, 20)
	for i := 0; i < 20; i++ {
		data[i] = float64(120 - i)
	}

	rsi := CalculateRSI(data, 14)

	if rsi == nil {
		t.Fatal("RSI should not be nil")
	}

	// 在持续下跌趋势中，RSI应该接近0
	lastRSI := rsi[len(rsi)-1]
	if lastRSI > 30 {
		t.Errorf("RSI in downtrend should be low, got %f", lastRSI)
	}
}

func TestCalculateMACD(t *testing.T) {
	// 创建测试数据
	data := make([]float64, 50)
	for i := 0; i < 50; i++ {
		data[i] = float64(100 + i)
	}

	macd := CalculateMACD(data, 12, 26, 9)

	if macd == nil {
		t.Fatal("MACD should not be nil")
	}

	// 检查最后一个MACD值
	lastMACD := macd[len(macd)-1]
	if lastMACD.MACD == 0 && lastMACD.Signal == 0 {
		t.Error("MACD values should not all be 0")
	}
}

func TestCalculateATR(t *testing.T) {
	high := []float64{105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119}
	low := []float64{95, 96, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109}
	close := []float64{100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114}

	atr := CalculateATR(high, low, close, 14)

	if atr == nil {
		t.Fatal("ATR should not be nil")
	}

	// ATR应该大于0
	lastATR := atr[len(atr)-1]
	if lastATR <= 0 {
		t.Errorf("ATR should be positive, got %f", lastATR)
	}
}

func TestCalculateStochastic(t *testing.T) {
	high := []float64{105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120}
	low := []float64{95, 96, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110}
	close := []float64{100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115}

	stoch := CalculateStochastic(high, low, close, 14, 3)

	if stoch == nil {
		t.Fatal("Stochastic should not be nil")
	}

	// 检查最后的K和D值
	lastStoch := stoch[len(stoch)-1]
	if lastStoch.K < 0 || lastStoch.K > 100 {
		t.Errorf("Stochastic K should be between 0 and 100, got %f", lastStoch.K)
	}
	if lastStoch.D < 0 || lastStoch.D > 100 {
		t.Errorf("Stochastic D should be between 0 and 100, got %f", lastStoch.D)
	}
}

func TestCalculateMA(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5}
	period := 3

	ma := CalculateMA(data, period)

	if len(ma) != 3 {
		t.Errorf("Expected 3 MA values, got %d", len(ma))
	}

	// 验证第一个MA值
	expectedFirst := 2.0 // (1+2+3)/3
	if math.Abs(ma[0]-expectedFirst) > 0.0001 {
		t.Errorf("First MA should be %f, got %f", expectedFirst, ma[0])
	}
}

func TestCalculateBollingerBands(t *testing.T) {
	data := make([]float64, 25)
	for i := 0; i < 25; i++ {
		data[i] = float64(100 + i)
	}

	bb := CalculateBollingerBands(data, 20, 2.0)

	if len(bb) != 25 {
		t.Errorf("Expected 25 BB values, got %d", len(bb))
	}

	// 检查最后一个BB值
	lastBB := bb[len(bb)-1]
	if lastBB.Upper <= lastBB.Middle {
		t.Error("Upper band should be greater than middle")
	}
	if lastBB.Middle <= lastBB.Lower {
		t.Error("Middle band should be greater than lower")
	}
}
