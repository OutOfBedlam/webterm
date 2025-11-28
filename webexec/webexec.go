package webexec

import (
	"html/template"
	"os"
	"os/exec"

	"github.com/OutOfBedlam/webterm"
	"github.com/creack/pty"
)

var _ webterm.Runner = (*WebExec)(nil)
var _ webterm.Session = (*WebExecSession)(nil)

type WebExec struct {
	Command string
	Args    []string
	Dir     string
}

func (we *WebExec) Session() (webterm.Session, error) {
	return &WebExecSession{WebExec: *we}, nil
}

func (we *WebExec) Template() (*template.Template, any) {
	return nil, nil
}

type WebExecSession struct {
	WebExec
	cmd *exec.Cmd
	tty *os.File
}

func (wes *WebExecSession) Open() error {
	wes.cmd = exec.Command(wes.Command, wes.Args...)
	wes.cmd.Dir = wes.Dir

	if tty, err := pty.Start(wes.cmd); err != nil {
		return err
	} else {
		wes.tty = tty
	}
	return nil
}

func (wes *WebExecSession) Close() error {
	if wes.cmd != nil {
		if wes.cmd.Process != nil {
			wes.cmd.Process.Kill()
		}
		wes.cmd.Wait()
	}
	if wes.tty != nil {
		wes.tty.Close()
		wes.tty = nil
	}
	return nil
}

func (wes *WebExecSession) Read(p []byte) (n int, err error) {
	return wes.tty.Read(p)
}

func (wes *WebExecSession) Write(p []byte) (n int, err error) {
	return wes.tty.Write(p)
}

func (wes *WebExecSession) SetWinSize(cols int, rows int) error {
	return pty.Setsize(wes.tty, &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)})
}

func (wes *WebExecSession) Control(data []byte) error {
	return nil
}
