package okx

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/twoonefour/sigmaflow/internal/model"
	"github.com/twoonefour/sigmaflow/pkg/currency"
	"net/http"
	"resty.dev/v3"
	"strconv"
	"strings"
	"time"
)

type loginReq struct {
	Op   string `json:"op"`
	Args []args `json:"args"`
}

type loginResp struct {
	Event  string `json:"event"`
	Code   string `json:"code"`
	Msg    string `json:"msg"`
	ConnId string `json:"connId"`
}

type args struct {
	ApiKey     string `json:"apiKey"`
	Passphrase string `json:"passphrase"`
	Timestamp  string `json:"timestamp"`
	Sign       string `json:"sign"`
}

type candlesResp struct {
	Code string     `json:"code"`
	Msg  string     `json:"msg"`
	Data [][]string `json:"data"`
}

const wsApiBase = "wss://ws.okx.com:8443/ws/v5/public"
const restApiBase = "https://www.okx.com"

type Client struct {
	// The websocket connection.
	wsConn     []*websocket.Conn // use string to get a connection
	restClient *resty.Client
	wsRespChan chan map[string]interface{}
	apiKey     string
	secretKey  string
	passPhrase string
	simulate   bool
}

func (oc *Client) Sell(pct float64) error {
	panic("implement me")
}

func NewOkxClient(passPhrase, secretKey, apiKey, simulate string) (*Client, error) {
	_okxClient := &Client{
		apiKey:     apiKey,
		passPhrase: passPhrase,
		secretKey:  secretKey,
	}
	if simulate == "1" {
		_okxClient.simulate = true
	}

	//if err := _okxClient.initWebsocketConn(); err != nil {
	//	return nil, err
	//}
	_okxClient.restClient = resty.New().SetTimeout(5 * time.Second)
	_okxClient.restClient.SetProxy("http://127.0.0.1:10808")
	_okxClient.restClient.SetHeaders(map[string]string{
		"OK-ACCESS-PASSPHRASE": passPhrase,
		"OK-ACCESS-KEY":        apiKey,
	})
	_okxClient.restClient.SetBaseURL(restApiBase)
	return _okxClient, nil
}

func (oc *Client) newConn() (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(wsApiBase, nil)
	if err != nil {
		return nil, err
	}
	if err := oc.wsAuth(conn); err != nil {
		return nil, err
	}
	return conn, nil
}

func (oc *Client) wsAuth(conn *websocket.Conn) error {
	uriPath := "/users/self/verify"
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	req := &loginReq{
		Op: "login",
		Args: []args{
			{
				ApiKey:     oc.apiKey,
				Passphrase: oc.passPhrase,
				Timestamp:  ts,
				Sign:       AccessSign(ts, http.MethodGet, uriPath, "", oc.secretKey),
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return err
	}
	return nil
}

func (oc *Client) readMessage(conn *websocket.Conn) ([]byte, error) {
	messageType, data, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	var req map[string]interface{}
	err = json.Unmarshal(data, &req)
	if err != nil {
		return nil, err
	}
	oc.wsRespChan <- req
	fmt.Println(messageType)
	//fmt.Printf("%v\n", req)
	return nil, nil
}

func (oc *Client) initWebsocketConn() error {
	oc.wsRespChan = make(chan map[string]interface{})
	go oc.handleWsMessage()
	connSlice := make([]*websocket.Conn, 5)
	for i := 0; i < len(connSlice); i++ {
		newConn, err := oc.newConn()
		if err != nil {
			return err
		}
		go func() {
			_, err := oc.readMessage(newConn)
			if err != nil {
				fmt.Println("readMessage err:", err)
			}
		}()
		connSlice[i] = newConn
	}
	oc.wsConn = connSlice
	return nil
}

func (oc *Client) handleWsMessage() {
	for {
		select {
		case msg := <-oc.wsRespChan:
			fmt.Println("wsRespChan:", msg)
		}
	}
}

func (oc *Client) doRestyRequest(req *resty.Request, method, path string, body ...interface{}) error {
	signPath := path
	ts := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	var bodyStr string
	if len(body) > 0 && method == http.MethodPost {
		_body, ok := body[0].(map[string]string)
		if !ok {
			return fmt.Errorf("[client.doRestyRequest] Can not transform body to type (map[string]string)")
		}
		bodyBytes, err := json.Marshal(&_body)
		if err != nil {
			return err
		}
		bodyStr = string(bodyBytes)
		req.SetHeader("Content-Type", "application/json").SetBody(bodyBytes)
	} else {
		if req.QueryParams.Encode() != "" {
			signPath += "?" + req.QueryParams.Encode()
		}
	}
	if oc.simulate {
		req.SetHeader("x-simulated-trading", "1")
	}
	req.SetHeaders(map[string]string{
		"OK-ACCESS-SIGN":      AccessSign(ts, method, signPath, bodyStr, oc.secretKey),
		"OK-ACCESS-TIMESTAMP": ts,
	})
	resp, err := req.Execute(method, path)
	if err != nil || resp.StatusCode() != 200 {
		if resp != nil {
			return fmt.Errorf("httpcode: %d, response", resp.StatusCode())
		}
		return err
	}
	return nil
}

// Fetch last [period, now] data
func (oc *Client) GetCandle(pair currency.Pair, period int) ([]model.Candlestick, error) {
	m := make([]model.Candlestick, period)
	resp := &candlesResp{}
	req := oc.restClient.R().SetResult(resp).SetQueryParams(map[string]string{
		"instId": pair.Quote.String() + "-" + pair.Base.String(),
		"bar":    "1Dutc",
	})
	urlPath := "/api/v5/market/candles"
	cnt := 0
	for period > 0 {
		err := oc.doRestyRequest(req, http.MethodGet, urlPath)
		if err != nil || resp == nil || resp.Code != "0" {
			return nil, err
		}
		for i, data := range resp.Data {
			o, err := strconv.ParseFloat(data[1], 64)
			if err != nil {
				return nil, err
			}
			h, err := strconv.ParseFloat(data[2], 64)
			if err != nil {
				return nil, err
			}
			l, err := strconv.ParseFloat(data[3], 64)
			if err != nil {
				return nil, err
			}
			c, err := strconv.ParseFloat(data[4], 64)
			if err != nil {
				return nil, err
			}
			vol, err := strconv.ParseFloat(data[5], 64)
			if err != nil {
				return nil, err
			}
			m[i+cnt] = model.Candlestick{
				Ts:  data[0],
				O:   o,
				H:   h,
				L:   l,
				C:   c,
				Vol: vol,
			}
		}
		period -= len(resp.Data)
		cnt += len(resp.Data)
		req.SetQueryParam("after", m[cnt-1].Ts)
		req.SetQueryParam("limit", func(a, b int) string {
			if a > b {
				return strconv.Itoa(b)
			}
			return strconv.Itoa(a)
		}(period, 100))
	}

	return m, nil
}

func (oc *Client) GetBalance(ctx context.Context, coin ...currency.Coin) (*model.TradeData, error) {
	m := &BalanceResponse{}
	path := "/api/v5/account/balance"
	req := oc.restClient.R().WithContext(ctx)
	if len(coin) > 0 {
		c := make([]string, len(coin))
		for i, v := range coin {
			c[i] = v.String()
		}
		req.SetQueryParam("ccy", strings.Join(c, ","))
	}
	req.SetResult(m)
	err := oc.doRestyRequest(req, http.MethodGet, path)
	if err != nil {
		return nil, err
	}

	var res model.TradeData
	accountAssets := make(map[currency.Coin]*model.Asset)
	res.TotalEquity = m.Data[0].TotalEq.Float64()
	for _, c := range m.Data[0].Details {
		_coin := currency.Coin(c.Ccy)
		accountAssets[_coin] = &model.Asset{
			Equity:             c.EQ.Float64(),
			Currency:           _coin,
			TotalProfit:        c.TotalPNL.Float64(),
			UnrealizedPNL:      c.SpotUpl.Float64(),
			UnrealizedPNLRatio: c.SpotUplRatio.Float64(),
			AVGPrice:           c.OpenAvgPx.Float64(),
			EquityUSD:          c.EQusd.Float64(),
		}
	}
	res.AccountAssets = accountAssets
	return &res, nil
}

func (oc *Client) Order(instId, side, sz string) error {
	path := "/api/v5/trade/order"
	req := oc.restClient.R()
	body := map[string]string{
		"instId":  instId,
		"side":    side,
		"sz":      sz,
		"ordType": "market",
		"tdMode":  "cash",
	}
	return oc.doRestyRequest(req, http.MethodPost, path, body)
}
