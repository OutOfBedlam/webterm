package webssh

import (
	"fmt"
	"html/template"
	"io"
	"os"

	"github.com/OutOfBedlam/webterm"
	"golang.org/x/crypto/ssh"
)

type Hop struct {
	Network string
	Host    string
	Port    int
	User    string
	Auth    []ssh.AuthMethod
}

type Hops []Hop

func (hops Hops) Connect() (*ssh.Client, error) {
	var client *ssh.Client
	var err error
	for i, hop := range hops {
		network := hop.Network
		host := hop.Host
		port := hop.Port
		user := hop.User
		if network == "" {
			network = "tcp"
		}
		if port == 0 {
			port = 22
		}
		addr := fmt.Sprintf("%s:%d", host, port)
		if user == "" {
			user = os.Getenv("USER")
		}
		conf := &ssh.ClientConfig{
			User:            user,
			Auth:            hop.Auth,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		if i == 0 {
			client, err = ssh.Dial(network, addr, conf)
		} else {
			conn, err := client.Dial(network, fmt.Sprintf("%s:%d", hop.Host, hop.Port))
			if err != nil {
				return nil, err
			}
			ncc, chans, reqs, err := ssh.NewClientConn(conn, addr, conf)
			if err != nil {
				return nil, err
			}
			client = ssh.NewClient(ncc, chans, reqs)
		}
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

var _ webterm.Runner = (*WebSSH)(nil)
var _ webterm.Session = (*WebSSHSession)(nil)

type WebSSH struct {
	Hops     Hops
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
	if conn, err := ws.Hops.Connect(); err != nil {
		return err
	} else {
		ws.conn = conn
	}

	if session, err := ws.conn.NewSession(); err != nil {
		return err
	} else {
		ws.session = session
	}

	stdout, err := ws.session.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := ws.session.StderrPipe()
	if err != nil {
		return err
	}
	stdin, err := ws.session.StdinPipe()
	if err != nil {
		return err
	}

	ws.reader = io.MultiReader(stdout, stderr)
	ws.writer = stdin

	termType := ws.TermType
	if termType == "" {
		termType = "xterm"
	}
	err = ws.session.RequestPty(termType, 40, 80, ssh.TerminalModes{
		ssh.ECHO: 1, // enable echoing
	})
	if err != nil {
		ws.session.Close()
		ws.conn.Close()
		return err
	}

	if ws.Command != "" {
		err = ws.session.Start(ws.Command)
	} else {
		err = ws.session.Shell()
	}
	if err != nil {
		ws.session.Close()
		ws.conn.Close()
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

func (ws *WebSSHSession) Control(data []byte) error {
	return nil
}
