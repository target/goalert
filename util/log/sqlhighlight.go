// +build sql_highlight

package log

import (
	"io"

	"github.com/alecthomas/chroma/quick"
)

func init() {
	sqlHighlight = func(w io.Writer, q string) error {
		return quick.Highlight(w, q, "sql", "terminal256", "monokai")
	}
}
