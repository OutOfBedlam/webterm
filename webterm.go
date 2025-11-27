package webterm

import (
	"embed"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

type Runner interface {
	Session() (Session, error)
	// provide custom template, and template-data if nil default will be used
	Template() (*template.Template, any)
}

type Session interface {
	Open() error
	Close() error
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	SetWinSize(cols, rows int) error
}

type ExtSession interface {
	ExtMessage(data []byte) // handle extension messages from client
}

type WebTerm struct {
	runner          Runner
	fsServer        http.Handler
	cutPrefix       string
	terminalOptions TerminalOptions
	localization    map[string]string
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

func WithFontFamily(fontFamily string) Option {
	return func(wt *WebTerm) {
		wt.terminalOptions.FontFamily = fontFamily
	}
}

func WithFontSize(fontSize int) Option {
	return func(wt *WebTerm) {
		wt.terminalOptions.FontSize = fontSize
	}
}

func WithScrollback(scrollback int) Option {
	return func(wt *WebTerm) {
		wt.terminalOptions.Scrollback = scrollback
	}
}

func WithLocalization(localization map[string]string) Option {
	return func(wt *WebTerm) {
		wt.localization = localization
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
	path := strings.TrimPrefix(r.URL.Path, wt.cutPrefix)
	defer func() {
		if e := recover(); e != nil {
			slog.Error("panic recovered", "error", e, "path", r.URL.Path)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}()
	switch path {
	case "":
		wt.index(w, r)
	case "data":
		wt.data(w, r)
	default:
		if strings.HasPrefix(path, "index") {
			http.NotFound(w, r)
			return
		}
		r.URL.Path = "static/" + path
		wt.fsServer.ServeHTTP(w, r)
	}
}

func (wt *WebTerm) index(w http.ResponseWriter, _ *http.Request) {
	var tmpl *template.Template
	var extData any
	if extTmpl, data := wt.runner.Template(); extTmpl != nil {
		tmpl = extTmpl
		extData = data
	}
	if tmpl == nil {
		if tmplIndex == nil {
			if b, err := staticFS.ReadFile("static/index.html"); err != nil {
				http.Error(w, "Failed to read index.html", http.StatusInternalServerError)
				return
			} else {
				tmplIndex = template.Must(template.New("index").Parse(string(b)))
			}
		}
		tmpl = tmplIndex
	}
	if tmpl == nil {
		http.Error(w, "Template not provided", http.StatusInternalServerError)
		return
	}
	tmplData := TemplateData{
		Terminal:     wt.terminalOptions,
		Localization: wt.localization,
		Ext:          extData,
	}
	if err := tmpl.Execute(w, tmplData); err != nil {
		http.Error(w, "Failed to render index.html", http.StatusInternalServerError)
	}
}

func (wt *WebTerm) data(w http.ResponseWriter, r *http.Request) {
	session, err := wt.runner.Session()
	if err != nil {
		slog.Error("webterm failed to create runner", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := session.Open(); err != nil {
		slog.Error("webterm failed to run", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer session.Close()

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
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer conn.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		pumpStdout(conn, session)
	}()
	pumpStdin(conn, session)
	wg.Wait()
	slog.Info("webterm data closed")
}

func pumpStdin(ws *websocket.Conn, runner Session) {
	defer func() {
		ws.Close()
		runner.Close()
	}()
	ws.SetReadLimit(8192)
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			slog.Error("webterm failed to read from websocket", "error", err)
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
				slog.Error("webterm failed to unmarshal resize message", "error", err)
				continue
			}
			runner.SetWinSize(int(sz.Cols), int(sz.Rows))
		case 1: // Data message
			_, err = runner.Write(data)
			if err != nil {
				slog.Error("webterm failed to write to runner", "error", err)
				return
			}
		case 2: // Ext message
			if extRun, ok := runner.(ExtSession); ok {
				extRun.ExtMessage(data)
			}
		}
	}
}

func pumpStdout(ws *websocket.Conn, runner Session) {
	defer ws.Close()
	buffer := make([]byte, 8192)
	for {
		n, err := runner.Read(buffer)
		if err != nil {
			slog.Error("webterm failed to read from runner", "error", err)
			break
		}
		err = ws.WriteMessage(websocket.BinaryMessage, buffer[:n])
		if err != nil {
			slog.Error("webterm failed to write to websocket", "error", err)
			break
		}
	}
}
