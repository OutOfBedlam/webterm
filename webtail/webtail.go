package webtail

import (
	"crypto/md5"
	_ "embed"
	"encoding/hex"
	"html/template"
	"log/slog"
	"path/filepath"

	"github.com/OutOfBedlam/webterm"
)

type Tail struct {
	id       string
	filename string
	options  []Option
}

func NewTail(filename string, opts ...Option) *Tail {
	sum := md5.Sum([]byte(filename))
	rt := &Tail{
		filename: filename,
		id:       string(hex.EncodeToString(sum[:])),
		options:  opts,
	}
	return rt
}

var _ webterm.Runner = (*WebTail)(nil)
var _ webterm.ExtRunner = (*WebTail)(nil)

type WebTail struct {
	Tails  []*Tail
	tailer ITail
}

func (wt *WebTail) Open() error {
	if len(wt.Tails) == 0 {
		return nil
	}
	if len(wt.Tails) == 1 {
		wt.tailer = NewSingleTail(wt.Tails[0].filename, wt.Tails[0].options...)
	} else {
		var tails []ITail
		for _, t := range wt.Tails {
			tails = append(tails, NewSingleTail(t.filename, t.options...))
		}
		wt.tailer = NewMultiTail(tails...)
	}
	if err := wt.tailer.Start(); err != nil {
		return err
	}
	return nil
}

func (wt *WebTail) Close() error {
	if wt.tailer != nil {
		wt.tailer.Stop()
	}
	return nil
}

func (wt *WebTail) Read(p []byte) (n int, err error) {
	line := <-wt.tailer.Lines()
	line += "\r\n"
	return copy(p, line), nil
}

func (wt *WebTail) Write(p []byte) (n int, err error) {
	// No-op
	return len(p), nil
}

func (wt *WebTail) SetWinSize(cols, rows int) error {
	return nil
}

var tmpl *template.Template

//go:embed webtail.html
var webtailHTML string

func (wt *WebTail) Template() *template.Template {
	if tmpl == nil {
		tmpl = template.Must(template.New("webtail").Parse(webtailHTML))
	}
	return tmpl
}

type ControlBar struct {
	Hide       bool   `json:"hide"`
	FontSize   int    `json:"fontSize,omitempty"`
	FontFamily string `json:"fontFamily,omitempty"`
}

func (wt *WebTail) ExtData() any {
	files := []string{}
	for _, t := range wt.Tails {
		files = append(files, filepath.Base(t.filename))
	}
	return map[string]any{
		"ControlBar": ControlBar{
			Hide:       false,
			FontSize:   12,
			FontFamily: "monospace",
		},
		"Files": files,
	}
}

func (wt *WebTail) ExtMessage(data []byte) {
	slog.Debug("received ext message", "data", string(data))
}
