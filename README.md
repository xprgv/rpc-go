# rpc-go

Minimalistic golang library that allows you to create a rpc client/server on top of any bidirectional connections

## Installation

```sh
go get -u github.com/xprgv/rpc-go
```

## Usage

nhooyr.io/websocket server:

```go
http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    wsconn, err := websocket.Accept(w, r, nil)
    if err != nil {
        log.Fatal(err)
    }

    go func() {
        rpcConn := rpc.NewRpcConn(
            adapter.NewNhooyrAdapter(wsconn),
            logger,
            map[string]rpc.Handler{
                "ping-handler": func(ctx context.Context, request []byte) (response []byte, err error) {
                    return []byte("pong"), nil
                },
            },
            func(closeError error) {
                fmt.Println("rpc connection closed", err.Error())
            },
        )
        defer rpcConn.Close()
        select {}
    }()
}

```

nhooyr.io/websocket client:

```go
wsconn, _, err := websocket.Dial(context.Background(), websocketAddress, nil)
if err != nil {
    log.Fatal(err)
}

rpcConn := rpc.NewRpcConn(adapter.NewNhooyrAdapter(wsconn), logger, map[string]rpc.Handler{}, func(closeError error) {})
defer rpcConn.Close()

response, err := rpcConn.Call(context.Background(), "ping-handler", []byte("ping"))
if err != nil {
    log.Fatal(err)
}
```
