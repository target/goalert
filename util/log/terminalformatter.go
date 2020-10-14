package log

import (
	"bytes"
	"io"
	"strings"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/target/goalert/util/sqlutil"
)

var sqlHighlight = func(w io.Writer, q string) error {
	_, err := io.WriteString(w, q)
	return err
}

type terminalFormatter struct {
	fallback logrus.TextFormatter
}
type queryError interface {
	Query() string
	Cause() *sqlutil.Error
}

func lineCol(q string, pos int) (int, int) {
	if pos > len(q) {
		pos = len(q)
	}
	lines := strings.Split(q[:pos], "\n")
	lastLine := lines[len(lines)-1]
	lastLine = strings.Replace(lastLine, "\t", strings.Repeat(" ", 8), -1)
	return len(lines), len(lastLine) - 1
}
func makeCodeFrame(q string, e *sqlutil.Error) string {

	buf := new(bytes.Buffer)
	err := sqlHighlight(buf, q)
	var code string
	if err != nil {
		code = q
	} else {
		code = buf.String()
	}

	buf.Reset()
	pos := e.Position
	if err == nil {
		l, c := lineCol(q, pos)
		lines := strings.Split(code, "\n")
		buf.WriteString(strings.Join(lines[:l], "\n") + "\n")
		buf.WriteString(
			color.New(color.Bold, color.FgRed).Sprint(
				strings.Repeat("_", c)) + color.New(color.Bold, color.FgMagenta).Sprint("^") +
				"\n" + color.New(color.FgRed).Sprint(e.Message) + "\n\n",
		)
		buf.WriteString(strings.Join(lines[l:], "\n"))
	}

	return buf.String()
}
func (t *terminalFormatter) Format(e *logrus.Entry) ([]byte, error) {
	err, ok := e.Data["error"].(error)
	if ok {
		var qe queryError
		if errors.As(err, &qe) {
			e.Message += qe.Cause().Message
			frame := makeCodeFrame(qe.Query(), qe.Cause())
			if frame != "" {
				e.Message += "\n\n" + frame
			}

		} else {
			e.Message += err.Error()
		}
		delete(e.Data, "error")
	}
	src, ok := e.Data["Source"].(string)
	if ok {
		delete(e.Data, "Source")
		e.Message += "\n" + src
	}

	return t.fallback.Format(e)
}
