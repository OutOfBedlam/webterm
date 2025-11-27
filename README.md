# WebTerm

WebTerm is a Go library that provides web-based terminal functionality using WebSockets and xterm.js. It supports both local command execution and SSH remote connections through a simple HTTP handler interface.

## Features

- 🖥️ **Local Command Execution** - Execute local shell commands in a web-based terminal
- 🌐 **SSH Remote Connection** - Connect to remote servers via SSH in the browser
- 📄 **File Tailing** - Tail one or multiple log files in real-time in the browser
- 🎨 **Multiple Themes** - Built-in color schemes (Solarized, Dracula, Molokai, etc.)
- 📦 **Embedded Static Assets** - All frontend assets are embedded in the binary
- 🔌 **Simple HTTP Handler** - Easy integration with standard Go HTTP servers
- 🪟 **Dynamic Window Resizing** - Supports terminal window size adjustments

## Installation

```bash
go get github.com/OutOfBedlam/webterm
```

## Quick Start

### Local Command Execution

```go
package main

import (
    "net/http"
    
    "github.com/OutOfBedlam/webterm"
    "github.com/OutOfBedlam/webterm/webexec"
)

func main() {
    term := webterm.New(
        &webexec.WebExec{
            Command: "/bin/bash",
            Args:    []string{"-il"},
            Dir:     "/tmp/",
        },
        webterm.WithCutPrefix("/terminal/"),
        webterm.WithTheme(webterm.ThemeDracula),
    )
    
    http.Handle("/terminal/", term)
    http.ListenAndServe(":8080", nil)
}
```

### SSH Remote Connection

```go
package main

import (
    "net/http"
    "os"
    "path/filepath"
    
    "github.com/OutOfBedlam/webterm"
    "github.com/OutOfBedlam/webterm/webssh"
)

func main() {
    key, _ := os.ReadFile(filepath.Join(os.Getenv("HOME"), ".ssh/id_rsa"))
    
    term := webterm.New(
        &webssh.WebSSH{
            Host: "example.com",
            Port: 22,
            User: "username",
            Auth: []ssh.AuthMethod{
                webssh.AuthPrivateKey(key),
            },
            TermType: "xterm-256color",
        },
        webterm.WithCutPrefix("/ssh/"),
        webterm.WithTheme(webterm.ThemeMolokai),
    )
    
    http.Handle("/ssh/", term)
    http.ListenAndServe(":8080", nil)
}
```

### File Tailing

```go
package main

import (
    "net/http"
    
    "github.com/OutOfBedlam/webterm"
    "github.com/OutOfBedlam/webterm/webtail"
)

func main() {
    // Tail a single file
    term := webterm.New(
        &webtail.WebTail{
            Tails: []*webtail.Tail{
                webtail.NewTail("/var/log/syslog"),
            },
        },
        webterm.WithCutPrefix("/logs/"),
        webterm.WithTheme(webterm.ThemeSolarizedDark),
    )
    
    // Or tail multiple files
    multiTerm := webterm.New(
        &webtail.WebTail{
            Tails: []*webtail.Tail{
                webtail.NewTail("/var/log/syslog"),
                webtail.NewTail("/var/log/auth.log"),
                webtail.NewTail("/var/log/nginx/access.log"),
            },
        },
        webterm.WithCutPrefix("/logs/"),
        webterm.WithTheme(webterm.ThemeDracula),
    )
    
    http.Handle("/logs/", term)
    http.ListenAndServe(":8080", nil)
}
```

## Configuration Options

### WebTerm Options

```go
// Set URL prefix to cut from requests
webterm.WithCutPrefix("/my-terminal/")

// Set terminal color theme
webterm.WithTheme(webterm.ThemeDracula)
```

### WebExec Configuration

```go
&webexec.WebExec{
    Command: "/usr/bin/zsh",  // Command to execute
    Args:    []string{"-il"}, // Command arguments
    Dir:     "/home/user",    // Working directory
}
```

### WebSSH Configuration

```go
&webssh.WebSSH{
    Network:  "tcp",              // Network type (default: "tcp")
    Host:     "example.com",      // SSH host
    Port:     22,                 // SSH port (default: 22)
    User:     "username",         // SSH user (default: $USER)
    Auth:     []ssh.AuthMethod{}, // Authentication methods
    TermType: "xterm-256color",   // Terminal type (default: "xterm")
    Command:  "",                 // Optional command to run (default: shell)
}
```

### SSH Authentication

```go
// Private key authentication
key, _ := os.ReadFile("/path/to/private/key")
auth := webssh.AuthPrivateKey(key)

// Password authentication
auth := webssh.AuthPassword("your-password")

// Multiple authentication methods
Auth: []ssh.AuthMethod{
    webssh.AuthPrivateKey(key),
    webssh.AuthPassword(password),
}
```

### WebTail Configuration

```go
&webtail.WebTail{
    Tails: []*webtail.Tail{
        webtail.NewTail("/path/to/file.log"),  // Single file
        webtail.NewTail("/path/to/other.log"), // Multiple files
    },
}
```

## Available Themes

WebTerm includes several built-in color themes:

- `ThemeSolarizedDark`
- `ThemeSolarizedLight`
- `ThemeDracula`
- `ThemeMolokai`
- `ThemeNordic`

Example:
```go
webterm.WithTheme(webterm.ThemeDracula)
```

## Custom Runner

You can implement your own terminal backend by implementing the `Runner` interface:

```go
type Runner interface {
    Open() error
    Close() error
    Read(p []byte) (n int, err error)
    Write(p []byte) (n int, err error)
    SetWinSize(cols, rows int) error
}
```

## Sub-packages

- **webexec** - Local command execution runner
- **webssh** - SSH remote connection runner
- **webtail** - File tailing runner for monitoring log files

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.
