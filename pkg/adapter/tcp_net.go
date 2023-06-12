package adapter

import (
	"net"

	"github.com/xprgv/rpc-go"
)

var _ rpc.Conn = (*TcpConnAdapter)(nil)

type TcpConnAdapter struct {
	conn *net.TCPConn
	buf  []byte
}

func NewTcpConnAdapter(conn *net.TCPConn, bufSize int) *TcpConnAdapter {
	return &TcpConnAdapter{
		conn: conn,
		buf:  make([]byte, bufSize),
	}
}

func (a *TcpConnAdapter) Read() ([]byte, error) {
	n, err := a.conn.Read(a.buf)
	if err != nil {
		return []byte{}, err
	}
	return a.buf[:n], nil
}

func (a *TcpConnAdapter) Write(data []byte) error {
	_, err := a.conn.Write(data)
	return err
}

func (a *TcpConnAdapter) Close() error {
	return a.conn.Close()
}
