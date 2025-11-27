package webtail

import (
	"crypto/md5"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"io"
	"log/slog"
	"regexp"
	"strings"

	"github.com/OutOfBedlam/webterm"
)

var _ webterm.Runner = (*WebTail)(nil)
var _ webterm.Session = (*WebTailSession)(nil)
var _ webterm.ExtSession = (*WebTailSession)(nil)

type WebTail struct {
	Tails []TailConfig
}

func (wt *WebTail) Session() (webterm.Session, error) {
	return &WebTailSession{
		configs: wt.Tails,
	}, nil
}

var tmpl *template.Template

//go:embed webtail.html
var webtailHTML string

type ControlBar struct {
	Hide       bool   `json:"hide"`
	FontSize   int    `json:"fontSize,omitempty"`
	FontFamily string `json:"fontFamily,omitempty"`
}

type FileEntry struct {
	Label string `json:"label"`
	ID    string `json:"id"`
}

func (wt *WebTail) Template() (*template.Template, any) {
	if tmpl == nil {
		tmpl = template.Must(template.New("webtail").Parse(webtailHTML))
	}
	files := []FileEntry{}
	for _, t := range wt.Tails {
		sum := md5.Sum([]byte(t.Filename))
		id := string(hex.EncodeToString(sum[:]))
		files = append(files, FileEntry{Label: webterm.StripAnsiCodes(t.Label), ID: id})
	}
	data := map[string]any{
		"ControlBar": ControlBar{
			Hide:       false,
			FontSize:   12,
			FontFamily: "monospace",
		},
		"Files": files,
	}
	return tmpl, data
}

type TailConfig struct {
	Filename   string
	Label      string
	Highlights []string
}

type WebTailSession struct {
	configs []TailConfig
	tailer  ITail
	tails   []*Tail
	filter  func(string) bool
}

var labelColors = []string{
	webterm.ColorPurple,
	webterm.ColorMagenta,
	webterm.ColorYellow,
	webterm.ColorCyan,
	webterm.ColorGreen,
	webterm.ColorRed,
}

func (wts *WebTailSession) Open() error {
	wts.tails = wts.tails[:0]
	for i, tc := range wts.configs {
		wts.tails = append(wts.tails, NewTail(
			tc.Filename,
			WithLabel(webterm.Colorize(tc.Label, labelColors[i%len(labelColors)])),
			WithSyntaxHighlighting(tc.Highlights...),
		))
	}
	if len(wts.tails) == 0 {
		return nil
	}
	if len(wts.tails) == 1 {
		wts.tailer = wts.tails[0]
	} else {
		var iTails []ITail
		for _, t := range wts.tails {
			iTails = append(iTails, t)
		}
		wts.tailer = NewMultiTail(iTails...)
	}
	if err := wts.tailer.Start(); err != nil {
		return err
	}
	return nil
}

func (wts *WebTailSession) Close() error {
	if wts.tailer != nil {
		wts.tailer.Stop()
	}
	return nil
}

func (wts *WebTailSession) Read(p []byte) (n int, err error) {
	var line string
	for {
		if ln, ok := <-wts.tailer.Lines(); ok {
			line = ln
		} else {
			return 0, io.EOF
		}
		if wts.filter == nil || wts.filter(line) {
			break
		}
	}
	line += "\r\n"
	return copy(p, line), nil
}

func (wts *WebTailSession) SetFilter(filter string) {
	if filter == "" {
		wts.filter = nil
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
	wts.filter = func(line string) bool {
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

func (wts *WebTailSession) Write(p []byte) (n int, err error) {
	// No-op
	return len(p), nil
}

func (wts *WebTailSession) SetWinSize(cols, rows int) error {
	return nil
}

type ExtMessage struct {
	Filter string   `json:"filter"`
	Files  []string `json:"files"`
}

func (wts *WebTailSession) ExtMessage(data []byte) {
	m := ExtMessage{}
	if err := json.Unmarshal(data, &m); err != nil {
		slog.Error("webtail failed to unmarshal ext message", "error", err)
		return
	}
	wts.SetFilter(m.Filter)
	for _, t := range wts.tails {
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
