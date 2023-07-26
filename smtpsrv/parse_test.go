package smtpsrv

import (
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
