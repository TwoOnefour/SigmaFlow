package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"okx/internal/client/exchange/okx"
	"okx/internal/client/llm/gemini"
	"okx/internal/service/llm"
	"okx/internal/service/trade"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("注意: 未找到 .env 文件，将尝试使用系统环境变量")
	}

	geminiApiKey := os.Getenv("GEMINI_API_KEY")
	okxSimulate := os.Getenv("OKX_SIMULATE")
	var okxKey string
	var okxSecret string
	var okxPhrase string
	okxPrefix := "OKX"
	if okxSimulate == "1" {
		okxPrefix += "_SIMULATE_"
	}
	okxKey = os.Getenv(okxPrefix + "API_KEY")
	okxSecret = os.Getenv(okxPrefix + "API_SECRET")
	okxPhrase = os.Getenv(okxPrefix + "API_PASSPHRASE")
	tradeService, err := di(geminiApiKey, okxKey, okxSecret, okxPhrase, okxSimulate)
	if err != nil {
		return
	}

	candle, err := tradeService.GetCandle()
	if err != nil {
		return
	}

	balance, err := tradeService.GetBalance("BTC", "USDT")
	if err != nil {
		return
	}
	completion, err := tradeService.AnalyzeMarket(context.Background(), balance, candle)
	if err != nil {
		return
	}
	fmt.Println(completion)
	fmt.Println(balance)
	err = tradeService.Order(*completion)
	if err != nil {
		return
	}
}

func di(geminiApiKey, okxKey, okxSecret, okxPhrase, okxSimulate string) (*trade.Service, error) {

	_okx, _ := okx.NewOkxClient(okxPhrase, okxSecret, okxKey, okxSimulate)

	_gemini, err := gemini.NewClient(context.Background(), geminiApiKey)
	if err != nil {
		return nil, err
	}
	_llm, err := llm.NewClient(context.Background(), _gemini)
	if err != nil {
		return nil, err
	}
	tradeService := trade.NewTradeService(_okx, _llm)
	return tradeService, nil
}
