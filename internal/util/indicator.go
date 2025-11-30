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
