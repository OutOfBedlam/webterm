# WebPort

WebPort is a Go package that provides bidirectional port forwarding between TCP/UDP and WebSocket protocols. It supports connecting local TCP ports to WebSocket endpoints and vice versa.

## Features

- TCP, UDP, and WebSocket protocol support
- Bidirectional data transfer
- WebSocket upgrade through HTTP handlers
- Simple address parsing and connection management
- Graceful shutdown support

## Installation

```bash
go get github.com/OutOfBedlam/webterm/webport
```

## Usage

### Basic Example

```go
package main

import (
    "github.com/OutOfBedlam/webterm/webport"
)

func main() {
    // Configure WebPort
    cfg := webport.Config{
        LocalAddr:  "tcp://localhost:8080",
        RemoteAddr: "ws://remote.example.com:9000/ws",
    }
    
    // Create WebPort instance
    wp := webport.New(cfg)
    
    // Start server
    if err := wp.Start(); err != nil {
        panic(err)
    }
    defer wp.Stop()
    
    // Server running...
}
```

### Using as HTTP Handler

```go
package main

import (
    "net/http"
    "github.com/OutOfBedlam/webterm/webport"
)

func main() {
    cfg := webport.Config{
        RemoteAddr: "tcp://localhost:22",
    }
    
    wp := webport.New(cfg)
    
    http.HandleFunc("/ws", wp.HandleHTTP)
    http.ListenAndServe(":8080", nil)
}
```

## Address Format

WebPort supports the following address format:

```
network://host:port[/path]
```

### Supported Network Types

- `tcp`, `tcp4`, `tcp6`: TCP connections
- `udp`, `udp4`, `udp6`: UDP connections
- `ws`, `wss`: WebSocket connections (path can be specified)

### Address Examples

```go
"tcp://localhost:8080"
"tcp4://0.0.0.0:3000"
"ws://example.com:9000/websocket"
"wss://secure.example.com:443/ws/connect"
```

## API

### Config

```go
type Config struct {
    LocalAddr  string  // Local address to listen on
    RemoteAddr string  // Remote address to connect to
}
```

### WebPort

#### New(cfg Config) *WebPort

Creates a new WebPort instance.

#### (*WebPort) Start() error

Starts the WebPort server and accepts connections on the local address.

#### (*WebPort) Stop() error

Stops the WebPort server and closes all connections.

#### (*WebPort) Err() error

Returns the error that occurred during server shutdown, if any. Returns nil if shutdown was graceful.

#### (*WebPort) HandleHTTP(w http.ResponseWriter, r *http.Request)

Upgrades the HTTP request to WebSocket and connects to the remote address.

### Addr

#### ParseAddr(addr string) (*Addr, error)

Parses a string address and returns an Addr struct.

#### (*Addr) String() string

Returns the Addr in string format.

#### (*Addr) Listen() (net.Listener, error)

Starts a listener on the address.

#### (*Addr) Dial() (net.Conn, error)

Connects to the address.

### PumpBiDirectional

```go
func PumpBiDirectional(localConn net.Conn, remoteConn net.Conn, done ...<-chan struct{})
```

Pumps data bidirectionally between two connections. Stops when the done channel is closed or when either connection terminates.

## CLI Usage

WebPort can also be used as a standalone CLI tool:

```bash
# Run in server mode
webport serve -l tcp://localhost:8080 -r ws://remote.example.com:9000/ws

# Expose local SSH port via WebSocket
webport serve -l tcp://0.0.0.0:2222 -r tcp://localhost:22
```

### Command Options

- `serve`: Starts the WebPort server
  - `-l`: Local listen address
  - `-r`: Remote connection address

## Use Cases

1. **Expose TCP services via WebSocket**: Expose TCP services behind firewalls through WebSocket
2. **Protocol bridging**: Connect legacy TCP clients to WebSocket servers
3. **Remote port forwarding**: Implement SSH-style port forwarding over WebSocket
4. **Development/Testing**: Test local services remotely

## Dependencies

- [gorilla/websocket](https://github.com/gorilla/websocket): WebSocket implementation
