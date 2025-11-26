
## Make HTTP Handler

```go
import (
	"github.com/OutOfBedlam/webterm"
	"github.com/OutOfBedlam/webterm/webexec"
)

mux := http.NewServeMux()
mux.Handle("/web/term/", mc.makeEXEC("/web/term/"))

func makeEXEC(cutPrefix string) http.Handler {
	term := webterm.New(
		&webexec.WebExec{
			Command: "/usr/bin/zsh",
			Args:    []string{"-il"},
			Dir:     "/tmp/",
		},
		webterm.WithCutPrefix(cutPrefix),
		webterm.WithTheme(webterm.ThemeDracula),
	)
	return term
}

```