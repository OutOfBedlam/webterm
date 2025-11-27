package webtail

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"io"
	"log/slog"
	"regexp"
	"strings"

	"github.com/OutOfBedlam/webterm"
)

var _ webterm.Runner = (*WebTail)(nil)
var _ webterm.ExtRunner = (*WebTail)(nil)

type WebTail struct {
	Tails  []*Tail
	tailer ITail
	filter func(string) bool
}

func (wt *WebTail) Open() error {
	if len(wt.Tails) == 0 {
		return nil
	}
	if len(wt.Tails) == 1 {
		wt.tailer = wt.Tails[0]
	} else {
		var tails []ITail
		for _, t := range wt.Tails {
			tails = append(tails, t)
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
	var line string
	for {
		if ln, ok := <-wt.tailer.Lines(); ok {
			line = ln
		} else {
			return 0, io.EOF
		}
		if wt.filter == nil || wt.filter(line) {
			break
		}
	}
	line += "\r\n"
	return copy(p, line), nil
}

func (wt *WebTail) SetFilter(filter string) {
	if filter == "" {
		wt.filter = nil
		return
	}
	filters := strings.Split(filter, "||")
	var patterns []Pattern
	for _, filter := range filters {
		splits := strings.Split(filter, "&&")
		var pattern Pattern
		for _, tok := range splits {
			tok = strings.TrimSpace(tok)
			if tok != "" {
				reg, err := regexp.Compile(tok)
				if err != nil {
					slog.Error("webtail invalid filter", "pattern", tok, "error", err)
					return
				}
				pattern = append(pattern, reg)
			}
		}
		patterns = append(patterns, pattern)
	}
	wt.filter = func(line string) bool {
		for _, pattern := range patterns {
			matched := true
			for _, reg := range pattern {
				if !reg.MatchString(line) {
					matched = false
					break
				}
			}
			if matched {
				return true
			}
		}
		return false
	}
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

func (wt *WebTail) ExtTemplate() *template.Template {
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

type FileEntry struct {
	Label string `json:"label"`
	ID    string `json:"id"`
}

func (wt *WebTail) ExtData() any {
	files := []FileEntry{}
	for _, t := range wt.Tails {
		files = append(files, FileEntry{Label: webterm.StripAnsiCodes(t.label), ID: t.id})
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

type ExtMessage struct {
	Filter string   `json:"filter"`
	Files  []string `json:"files"`
}

func (wt *WebTail) ExtMessage(data []byte) {
	m := ExtMessage{}
	if err := json.Unmarshal(data, &m); err != nil {
		slog.Error("webtail failed to unmarshal ext message", "error", err)
		return
	}
	wt.SetFilter(m.Filter)
	for _, t := range wt.Tails {
		include := false
		for _, fid := range m.Files {
			if t.id == fid {
				include = true
				break
			}
		}
		t.SetSilent(!include)
	}
}
