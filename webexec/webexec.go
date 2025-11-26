package webexec

import (
	"os"
	"os/exec"

	"github.com/OutOfBedlam/webterm"
	"github.com/creack/pty"
)

func New(command string, args []string, workDir string, opts ...webterm.Option) *webterm.WebTerm {
	runner := &WebTermExec{
		Command: command,
		Args:    args,
		WorkDir: workDir,
	}
	return webterm.New(runner, opts...)
}

var _ webterm.Runner = (*WebTermExec)(nil)

type WebTermExec struct {
	Command string
	Args    []string
	WorkDir string

	cmd *exec.Cmd
	tty *os.File
}

func (wt *WebTermExec) Open() error {
	wt.cmd = exec.Command(wt.Command, wt.Args...)
	wt.cmd.Dir = wt.WorkDir

	if tty, err := pty.Start(wt.cmd); err != nil {
		return err
	} else {
		wt.tty = tty
	}
	return nil
}

func (we *WebTermExec) Close() error {
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

func (we *WebTermExec) Read(p []byte) (n int, err error) {
	return we.tty.Read(p)
}

func (we *WebTermExec) Write(p []byte) (n int, err error) {
	return we.tty.Write(p)
}

func (we *WebTermExec) SetWinSize(cols, rows uint16) error {
	return pty.Setsize(we.tty, &pty.Winsize{Cols: cols, Rows: rows})
}
