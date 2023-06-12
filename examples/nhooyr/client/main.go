package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sirupsen/logrus"
	"nhooyr.io/websocket"

	"github.com/xprgv/rpc-go"
	"github.com/xprgv/rpc-go/pkg/adapter"
)

const websocketAddress = "ws://localhost:3000"

func main() {
	logger := logrus.New()
	logger.Level = logrus.InfoLevel

	wsconn, _, err := websocket.Dial(context.Background(), websocketAddress, nil)
	if err != nil {
		log.Fatal(err)
	}

	rpcConn := rpc.NewRpcConn(adapter.NewNhooyrAdapter(wsconn), logger, map[string]rpc.Handler{}, func(closeError error) {})
	defer rpcConn.Close()

	go func() {
		for {
			response, err := rpcConn.Call(context.Background(), "ping-handler", []byte("ping"))
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(response))
			time.Sleep(time.Second * 2)
		}
	}()

	select {}
}
