package adapter

import (
	"fmt"

	"github.com/gorilla/websocket"

	"github.com/xprgv/rpc-go"
)

var _ rpc.Conn = (*GorillaAdapter)(nil)

type GorillaAdapter struct {
	wsconn *websocket.Conn
}

func NewGorillaAdapter(wsconn *websocket.Conn) *GorillaAdapter {
	return &GorillaAdapter{
		wsconn: wsconn,
	}
}

func (a *GorillaAdapter) Read() ([]byte, error) {
	msgType, binMsg, err := a.wsconn.ReadMessage()
	if err != nil {
		return []byte{}, err
	}
	if msgType != websocket.BinaryMessage {
		return []byte{}, fmt.Errorf("unsupported message type")
	}
	return binMsg, nil
}

func (a *GorillaAdapter) Write(data []byte) error {
	return a.wsconn.WriteMessage(websocket.BinaryMessage, data)
}

func (a *GorillaAdapter) Close() error {
	return a.wsconn.Close()
}
