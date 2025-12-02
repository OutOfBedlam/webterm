package webport

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

type Addr struct {
	Network string
	Host    string
	Port    int
	Path    string
}

// Parse address with optional path using regexp
// Format: network://host:port[/path]
var parseAddrRegexp = regexp.MustCompile(`^(\w+)://([^:/]+):(\d+)(/.*)?$`)

func ParseAddr(addr string) (*Addr, error) {
	matches := parseAddrRegexp.FindStringSubmatch(addr)

	if matches == nil {
		return nil, fmt.Errorf("invalid address format: %s", addr)
	}

	port, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, fmt.Errorf("invalid port: %s", matches[3])
	}

	return &Addr{
		Network: matches[1],
		Host:    matches[2],
		Port:    port,
		Path:    matches[4],
	}, nil
}

func NewAddr(network string, host string, port int, path string) *Addr {
	return &Addr{
		Network: network,
		Host:    host,
		Port:    port,
		Path:    path,
	}
}

func (a Addr) String() string {
	return fmt.Sprintf("%s://%s:%d%s", a.Network, a.Host, a.Port, a.Path)
}

func (a Addr) Listen() (net.Listener, error) {
	switch a.Network {
	case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6":
		return net.Listen(a.Network, fmt.Sprintf("%s:%d", a.Host, a.Port))
	default:
		return nil, fmt.Errorf("unsupported network type: %s", a.Network)
	}
}

func (a Addr) Dial() (net.Conn, error) {
	switch a.Network {
	case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6":
		return net.Dial(a.Network, net.JoinHostPort(a.Host, fmt.Sprintf("%d", a.Port)))
	case "ws", "wss":
		var header http.Header
		ws, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("%s://%s:%d%s", a.Network, a.Host, a.Port, a.Path), header)
		if err != nil {
			return nil, err
		}
		return ws.NetConn(), nil
	default:
		return nil, fmt.Errorf("unsupported network type: %s", a.Network)
	}
}

// PumpBiDirectional pumps data bi-directionally between localConn and remoteConn.
// It stops when either direction is done or when the done channel is closed.
func PumpBiDirectional(localConn net.Conn, remoteConn net.Conn, done ...<-chan struct{}) {
	remoteToLocalDone := make(chan struct{})
	localToRemoteDone := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		io.Copy(localConn, remoteConn)
		slog.Debug("remote to local copy done", "client", localConn.RemoteAddr())
		close(remoteToLocalDone)
	}()
	go func() {
		defer wg.Done()
		io.Copy(remoteConn, localConn)
		slog.Debug("local to remote copy done", "client", localConn.RemoteAddr())
		close(localToRemoteDone)
	}()

	<-OrChans(append([]<-chan struct{}{remoteToLocalDone, localToRemoteDone}, done...)...)
	remoteConn.Close()
	localConn.Close()

	wg.Wait()
}

// OrChans returns a channel that is closed when any of the provided channels are closed.
func OrChans[T any](doneChans ...<-chan T) <-chan T {
	switch len(doneChans) {
	case 0:
		return nil
	case 1:
		return doneChans[0]
	}

	orDone := make(chan T)
	go func() {
		defer close(orDone)
		switch len(doneChans) {
		case 2:
			select {
			case <-doneChans[0]:
			case <-doneChans[1]:
			}
		default:
			select {
			case <-doneChans[0]:
			case <-doneChans[1]:
			case <-doneChans[2]:
			case <-OrChans(append(doneChans[3:], orDone)...):
			}
		}
	}()
	return orDone
}
