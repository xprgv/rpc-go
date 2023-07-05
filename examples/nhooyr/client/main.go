package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/exp/slog"
	"nhooyr.io/websocket"

	"github.com/xprgv/rpc-go"
	"github.com/xprgv/rpc-go/pkg/adapter"
)

const websocketAddress = "ws://localhost:3000"

func main() {
	slog.SetDefault(
		slog.New(
			slog.NewTextHandler(
				os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo},
			),
		),
	)

	wsconn, _, err := websocket.Dial(context.Background(), websocketAddress, nil)
	if err != nil {
		slog.Error("Failed to connect", slog.String("error", err.Error()))
		os.Exit(1)
	}
	conn := adapter.NewNhooyrAdapter(wsconn)

	rpcConn := rpc.NewRpcConn(conn, slog.Default(), map[string]rpc.Handler{}, func(closeError error) {})
	defer rpcConn.Close()

	go func() {
		for {
			response, err := rpcConn.Call(context.Background(), "ping-handler", []byte("ping"))
			if err != nil {
				slog.Error("Failed to call rpc",
					slog.String("error", err.Error()),
					slog.String("rpc", "ping-handler"),
				)
				os.Exit(1)
			}
			fmt.Println(string(response))
			time.Sleep(time.Second * 2)
		}
	}()

	select {}
}
