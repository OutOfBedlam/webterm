
## Make HTTP Handler

```go
import (
	"github.com/OutOfBedlam/webterm"
	"github.com/OutOfBedlam/webterm/webtail"
)

mux := http.NewServeMux()
mux.Handle("/web/logs/", makeTail("/web/logs/"))

func makeTail(cutPrefix string) http.Handler {
	tails := []*webtail.Tail{
		webtail.NewTail(filename),
	}

	term := webterm.New(
		&webtail.WebTail{Tails: tails},
        webterm.WithCutPrefix(cutPrefix),
        webterm.WithTheme(webterm.ThemeSolarizedDark),
        webterm.WithFontSize(11),
	)
	return term
}
```