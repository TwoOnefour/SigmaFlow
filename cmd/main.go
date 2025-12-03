package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/twoonefour/sigmaflow/internal/client/exchange/okx"
	"github.com/twoonefour/sigmaflow/internal/service/cron"
	"github.com/twoonefour/sigmaflow/internal/service/llm"
	"github.com/twoonefour/sigmaflow/internal/service/trade"
	"github.com/twoonefour/sigmaflow/pkg/currency"
	"github.com/twoonefour/sigmaflow/pkg/llm/gemini"
	"log"
	"os"
	"time"
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
	pair := currency.NewPair(currency.USDT, currency.BTC)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	c := cron.NewService()
	err = c.AddCron("1 8 * * *", func() {
		err := run(ctx, tradeService, pair)
		if err != nil {
			log.Println(err.Error())
		}
	})
	if err != nil {
		return
	}
	// run(ctx, tradeService, pair)
	select {}
}

// Dependency Injection
func di(geminiApiKey, okxKey, okxSecret, okxPhrase, okxSimulate string) (*trade.Service, error) {
	_okx, _ := okx.NewOkxClient(okxPhrase, okxSecret, okxKey, okxSimulate)
	_gemini, err := gemini.NewClient(geminiApiKey, "gemini-2.5-pro", 32768)
	if err != nil {
		return nil, err
	}
	_llm, err := llm.NewClient(_gemini)
	if err != nil {
		return nil, err
	}
	tradeService := trade.NewTradeService(_okx, _llm)
	return tradeService, nil
}

func run(ctx context.Context, tradeService *trade.Service, pair currency.Pair) error {
	candle, err := tradeService.GetCandle(pair)
	if err != nil {
		return err
	}

	balance, err := tradeService.GetBalance(ctx, pair.Base, pair.Quote)
	if err != nil {
		return err
	}
	completion, err := tradeService.AnalyzeMarket(ctx, currency.NewPair(currency.USDT, currency.BTC), balance, candle)
	if err != nil {
		return err
	}
	log.Println(fmt.Sprintf("AI决策:%s, 数量：%.2f %%, 理由：%s", completion.Action, completion.PositionPct*100, completion.Reason))
	err = tradeService.Order(*completion)
	if err != nil {
		return err
	}
	AfterOrder, err := tradeService.GetBalance(ctx, pair.Base, pair.Quote)
	if err != nil {
		return err
	}
	log.Println(fmt.Sprintf("目前剩余(单位USD) %s:%.2f, %s:%.2f",
		pair.Base.String(), AfterOrder.AccountAssets[pair.Base].EquityUSD, pair.Quote.String(), AfterOrder.AccountAssets[pair.Quote].EquityUSD))
	return nil
}
