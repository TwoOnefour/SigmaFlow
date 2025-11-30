package currency

type Coin string

const (
	BTC  Coin = "BTC"
	ETH  Coin = "ETH"
	USDT Coin = "USDT"
	USD  Coin = "USD"
)

func (c Coin) IsStable() bool {
	return c == USDT || c == USD
}

func (c Coin) String() string {
	return string(c)
}

type Pair struct {
	Base  Coin
	Quote Coin
}

func NewPair(base, quote Coin) Pair {
	return Pair{Base: base, Quote: quote}
}

func (p Pair) String() string {
	return string(p.Base) + "_" + string(p.Quote) // 比如用下划线区分
}
