package webssh

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"

	"github.com/OutOfBedlam/webterm"
	"golang.org/x/crypto/ssh"
)

var _ webterm.Runner = (*WebSSH)(nil)
var _ webterm.Session = (*WebSSHSession)(nil)

type WebSSH struct {
	Network  string
	Host     string
	Port     int
	User     string
	Auth     []ssh.AuthMethod
	TermType string
	Command  string
}

func (ws *WebSSH) Session() (webterm.Session, error) {
	return &WebSSHSession{WebSSH: *ws}, nil
}

func (ws *WebSSH) Template() (*template.Template, any) {
	return nil, nil
}

type WebSSHSession struct {
	WebSSH

	conn    *ssh.Client
	session *ssh.Session
	reader  io.Reader
	writer  io.Writer
}

func (ws *WebSSHSession) Open() error {
	network := ws.Network
	host := ws.Host
	port := ws.Port
	user := ws.User
	termType := ws.TermType
	if network == "" {
		network = "tcp"
	}
	if port == 0 {
		port = 22
	}
	if user == "" {
		user = os.Getenv("USER")
	}
	if termType == "" {
		termType = "xterm"
	}
	auth := append(ws.Auth, ssh.PasswordCallback(ws.no_password))

	conn, err := ssh.Dial(network, fmt.Sprintf("%s:%d", host, port), &ssh.ClientConfig{
		User:            user,
		Auth:            auth,
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

	err = session.RequestPty(termType, 40, 80, ssh.TerminalModes{
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

func (ws *WebSSHSession) Close() error {
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

func (ws *WebSSHSession) Read(p []byte) (n int, err error) {
	return ws.reader.Read(p)
}

func (ws *WebSSHSession) Write(p []byte) (n int, err error) {
	return ws.writer.Write(p)
}

func (ws *WebSSHSession) SetWinSize(cols, rows int) error {
	return ws.session.WindowChange(rows, cols)
}

func (ws *WebSSHSession) no_password() (string, error) {
	return "", errors.New("no password provided")
}

func (ws *WebSSHSession) Control(data []byte) error {
	return nil
}
