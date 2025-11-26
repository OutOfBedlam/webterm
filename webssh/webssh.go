package webssh

import (
	"fmt"
	"io"
	"os"

	"github.com/OutOfBedlam/webterm"
	"golang.org/x/crypto/ssh"
)

var _ webterm.Runner = (*WebSSH)(nil)

type WebSSH struct {
	Network  string
	Host     string
	Port     int
	User     string
	Auth     []ssh.AuthMethod
	TermType string
	Command  string

	conn    *ssh.Client
	session *ssh.Session
	reader  io.Reader
	writer  io.Writer
}

func (ws *WebSSH) Open() error {
	if ws.Network == "" {
		ws.Network = "tcp"
	}
	if ws.Port == 0 {
		ws.Port = 22
	}
	if ws.User == "" {
		ws.User = os.Getenv("USER")
	}
	if ws.TermType == "" {
		ws.TermType = "xterm"
	}
	conn, err := ssh.Dial(ws.Network, fmt.Sprintf("%s:%d", ws.Host, ws.Port), &ssh.ClientConfig{
		User:            ws.User,
		Auth:            ws.Auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return err
	}
	ws.conn = conn

	session, err := conn.NewSession()
	if err != nil {
		return err
	}
	ws.session = session

	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return err
	}
	stdin, err := session.StdinPipe()
	if err != nil {
		return err
	}

	ws.reader = io.MultiReader(stdout, stderr)
	ws.writer = stdin

	err = session.RequestPty(ws.TermType, 40, 80, ssh.TerminalModes{
		ssh.ECHO: 1, // enable echoing
	})
	if err != nil {
		session.Close()
		conn.Close()
		return err
	}

	if ws.Command != "" {
		err = ws.session.Start(ws.Command)
	} else {
		err = ws.session.Shell()
	}
	if err != nil {
		ws.session.Close()
		conn.Close()
		return err
	}

	return nil
}

func (ws *WebSSH) Close() error {
	if ws.session != nil {
		ws.session.Signal(ssh.SIGKILL)
		ws.session.Close()
		ws.session = nil
	}
	if ws.conn != nil {
		ws.conn.Close()
		ws.conn = nil
	}
	return nil
}

func (ws *WebSSH) Read(p []byte) (n int, err error) {
	return ws.reader.Read(p)
}

func (ws *WebSSH) Write(p []byte) (n int, err error) {
	return ws.writer.Write(p)
}

func (ws *WebSSH) SetWinSize(cols, rows int) error {
	return ws.session.WindowChange(rows, cols)
}
