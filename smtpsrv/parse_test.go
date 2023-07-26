package smtpsrv

import (
	"bytes"
	"io"
	"net/mail"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSanitizeMessage(t *testing.T) {
	type args struct {
		m *mail.Message
	}

	plain := "text/plain;"
	html := "text/html;"
	mixed := "multipart/mixed; boundary=foo"

	tests := map[string]struct {
		args    args
		want    []byte
		wantErr bool
	}{
		"plain only": {
			args: args{m: &mail.Message{Header: map[string][]string{"Content-Type": {plain}}, Body: strings.NewReader(
				"Hello, world!"),
			}},
			want:    []byte("Hello, world!"),
			wantErr: false,
		},
		"html only": {
			args: args{m: &mail.Message{Header: map[string][]string{"Content-Type": {html}}, Body: strings.NewReader(
				"<h1>Hello, world!</h1>"),
			}},
			want:    []byte("Hello, world!"),
			wantErr: false,
		},
		"multipart plain only": {
			args: args{m: &mail.Message{Header: map[string][]string{"Content-Type": {mixed}}, Body: strings.NewReader(
				"--foo\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nHello, world!\r\n--foo--"),
			}},
			want:    []byte("Hello, world!"),
			wantErr: false,
		},
		"multipart html only": {
			args: args{m: &mail.Message{Header: map[string][]string{"Content-Type": {mixed}}, Body: strings.NewReader(
				"--foo\r\nContent-Type: text/html; charset=utf-8\r\n\r\n<h1>Hello, world!</h1>\r\n--foo--"),
			}},
			want:    []byte("Hello, world!"),
			wantErr: false,
		},
		"multipart plain and html": {
			args: args{m: &mail.Message{Header: map[string][]string{"Content-Type": {mixed}}, Body: strings.NewReader(
				"--foo\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nHello, world TEXT!\r\n--foo\r\nContent-Type: text/html; charset=utf-8\r\n\r\n<h1>Hello, world!</h1>\r\n--foo--"),
			}},
			want:    []byte("Hello, world TEXT!"),
			wantErr: false,
		},
		"multipart html and plain": {
			args: args{m: &mail.Message{Header: map[string][]string{"Content-Type": {mixed}}, Body: strings.NewReader(
				"--foo\r\nContent-Type: text/html; charset=utf-8\r\n\r\n<h1>Hello, world!</h1>\r\n--foo\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nHello, world TEXT!\r\n--foo--"),
			}},
			want:    []byte("Hello, world TEXT!"),
			wantErr: false,
		},
		"multipart html and json": {
			args: args{m: &mail.Message{Header: map[string][]string{"Content-Type": {mixed}}, Body: strings.NewReader(
				"--foo\r\nContent-Type: text/html; charset=utf-8\r\n\r\n<h1>Hello, world!</h1>\r\n--foo\r\nContent-Type: application/json; charset=utf-8\r\n\r\n{\"foo\": \"bar\"}\r\n--foo--"),
			}},
			want:    []byte("Hello, world!"),
			wantErr: false,
		},
		"multipart xml and json": {
			args: args{m: &mail.Message{Header: map[string][]string{"Content-Type": {mixed}}, Body: strings.NewReader(
				"--foo\r\nContent-Type: application/xml; charset=utf-8\r\n\r\n<h1>Hello, world!</h1>\r\n--foo\r\nContent-Type: application/json; charset=utf-8\r\n\r\n{\"foo\": \"bar\"}\r\n--foo--"),
			}},
			want:    nil,
			wantErr: true,
		},
		"multipart invalid boundary": {
			args: args{m: &mail.Message{Header: map[string][]string{"Content-Type": {mixed}}, Body: strings.NewReader(
				"--bar\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nHello, world!\r\n--bar--"),
			}},
			want:    nil,
			wantErr: true,
		},
		"no media type": {
			args: args{m: &mail.Message{Header: map[string][]string{"Content-Type": {""}}, Body: strings.NewReader(
				"Hello, world!"),
			}},
			want:    []byte("Hello, world!"),
			wantErr: false,
		},
		"no media type html": {
			args: args{m: &mail.Message{Header: map[string][]string{"Content-Type": {""}}, Body: strings.NewReader(
				"<h1>Hello, world!</h1>"),
			}},
			want:    []byte("<h1>Hello, world!</h1>"),
			wantErr: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ParseSanitizeMessage(tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSanitizeMessage() error = %v, wantErr %v\n", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSanitizeMessage() = got %q, want %q\n", string(got), string(tt.want))
			}
		})
	}
}

func Fuzz_parseMultipart(f *testing.F) {
	f.Add([]byte("--foo\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nHello, world!\r\n--foo--"), "foo")
	f.Add([]byte("--bar\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nHello, world!\r\n--bar--"), "bar")
	f.Add([]byte("--foo\r\nContent-Type: text/html; charset=utf-8\r\n\r\n<h1>Hello, world!</h1>\r\n--foo--"), "foo")
	f.Add([]byte("--foo\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nHello, world!\r\n--foo\r\nContent-Type: text/html; charset=utf-8\r\n\r\n<h1>Hello, world!</h1>\r\n--foo--"), "foo")

	f.Add([]byte("--foo\r\nContent-Type: text/plain; charset=iso-8859-1\r\nContent-Transfer-Encoding: base64\r\n\r\nSGVsbG8sIHdvcmxkIQ==\r\n--foo--"), "foo")

	f.Add([]byte("--bar\r\nContent-Type: text/plain; charset=utf-8\r\nContent-Disposition: attachment; filename=\"foo.txt\"\r\n\r\nHello, world!\r\n--bar--"), "bar")

	f.Add([]byte("--foo\r\nContent-Type: multipart/alternative; boundary=bar\r\n\r\n--bar\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nHello, world!\r\n--bar\r\nContent-Type: text/html; charset=utf-8\r\n\r\n<h1>Hello, world!</h1>\r\n--bar--\r\n--foo--"), "foo")

	f.Add([]byte("--foo\r\n\r\nHello, world!\r\n--foo--"), "foo")

	f.Add([]byte("--foo\r\nContent-Type: text/plain; charset=utf-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n=48=65=6C=6C=6F=2C=20=77=6F=72=6C=64=21\r\n--foo--"), "foo")

	f.Add([]byte("--foo\r\nContent-Type: image/jpeg\r\nContent-Transfer-Encoding: base64\r\n\r\n/9j/4AAQSkZJRgABAQEAAAAAAAD/2wBDAAgGBgcGBQgHBwcJCQgKDBQNDAsLDBkSEw8UHRofHh0aHBwgJC4nICIsIxwcKDcpLDAxNDQ0Hyc5PTgyPC4zNDL/2wBDAQkJCQwLDBgNDRgyIRwhMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjL/wAARCAFyAlgDASIAAhEBAxEB/8QAHwAAAQUBAQEBAQEAAAAAAAAAAAECAwQFBgcICQoL/8QAtRAAAgEDAwIEAwUFBAQAAAF9AQIDAAQRBRIhMUEGE1FhByJxFDKBkaEII0KxwRVS0fAkM2JyggkKFhcYGRolJicoKSo0NTY3ODk6Q0RFRkdISUpTVFVWV1hZWmNkZWZnaGlqc3R1dnd4eXqDhIWGh4iJipKTlJWWl5iZmqKjpKWmp6ipqrKztLW2t7i5usLDxMXGx8jJytLT1NXW19jZ2uHi4+Tl5ufo6erx8vP09fb3+Pn6/8QAHwEAAwEBAQEBAQEBAQAAAAAAAAECAwQFBgcICQoL/8QAtREAAgECBAQDBAcFBAQAAQJ3AAECAxEEBSExBhJBUQdhcRMiMoEIFEKRobHBCSMzUvAVYnLRChYkNOEl8RcYGRomJygpKjU2Nzg5OkNERUZHSElKU1RVVldYWVpjZGVmZ2hpanN0dXZ3eHl6goOEhYaHiImKkpOUlZaXmJmaoqOkpaanqKmqsrO0tba3uLm6wsPExcbHyMnK0tPU1dbX2Nna4uPk5ebn6Onq8vP09fb3+Pn6/9oADAMBAAIRAxEAPwD9/KKKKAP/2Q==\r\n--foo--"), "foo")

	f.Add([]byte("--foo\r\nContent-Type: application/json; charset=utf-8\r\n\r\n{\"message\":\"Hello, world!\"}\r\n--foo--"), "foo")

	f.Fuzz(func(t *testing.T, data []byte, boundary string) {
		_, err := parseMultipart(bytes.NewReader(data), boundary)
		if err != nil {
			t.Skip()
		}
	})
}

func Test_parseMultipart(t *testing.T) {
	type args struct {
		body     io.Reader
		boundary string
	}

	tests := map[string]struct {
		args    args
		want    []byte
		wantErr bool
	}{
		"plain only": {
			args: args{
				body: strings.NewReader(
					"--foo\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nHello, world!\r\n--foo--"),
				boundary: "foo",
			},
			want:    []byte("Hello, world!"),
			wantErr: false,
		},
		"html only": {
			args: args{
				body: strings.NewReader(
					"--foo\r\nContent-Type: text/html; charset=utf-8\r\n\r\n<h1>Hello, world!</h1>\r\n--foo--"),
				boundary: "foo",
			},
			want:    []byte("Hello, world!"),
			wantErr: false,
		},
		"plain and html": {
			args: args{
				body: strings.NewReader(
					"--foo\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nHello, world!\r\n--foo\r\nContent-Type: text/html; charset=utf-8\r\n\r\n<h1>Hello, world!</h1>\r\n--foo--"),
				boundary: "foo",
			},
			want:    []byte("Hello, world!"),
			wantErr: false,
		},
		"no parts": {
			args: args{
				body: strings.NewReader(
					"--foo--"),
				boundary: "foo",
			},
			want:    nil,
			wantErr: true,
		},
		"missing header": {
			args: args{
				body: strings.NewReader(
					"--foo\r\nHello, world!\r\n--foo--"),
				boundary: "foo",
			},
			want:    nil,
			wantErr: true,
		},
		"invalid boundary": {
			args: args{
				body: strings.NewReader(
					"--bar\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nHello, world!\r\n--bar--"),
				boundary: "foo",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := parseMultipart(tt.args.body, tt.args.boundary)
			if tt.wantErr {
				assert.Error(t, err, "parseMultipart()")
				return
			}
			require.NoError(t, err, "parseMultipart()")
			assert.Equal(t, string(tt.want), string(got), "parseMultipart()")
		})
	}
}
