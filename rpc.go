package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.uber.org/atomic"
)

var (
	ErrRpcTimeout    = errors.New("rpc timeout")
	ErrNotConnected  = errors.New("not connected")
	ErrRpcChanIsFull = errors.New("rpc chan is full")
)

type rpc struct {
	Id        int64         //
	CreatedAt time.Time     //
	Timeout   time.Duration //
	DoneCh    chan struct{} //
	isDone    *atomic.Bool  //
	Handler   string        // handler name
	Request   []byte        // request for rpc
	Response  []byte        // response from rpc
	Error     error         // rpc error or network error
}

func (rpc *rpc) Ok(payload []byte) {
	rpc.Response = payload
	rpc.isDone.Store(true)
	rpc.DoneCh <- struct{}{}
}

func (rpc *rpc) Err(err error) {
	rpc.Error = err
	rpc.isDone.Store(true)
	rpc.DoneCh <- struct{}{}
}

func (rpc *rpc) Result() ([]byte, error) {
	return rpc.Response, rpc.Error
}

func (rpc *rpc) IsDone() bool {
	return rpc.isDone.Load()
}

type packetType byte

const (
	packetTypePing packetType = iota + 1
	packetTypePong
	packetTypeRequest
	packetTypeResponse
)

type packet struct {
	Id           int64      `json:"id"`
	Type         packetType `json:"type"`
	Handler      string     `json:"handler"`
	RpcTimeoutMs int64      `json:"rpc_timeout_ms"`
	RpcError     string     `json:"rpc_error"`
	Payload      []byte     `json:"payload"`
}

type Conn interface {
	Read() ([]byte, error)
	Write([]byte) error
	Close() error
}

type Handler func(ctx context.Context, request []byte) (response []byte, err error)

type CloseHandler func(closeError error)

type RpcConn struct {
	logger Logger

	conn      Conn
	isClosed  *atomic.Bool
	closeHook chan error
	ctx       context.Context
	cancel    context.CancelFunc

	inPktCh  chan *packet
	outPktCh chan *packet

	// inRpcCh  chan *rpc
	outRpcCh chan *rpc

	idx  *atomic.Int64
	rpcs threadSafeMap[int64, *rpc]

	handlers     map[string]Handler
	closeHandler CloseHandler
}

func NewRpcConn(c Conn, logger Logger, handlers map[string]Handler, closeHandler CloseHandler) *RpcConn {
	if logger == nil {
		logger = fakeLogger{}
	}

	ctx, cancel := context.WithCancel(context.Background())

	conn := &RpcConn{
		logger: logger,

		conn:      c,
		isClosed:  atomic.NewBool(false),
		closeHook: make(chan error, 1),
		ctx:       ctx,
		cancel:    cancel,

		inPktCh:  make(chan *packet, 10),
		outPktCh: make(chan *packet, 10),

		// inRpcCh:      make(chan *rpc, 10),
		outRpcCh: make(chan *rpc, 10),

		idx:  atomic.NewInt64(0),
		rpcs: newtsmap[int64, *rpc](),

		handlers:     map[string]Handler{},
		closeHandler: func(closeError error) {},
	}

	if handlers != nil {
		conn.handlers = handlers
	}

	if closeHandler != nil {
		conn.closeHandler = closeHandler
	}

	// go conn.deleteOutdatedRpcs()
	go conn.readIncomingPackets()
	go conn.runReactor()

	return conn
}

func (c *RpcConn) runReactor() {
	deleteOutdatedTicker := time.NewTicker(time.Second * 2)

	for {
		select {
		case outRpc := <-c.outRpcCh:
			outPkt := c.packetizeRpc(outRpc)
			if c.isClosed.Load() {
				outRpc.Err(ErrNotConnected)
				continue
				// FIXME
				// return
			}
			c.writePacket(outPkt)

		case outPkt := <-c.outPktCh:
			if c.isClosed.Load() {
				continue
				// FIXME
				// return
			}
			c.writePacket(outPkt)

		case inPkt := <-c.inPktCh:
			c.depacketize(inPkt)

		case <-deleteOutdatedTicker.C:
			outdated := []int64{}
			now := time.Now()
			c.rpcs.ForEach(func(rpc *rpc) {
				if rpc.isDone.Load() {
					outdated = append(outdated, rpc.Id)
					c.logger.Debug(fmt.Sprintf("rpc %d is done", rpc.Id))
					return
				}

				if rpc.Timeout != 0 {
					if rpc.CreatedAt.Add(rpc.Timeout).Before(now) {
						rpc.Err(ErrRpcTimeout)
						c.logger.Debug(fmt.Sprintf("rpc %d timeout", rpc.Id))
						outdated = append(outdated, rpc.Id)
					}
				}
			})
			c.rpcs.DeleteMultiple(outdated...)

		case closeError := <-c.closeHook:
			c.isClosed.Store(true)
			c.cancel()

			c.rpcs.ForEach(func(rpc *rpc) {
				rpc.Err(fmt.Errorf("connection is closed"))
			})
			c.rpcs.Flush()

			if c.conn != nil {
				c.conn.Close()
			}

			if c.closeHandler != nil {
				go c.closeHandler(closeError)
			}

			return
		}
	}
}

func (c *RpcConn) Call(ctx context.Context, handler string, request []byte) ([]byte, error) {
	rpc := c.newRpc(ctx, handler, request)

	c.rpcs.Set(rpc.Id, rpc)
	defer c.rpcs.Delete(rpc.Id)

	select {
	case c.outRpcCh <- rpc:
	default:
		return []byte{}, ErrRpcChanIsFull
	}

	select {
	case <-ctx.Done():
		return []byte{}, ctx.Err()
	case <-rpc.DoneCh:
		c.rpcs.Delete(rpc.Id)
		return rpc.Result()
	}
}

func (c *RpcConn) newRpc(ctx context.Context, handler string, request []byte) *rpc {
	return &rpc{
		Id:        c.idx.Inc(),
		CreatedAt: time.Now(),
		Timeout:   getTimeoutFromContext(ctx),
		DoneCh:    make(chan struct{}, 1),
		isDone:    atomic.NewBool(false),
		Handler:   handler,
		Request:   request,
		Response:  []byte{},
		Error:     nil,
	}
}

func (c *RpcConn) packetizeRpc(rpc *rpc) *packet {
	return &packet{
		Id:           rpc.Id,            //
		Type:         packetTypeRequest, //
		Handler:      rpc.Handler,       //
		RpcTimeoutMs: 0,                 // not used in request
		RpcError:     "",                // not used in request
		Payload:      rpc.Request,       //
	}
}

func (c *RpcConn) packetizeResponse(inPkt *packet, response []byte, err error) *packet {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}

	outPkt := &packet{
		Id:       inPkt.Id,
		Type:     packetTypeResponse,
		Handler:  inPkt.Handler,
		RpcError: errStr,
		Payload:  response,
	}

	return outPkt
}

func (c *RpcConn) depacketize(inPkt *packet) {
	switch inPkt.Type {
	case packetTypeRequest:
		handler, ok := c.handlers[inPkt.Handler]
		if !ok {
			errMsg := "no handler for: " + inPkt.Handler
			c.logger.Warn(errMsg)
			c.outPktCh <- c.packetizeResponse(inPkt, []byte{}, errors.New(errMsg))
			return
		}
		go func() {
			var (
				ctx    context.Context    = context.Background()
				cancel context.CancelFunc = func() {}
			)

			if inPkt.RpcTimeoutMs != 0 {
				ctx, cancel = context.WithTimeout(ctx, time.Millisecond*time.Duration(inPkt.RpcTimeoutMs))
			}

			defer cancel()

			response, err := handler(ctx, inPkt.Payload)
			c.outPktCh <- c.packetizeResponse(inPkt, response, err)
		}()

	case packetTypeResponse:
		rpc, ok := c.rpcs.Pop(inPkt.Id)
		if !ok {
			c.logger.Debug(fmt.Sprintf("rpc %d not found", inPkt.Id))
			return
		}
		if inPkt.RpcError != "" {
			rpc.Err(errors.New(inPkt.RpcError))
		} else {
			rpc.Ok(inPkt.Payload)
		}
	}
}

func (c *RpcConn) readIncomingPackets() {
	for {
		inPkt, err := c.readPacket()
		if err != nil {
			c.closeHook <- err
			return
		}
		c.inPktCh <- inPkt
	}
}

func (c *RpcConn) writePacket(pkt *packet) {
	c.writeData(serializePacket(pkt))
}

func serializePacket(pkt *packet) []byte {
	data, _ := json.Marshal(pkt)
	return data
}

func (c *RpcConn) writeData(data []byte) error {
	c.logger.Debug(fmt.Sprintf("write: %s", string(data)))
	if err := c.conn.Write(data); err != nil {
		c.closeHook <- err
		c.logger.Error(fmt.Sprintf("failed to write data: %s", err.Error()))
		return err
	}
	return nil
}

func (c *RpcConn) readPacket() (*packet, error) {
	pkt := packet{}
	data, err := c.readData()
	if err != nil {
		return &pkt, err
	}
	if err := json.Unmarshal(data, &pkt); err != nil {
		return &pkt, err
	}
	return &pkt, err
}

func (c *RpcConn) readData() ([]byte, error) {
	data, err := c.conn.Read()
	if err != nil {
		return []byte{}, err
	}
	c.logger.Debug(fmt.Sprintf("read: %s", string(data)))
	return data, nil
}

func (c *RpcConn) Close() {
	c.closeHook <- nil
}

func (c *RpcConn) IsClosed() bool {
	return c.isClosed.Load()
}
