package smtpsrv

import (
	"io"
	"net/mail"
	"reflect"
	"strings"
	"testing"
)

func TestParseSanitizeMessage(t *testing.T) {
	type args struct {
		m *mail.Message
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "multipart",
			args: args{m: &mail.Message{Header: map[string][]string{"Content-Type": {"multipart/mixed; boundary=foo"}}, Body: strings.NewReader(
				`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!` + "\r\n" + `--foo--`),
			}},
			want:    []byte{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSanitizeMessage(tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSanitizeMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSanitizeMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseMultipart(t *testing.T) {
	type args struct {
		body     io.Reader
		boundary string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "multipart",
			args: args{body: strings.NewReader(
				`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!` + "\r\n" + `--foo--`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(
				`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!` + "\r\n" + `--foo--`),
				boundary: "bar",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(
				`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!` + "\r\n" + `--foo--`),
				boundary: "baz",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(
				`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!` + "\r\n" + `--foo--`),
				boundary: "qux",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(
				`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!` + "\r\n" + `--foo--`),
				boundary: "quux",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(
				`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!` + "\r\n" + `--foo--`),
				boundary: "corge",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(
				`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!` + "\r\n" + `--foo--`),
				boundary: "grault",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(
				`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!` + "\r\n" + `--foo--`),
				boundary: "garply",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(
				`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!` + "\r\n" + `--foo--`),
				boundary: "waldo",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!` + "\r\n" + `--foo--
--bar
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!` + "\r\n" + `--bar--`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!` + "\r\n"),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n"),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		got, err := parseMultipart(tt.args.body, tt.args.boundary)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. readMultipart() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. readMultipart() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func Test_readMultipart2(t *testing.T) {
	type args struct {
		body     io.Reader
		boundary string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!`),
				boundary: "foo",
			},
			want:    []byte(`Hello, world!`),
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		got, err := parseMultipart(tt.args.body, tt.args.boundary)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. readMultipart2() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(string(got), string(tt.want)) {
			t.Errorf("%q. readMultipart2() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func Test_readMultipart3(t *testing.T) {
	type args struct {
		body     io.Reader
		boundary string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!`),
				boundary: "foo",
			},
			want:    []byte(`Hello, world!`),
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		got, err := parseMultipart(tt.args.body, tt.args.boundary)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. readMultipart3() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(string(got), string(tt.want)) {
			t.Errorf("%q. readMultipart3() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func Test_readMultipart4(t *testing.T) {
	type args struct {
		body     io.Reader
		boundary string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!`),
				boundary: "foo",
			},
			want:    []byte(`Hello, world!`),
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		got, err := parseMultipart(tt.args.body, tt.args.boundary)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. readMultipart4() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(string(got), string(tt.want)) {
			t.Errorf("%q. readMultipart4() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func Test_readMultipart5(t *testing.T) {
	type args struct {
		body     io.Reader
		boundary string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!`),
				boundary: "foo",
			},
			want:    []byte(`Hello, world!`),
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		got, err := parseMultipart(tt.args.body, tt.args.boundary)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. readMultipart5() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(string(got), string(tt.want)) {
			t.Errorf("%q. readMultipart5() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func Test_readMultipart6(t *testing.T) {
	type args struct {
		body     io.Reader
		boundary string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!`),
				boundary: "foo",
			},
			want:    []byte(`Hello, world!`),
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		got, err := parseMultipart(tt.args.body, tt.args.boundary)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. readMultipart6() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(string(got), string(tt.want)) {
			t.Errorf("%q. readMultipart6() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func Test_readMultipart7(t *testing.T) {
	type args struct {
		body     io.Reader
		boundary string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!`),
				boundary: "foo",
			},
			want:    []byte(`Hello, world!`),
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		got, err := parseMultipart(tt.args.body, tt.args.boundary)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. readMultipart7() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(string(got), string(tt.want)) {
			t.Errorf("%q. readMultipart7() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func Test_readMultipart8(t *testing.T) {
	type args struct {
		body     io.Reader
		boundary string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!`),
				boundary: "foo",
			},
			want:    []byte(`Hello, world!`),
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		got, err := parseMultipart(tt.args.body, tt.args.boundary)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. readMultipart8() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(string(got), string(tt.want)) {
			t.Errorf("%q. readMultipart8() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func Test_readMultipart9(t *testing.T) {
	type args struct {
		body     io.Reader
		boundary string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!`),
				boundary: "foo",
			},
			want:    []byte(`Hello, world!`),
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		got, err := parseMultipart(tt.args.body, tt.args.boundary)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. readMultipart9() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(string(got), string(tt.want)) {
			t.Errorf("%q. readMultipart9() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func Test_readMultipart10(t *testing.T) {
	type args struct {
		body     io.Reader
		boundary string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8` + "\r\n\r\n" + `Hello, world!`),
				boundary: "foo",
			},
			want:    []byte(`Hello, world!`),
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo
Content-Type: text/plain; charset=utf-8`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "multipart",
			args: args{body: strings.NewReader(`--foo`),
				boundary: "foo",
			},
			want:    []byte{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		got, err := parseMultipart(tt.args.body, tt.args.boundary)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. readMultipart10() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(string(got), string(tt.want)) {
			t.Errorf("%q. readMultipart10() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
