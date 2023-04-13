package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"

	"github.com/abapdev/BR_Group/package/api"
)

func TestAPIClient_SubscribeToChannel(t *testing.T) {
	u, err := url.Parse(apiURL)
	if err != nil {
		return
	}

	header := http.Header{}
	header.Add("Content-Type", "application/json")

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return
	}

	type fields struct {
		conn *websocket.Conn
	}
	type args struct {
		symbol string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "correct ch",
			fields: fields{
				conn: conn,
			},
			args: args{
				symbol: "BTC/USDT",
			},
			wantErr: false,
		},
		{
			name: "incorrect ch",
			fields: fields{
				conn: conn,
			},
			args: args{
				symbol: "BTC/USD",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &APIClient{
				conn: tt.fields.conn,
			}
			if err := c.SubscribeToChannel(tt.args.symbol); (err != nil) != tt.wantErr {
				t.Errorf("SubscribeToChannel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handlerToBeTested(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot upgrade: %v", err), http.StatusInternalServerError)
	}
	mt, p, err := conn.ReadMessage()
	if err != nil {
		log.Printf("cannot read message: %v", err)
		return
	}
	conn.WriteMessage(mt, p)
}

func TestAPIClient_ReadMessagesFromChannel(t *testing.T) {
	dataCorrect := &Response{
		M:      operationSubscribe,
		Symbol: "bbo:BTC/USDT",
		Data: struct {
			Ts  int64    `json:"ts"`
			Bid []string `json:"bid"`
			Ask []string `json:"ask"`
		}{
			Ts: 11111111,
			Bid: []string{"1",
				"2"},
			Ask: []string{"3",
				"4"},
		},
	}

	resultCorrect := api.BestOrderBook{
		Ask: api.Order{
			Amount: 3,
			Price:  4,
		},
		Bid: api.Order{
			Amount: 1,
			Price:  2,
		},
	}

	type args struct {
		ch chan api.BestOrderBook
	}
	tests := []struct {
		name   string
		args   args
		result api.BestOrderBook
		data   *Response
	}{
		{
			name: "correct BestOrderBook",
			args: args{
				ch: make(chan api.BestOrderBook),
			},
			result: resultCorrect,
			data:   dataCorrect,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(handlerToBeTested))
			u, _ := url.Parse(srv.URL)
			u.Scheme = "ws"
			conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				log.Fatalf("cannot make websocket connection: %v", err)
			}

			subBytes, err := json.Marshal(tt.data)
			if err != nil {
				log.Fatalf("cannot marshal dataCorrect: %v", err)
			}

			err = conn.WriteMessage(websocket.TextMessage, subBytes)
			if err != nil {
				log.Fatalf("cannot write message: %v", err)
			}

			c := &APIClient{
				conn: conn,
			}
			c.ReadMessagesFromChannel(tt.args.ch)

			assert.Equal(t, tt.result, <-tt.args.ch)

		})
	}
}
