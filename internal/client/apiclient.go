package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"

	"github.com/abapdev/BR_Group/package/api"
)

const (
	apiURL             = "wss://ascendex.com/1/api/pro/v1/stream"
	readWait           = 60 * time.Second
	pingPeriod         = (readWait * 9) / 10
	maxMessageSize     = 1024 * 1024
	operationSubscribe = "sub"
)

type Response struct {
	M      string `json:"m"`
	Symbol string `json:"symbol"`
	Data   struct {
		Ts  int64    `json:"ts"`
		Bid []string `json:"bid"`
		Ask []string `json:"ask"`
	} `json:"data"`
}

type APIClient struct {
	conn *websocket.Conn
}

func (c *APIClient) Connection() error {
	u, err := url.Parse(apiURL)
	if err != nil {
		return err
	}

	header := http.Header{}
	header.Add("Content-Type", "application/json")

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return err
	}

	c.conn = conn
	return nil
}

func (c *APIClient) Disconnect() {
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *APIClient) SubscribeToChannel(symbol string) error {
	sub := map[string]interface{}{
		"op": operationSubscribe,
		"ch": fmt.Sprintf("bbo:%s", symbol),
	}

	subBytes, err := json.Marshal(sub)
	if err != nil {
		return err
	}

	if err := c.conn.WriteMessage(websocket.TextMessage, subBytes); err != nil {
		return err
	}

	return nil
}

func (c *APIClient) ReadMessagesFromChannel(ch chan<- api.BestOrderBook) {
	go func() {
		c.conn.SetReadLimit(maxMessageSize)
		c.conn.SetReadDeadline(time.Now().Add(readWait))
		c.conn.SetPongHandler(func(string) error {
			c.conn.SetReadDeadline(time.Now().Add(readWait))
			return nil
		})

		defer close(ch)

		for {
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				continue
			}

			var response Response
			if err := json.Unmarshal(message, &response); err != nil {
				continue
			}

			if len(response.Data.Ask) < 2 || len(response.Data.Bid) < 2 {
				continue
			}

			askAmount, err := strconv.ParseFloat(response.Data.Ask[0], 32)
			if err != nil {
				log.Fatal(err)
			}

			askPrice, err := strconv.ParseFloat(response.Data.Ask[1], 32)
			if err != nil {
				log.Fatal(err)
			}

			bidAmount, err := strconv.ParseFloat(response.Data.Bid[0], 32)
			if err != nil {
				log.Fatal(err)
			}

			bidPrice, err := strconv.ParseFloat(response.Data.Bid[1], 32)
			if err != nil {
				log.Fatal(err)
			}

			if askPrice <= bidPrice { // ???
				continue
			}

			data := api.BestOrderBook{
				Ask: api.Order{
					Amount: askAmount,
					Price:  askPrice,
				},
				Bid: api.Order{
					Amount: bidAmount,
					Price:  bidPrice,
				},
			}

			ch <- data
		}
	}()
}

func (c *APIClient) WriteMessagesToChannel() {
	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()
}

func NewAPIClient() *APIClient {
	return &APIClient{}
}
