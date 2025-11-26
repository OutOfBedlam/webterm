package webexec

import (
	"os"
	"os/exec"

	"github.com/OutOfBedlam/webterm"
	"github.com/creack/pty"
)

var _ webterm.Runner = (*WebExec)(nil)

type WebExec struct {
	Command string
	Args    []string
	Dir     string

	cmd *exec.Cmd
	tty *os.File
}

func (wt *WebExec) Open() error {
	wt.cmd = exec.Command(wt.Command, wt.Args...)
	wt.cmd.Dir = wt.Dir

	if tty, err := pty.Start(wt.cmd); err != nil {
		return err
	} else {
		wt.tty = tty
	}
	return nil
}

func (we *WebExec) Close() error {
	if we.cmd != nil {
		if we.cmd.Process != nil {
			we.cmd.Process.Kill()
		}
		we.cmd.Wait()
	}
	if we.tty != nil {
		we.tty.Close()
		we.tty = nil
	}
	return nil
}

func (we *WebExec) Read(p []byte) (n int, err error) {
	return we.tty.Read(p)
}

func (we *WebExec) Write(p []byte) (n int, err error) {
	return we.tty.Write(p)
}

func (we *WebExec) SetWinSize(cols int, rows int) error {
	return pty.Setsize(we.tty, &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)})
}
