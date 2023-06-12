package adapter

import (
	"errors"

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
	msgType, msgBin, err := a.wsconn.ReadMessage()
	if err != nil {
		return []byte{}, err
	}
	if msgType != websocket.TextMessage {
		return []byte{}, errors.New("unsupported message type")
	}
	return msgBin, nil
}

func (a *GorillaAdapter) Write(data []byte) error {
	return a.wsconn.WriteMessage(websocket.TextMessage, data)
}

func (a *GorillaAdapter) Close() error {
	return a.wsconn.Close()
}
