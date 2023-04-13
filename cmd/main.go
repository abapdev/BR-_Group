package main

import (
	"log"

	"github.com/abapdev/BR_Group/internal/client"
	"github.com/abapdev/BR_Group/package/api"
)

func main() {
	ac := client.NewAPIClient()

	if err := ac.Connection(); err != nil {
		log.Fatalf("Error connecting to websocket: %v", err)
	}

	if err := ac.SubscribeToChannel("BTC/USDT"); err != nil {
		log.Fatalf("Error subscribing to channel: %v", err)
	}

	ac.WriteMessagesToChannel()

	data := make(chan api.BestOrderBook)

	ac.ReadMessagesFromChannel(data)

	for {
		log.Println(<-data)
	}
}
