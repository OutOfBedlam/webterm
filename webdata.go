package webterm

import "encoding/json"

type TemplateData struct {
	Terminal     TerminalOptions
	Localization map[string]string
}

func (td TemplateData) Localize(s string) string {
	if l, ok := td.Localization[s]; ok {
		return l
	}
	return s
}

type TerminalOptions struct {
	CursorBlink         bool          `json:"cursorBlink"`
	CursorInactiveStyle string        `json:"cursorInactiveStyle,omitempty"`
	CursorStyle         string        `json:"cursorStyle,omitempty"`
	FontSize            int           `json:"fontSize,omitempty"`
	FontFamily          string        `json:"fontFamily,omitempty"`
	LineHeight          float64       `json:"lineHeight,omitempty"`
	Theme               TerminalTheme `json:"theme"`
	Scrollback          int           `json:"scrollback,omitempty"`
	DisableStdin        bool          `json:"disableStdin"`
	ConvertEol          bool          `json:"convertEol,omitempty"`
}

func DefaultTerminalOptions() TerminalOptions {
	return TerminalOptions{
		CursorBlink:  true,
		FontSize:     12,
		FontFamily:   `"Monaspace Neon",Menlo,Consolas,ui-monospace,monospace`,
		LineHeight:   1.2,
		Scrollback:   1000,
		Theme:        ThemeDefault,
		DisableStdin: false,
		ConvertEol:   false,
	}
}

func (tt TerminalOptions) String() string {
	opts, _ := json.MarshalIndent(tt, "", "  ")
	return string(opts)
}

type TerminalTheme struct {
	Background                  string `json:"background,omitempty"`
	Foreground                  string `json:"foreground.omitempty"`
	SelectionBackground         string `json:"selectionBackground,omitempty"`
	SelectionForeground         string `json:"selectionForeground,omitempty"`
	SelectionInactiveBackground string `json:"selectionInactiveBackground,omitempty"`
	Cursor                      string `json:"cursor,omitempty"`
	CursorAccent                string `json:"cursorAccent,omitempty"`
	ExtendedAnsi                string `json:"extendedAnsi,omitempty"`
	Black                       string `json:"black,omitempty"`
	Blue                        string `json:"blue,omitempty"`
	BrightBlack                 string `json:"brightBlack,omitempty"`
	BrightBlue                  string `json:"brightBlue,omitempty"`
	BrightCyan                  string `json:"brightCyan,omitempty"`
	BrightGreen                 string `json:"brightGreen,omitempty"`
	BrightMagenta               string `json:"brightMagenta,omitempty"`
	BrightRed                   string `json:"brightRed,omitempty"`
	BrightWhite                 string `json:"brightWhite,omitempty"`
	BrightYellow                string `json:"brightYellow,omitempty"`
	Cyan                        string `json:"cyan,omitempty"`
	Green                       string `json:"green,omitempty"`
	Magenta                     string `json:"magenta,omitempty"`
	Red                         string `json:"red,omitempty"`
	White                       string `json:"white,omitempty"`
	Yellow                      string `json:"yellow,omitempty"`
}
