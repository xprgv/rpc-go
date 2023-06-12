package adapter

import (
	"context"
	"errors"

	"nhooyr.io/websocket"

	"github.com/xprgv/rpc-go"
)

var _ rpc.Conn = (*NhooyrAdapter)(nil)

type NhooyrAdapter struct {
	wsconn *websocket.Conn
}

func NewNhooyrAdapter(wsconn *websocket.Conn) *NhooyrAdapter {
	return &NhooyrAdapter{
		wsconn: wsconn,
	}
}

func (a *NhooyrAdapter) Read() ([]byte, error) {
	messageType, binMsg, err := a.wsconn.Read(context.Background())
	if err != nil {
		// fmt.Println("failed to read:", err)
		return []byte{}, err
	}
	if messageType != websocket.MessageText {
		return []byte{}, errors.New("unsupported message type")
	}
	return binMsg, nil
}

func (a *NhooyrAdapter) Write(data []byte) error {
	return a.wsconn.Write(context.Background(), websocket.MessageText, data)
}

func (a *NhooyrAdapter) Close() error {
	return a.wsconn.Close(websocket.StatusNormalClosure, "normal closure")
}
