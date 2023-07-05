package adapter

import (
	"net"

	"github.com/xprgv/rpc-go"
)

var _ rpc.Conn = (*TcpSocketAdapter)(nil)

type TcpSocketAdapter struct {
	conn *net.TCPConn
	buf  []byte
}

func NewTcpSocketAdapter(conn *net.TCPConn, bufSizeBytes int) *TcpSocketAdapter {
	return &TcpSocketAdapter{
		conn: conn,
		buf:  make([]byte, bufSizeBytes),
	}
}

func (s *TcpSocketAdapter) Read() ([]byte, error) {
	n, err := s.conn.Read(s.buf)
	if err != nil {
		return []byte{}, err
	}
	return s.buf[:n], nil
}

func (s *TcpSocketAdapter) Write(data []byte) error {
	_, err := s.conn.Write(data)
	return err
}

func (s *TcpSocketAdapter) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
