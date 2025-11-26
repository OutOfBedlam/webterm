package webterm

import (
	"embed"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"text/template"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

type WebTerm struct {
	runner          Runner
	fsServer        http.Handler
	cutPrefix       string
	terminalOptions TerminalOptions
}

type Option func(*WebTerm)

func WithCutPrefix(cutPrefix string) Option {
	return func(wt *WebTerm) {
		wt.cutPrefix = cutPrefix
	}
}

func WithTheme(theme TerminalTheme) Option {
	return func(wt *WebTerm) {
		wt.terminalOptions.Theme = theme
	}
}

func New(runner Runner, opts ...Option) *WebTerm {
	wt := &WebTerm{
		runner:          runner,
		fsServer:        http.FileServerFS(staticFS),
		terminalOptions: DefaultTerminalOptions(),
	}
	for _, opt := range opts {
		opt(wt)
	}
	if !strings.HasSuffix(wt.cutPrefix, "/") && wt.cutPrefix != "" {
		wt.cutPrefix += "/"
	}
	return wt
}

//go:embed static/*
var staticFS embed.FS

var tmplIndex *template.Template

func (wt *WebTerm) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if tmplIndex == nil {
		if b, err := staticFS.ReadFile("static/index.html"); err != nil {
			http.Error(w, "Failed to read index.html", http.StatusInternalServerError)
			return
		} else {
			tmplIndex = template.Must(template.New("index").Parse(string(b)))
		}
	}
	path := strings.TrimPrefix(r.URL.Path, wt.cutPrefix)
	defer func() {
		if e := recover(); e != nil {
			slog.Error("panic recovered", "error", e)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}()
	switch path {
	case "":
		err := tmplIndex.Execute(w, wt.dataMap())
		if err != nil {
			http.Error(w, "Failed to render index.html", http.StatusInternalServerError)
		}
	case "data":
		WsDataHandle(wt.runner)(w, r)
	default:
		r.URL.Path = "static/" + path
		wt.fsServer.ServeHTTP(w, r)
	}
}

func (wt *WebTerm) dataMap() TemplateData {
	return TemplateData{
		Terminal: wt.terminalOptions,
	}
}

func pumpStdin(ws *websocket.Conn, runner Runner) {
	defer func() {
		ws.Close()
		runner.Close()
	}()
	ws.SetReadLimit(8192)
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			slog.Error("failed to read from websocket", "error", err)
			break
		}
		if len(message) == 0 {
			continue
		}
		op := message[0]
		data := message[1:]
		switch op {
		case 0: // Resize message
			sz := pty.Winsize{}
			if err := json.Unmarshal(data, &sz); err != nil {
				slog.Error("failed to unmarshal resize message", "error", err)
				continue
			}
			runner.SetWinSize(sz.Cols, sz.Rows)
		case 1: // Data message
			_, err = runner.Write(data)
			if err != nil {
				slog.Error("failed to write to stdin", "error", err)
				return
			}
		}
	}
}

func pumpStdout(ws *websocket.Conn, runner Runner) {
	defer ws.Close()
	buffer := make([]byte, 8192)
	for {
		n, err := runner.Read(buffer)
		if err != nil {
			slog.Error("failed to read from runner", "error", err)
			break
		}
		err = ws.WriteMessage(websocket.BinaryMessage, buffer[:n])
		if err != nil {
			slog.Error("failed to write to websocket", "error", err)
			break
		}
	}
}

func WsDataHandle(runner Runner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			http.Error(w, "Failed to upgrade to websocket", http.StatusInternalServerError)
		}
		defer conn.Close()

		if err := runner.Open(); err != nil {
			slog.Error("failed to run", "error", err)
			http.Error(w, "Failed to start terminal", http.StatusInternalServerError)
			return
		}
		defer runner.Close()

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			pumpStdout(conn, runner)
		}()
		pumpStdin(conn, runner)
		wg.Wait()
		slog.Info("webterm data closed")
	}
}

type Runner interface {
	Open() error
	Close() error
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	SetWinSize(cols, rows uint16) error
}
