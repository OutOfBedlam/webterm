
## Make HTTP Handler

```go
import (
	"github.com/OutOfBedlam/webterm"
	"github.com/OutOfBedlam/webterm/webssh"
)

mux := http.NewServeMux()
mux.Handle("/web/ssh/", makeSSH("/web/ssh/"))

func makeWebSSH(cutPrefix string) http.Handler {
	key, _ := os.ReadFile(filepath.Join(os.Getenv("HOME"), ".ssh/id_rsa"))
	user := os.Getenv("USER")

	term := webterm.New(
		&webssh.WebSSH{
			Host: "127.0.0.1",
			Port: 22,
			User: user,
			Auth: []ssh.AuthMethod{
				webssh.AuthPrivateKey(key),
			},
			TermType: "xterm-256color",
		},
		webterm.WithCutPrefix(cutPrefix),
		webterm.WithTheme(webterm.ThemeMolokai),
	)
	return term
}
```