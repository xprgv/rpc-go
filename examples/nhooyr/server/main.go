package main

import (
	"context"
	"net/http"
	"os"

	"golang.org/x/exp/slog"
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
	slog.SetDefault(
		slog.New(
			slog.NewTextHandler(
				os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo},
			),
		),
	)

	websocketHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wsconn, err := websocket.Accept(w, r, nil)
		if err != nil {
			slog.Error("Failed to accept websocket connection", slog.String("error", err.Error()))
			return
		}
		conn := adapter.NewNhooyrAdapter(wsconn)

		go func() {
			rpcConn := rpc.NewRpcConn(
				conn,
				slog.Default(),
				rpcHandlers,
				closeHandler,
			)
			defer rpcConn.Close()
			select {}
		}()
	})

	http.ListenAndServe(address, websocketHandler)
}
