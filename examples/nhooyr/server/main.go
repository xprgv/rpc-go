package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"nhooyr.io/websocket"

	"github.com/xprgv/rpc-go"
	"github.com/xprgv/rpc-go/pkg/adapter"
)

const address = "localhost:3000"

var rpcHandlers = map[string]rpc.Handler{
	"ping-handler": func(ctx context.Context, request []byte) (response []byte, err error) {
		return []byte("pong"), nil
	},
}

var closeHandler rpc.CloseHandler = func(closeError error) {}

func main() {
	logger := logrus.New()
	logger.Level = logrus.InfoLevel

	websocketHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wsconn, err := websocket.Accept(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		go func() {
			rpcConn := rpc.NewRpcConn(adapter.NewNhooyrAdapter(wsconn), logger, rpcHandlers, closeHandler)
			defer rpcConn.Close()
			select {}
		}()
	})

	http.ListenAndServe(address, websocketHandler)
}
