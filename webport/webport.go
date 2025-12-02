package webport

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Config struct {
	LocalAddr  string
	RemoteAddr string
}

func New(cfg Config) *WebPort {
	return &WebPort{
		conf: cfg,
		done: make(chan struct{}),
	}
}

type WebPort struct {
	conf       Config
	localAddr  *Addr
	remoteAddr *Addr
	lsnr       net.Listener
	err        error
	done       chan struct{}
	wg         sync.WaitGroup
}

func (wp *WebPort) Start() error {
	if addr, err := ParseAddr(wp.conf.LocalAddr); err != nil {
		return fmt.Errorf("invalid local address: %v", err)
	} else {
		wp.localAddr = addr
	}
	if addr, err := ParseAddr(wp.conf.RemoteAddr); err != nil {
		return fmt.Errorf("invalid remote address: %v", err)
	} else {
		wp.remoteAddr = addr
	}
	if lsnr, err := wp.localAddr.Listen(); err != nil {
		return fmt.Errorf("failed to start listener: %v", err)
	} else {
		wp.lsnr = lsnr
	}
	wp.wg.Add(1)
	go func() {
		defer wp.wg.Done()
		for {
			conn, err := wp.lsnr.Accept()
			if err != nil {
				wp.err = err
				return
			}
			if c, ok := conn.(*net.TCPConn); ok {
				c.SetKeepAlive(true)
				c.SetLinger(0)
				c.SetNoDelay(true)
			}
			go wp.handleConnection(conn, wp.remoteAddr)
		}
	}()
	return nil
}

func (wp *WebPort) Stop() error {
	if wp.lsnr != nil {
		wp.lsnr.Close()
	}
	close(wp.done)
	wp.wg.Wait()
	return wp.err
}

// Err returns the error encountered during Serve operation, if any.
// it represents the reason for termination.
// If Serve stopped gracefully, Err returns nil.
func (wp *WebPort) Err() error {
	return wp.err
}

func (wp *WebPort) handleConnection(localConn net.Conn, remoteAddr *Addr) {
	defer localConn.Close()
	remoteConn, err := remoteAddr.Dial()
	if err != nil {
		slog.Error("failed to connect to remote address", "error", err)
		return
	}
	defer remoteConn.Close()

	slog.Debug("connection start", "client", localConn.RemoteAddr(), "remote", wp.remoteAddr.String())
	PumpBiDirectional(localConn, remoteConn, wp.done)
	slog.Debug("connection closed", "client", localConn.RemoteAddr())
}

// HandleHTTP upgrades HTTP connections to WebSocket and pumps data bi-directionally
// between the WebSocket connection and the remote address.
func (wp *WebPort) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	if wp.remoteAddr == nil {
		if addr, err := ParseAddr(wp.conf.RemoteAddr); err != nil {
			slog.Error("server not initialized", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		} else {
			wp.remoteAddr = addr
		}
	}
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade fail", "error", err)
		return
	}
	defer conn.Close()

	remoteConn, err := wp.remoteAddr.Dial()
	if err != nil {
		slog.Error("failed to connect to remote address", "error", err)
		return
	}
	defer remoteConn.Close()

	slog.Debug("connection start", "client", conn.RemoteAddr(), "remote", wp.remoteAddr.String())
	PumpBiDirectional(conn.NetConn(), remoteConn, r.Context().Done(), wp.done)
	slog.Debug("connection closed", "client", conn.RemoteAddr())
}
