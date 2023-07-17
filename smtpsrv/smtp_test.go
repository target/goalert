package smtpsrv

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/target/goalert/alert"
	"github.com/target/goalert/integrationkey"
)

func init() {
	Handler = IngressHandler{
		alerts:  &alert.Store{},
		intKeys: &integrationkey.Store{},
		cfg: &Config{
			Domain: "goalert.local",
			AllowedDomains: []string{
				"goalert.local",
			},
			ListenAddr: "localhost:9025",
		},
	}
}

func TestSession_Data(t *testing.T) {
	type args struct {
		r io.Reader
	}

	msg := `Date: %s
From: %s
To: %s
Subject: %s
Content-Type: %s

%s
`

	timeHeader := time.Now().Local().Format(time.RFC1123Z)
	from := "test@localhost"
	to := "test@goalert.local"
	plain := "text/plain; charset=utf-8"
	html := "text/html; charset=utf-8"
	mixed := "multipart/mixed; boundary=foobar"

	tests := []struct {
		name    string
		s       *Session
		args    args
		wantErr bool
	}{
		{
			name: "test 0",
			s:    &Session{},
			args: args{
				r: strings.NewReader(fmt.Sprintf(msg, timeHeader, from, to, "Test Case 0", plain, "Test Case 0 Body")),
			},
			wantErr: false,
		},
		{
			name: "test 1",
			s:    &Session{},
			args: args{
				r: strings.NewReader(fmt.Sprintf(msg, timeHeader, from, to, "Test Case 1", html, "<h1>Test Case 1</h1> </br><b>Body</b>")),
			},
			wantErr: false,
		},
		{
			name: "test 2",
			s:    &Session{},
			args: args{
				r: strings.NewReader(fmt.Sprintf(msg, timeHeader, from, to, "Test Case 2", mixed, "--foobar\r\nContent-Type: text/plain\r\n\r\nTest Case 2 Body\r\n--foobar\r\nContent-Type: text/html\r\n\r\n<h1>Test Case 2</h1> </br><b>Body</b>\r\n--foobar--\r\n")),
			},
			wantErr: false,
		},
		// {
		// 	name: "test 3",
		// 	s:    &Session{},
		// 	args: args{
		// 		r: strings.NewReader(fmt.Sprintf(msg, timeHeader, from, to, "Test Case 3", plain, "Test Case 3 Body")),
		// 	},
		// 	wantErr: false,
		// },
		// {
		// 	name: "test 4",
		// 	s:    &Session{},
		// 	args: args{
		// 		r: strings.NewReader(fmt.Sprintf(msg, timeHeader, from, to, "Test Case 4", plain, "Test Case 4 Body")),
		// 	},
		// 	wantErr: false,
		// },
		// {
		// 	name: "test 5",
		// 	s:    &Session{},
		// 	args: args{
		// 		r: strings.NewReader(fmt.Sprintf(msg, timeHeader, from, to, "Test Case 5", plain, "Test Case 5 Body")),
		// 	},
		// 	wantErr: false,
		// },
		// {
		// 	name: "test 6",
		// 	s:    &Session{},
		// 	args: args{
		// 		r: strings.NewReader(fmt.Sprintf(msg, timeHeader, from, to, "Test Case 6", plain, "Test Case 6 Body")),
		// 	},
		// 	wantErr: false,
		// },
	}
	for _, tt := range tests {
		if err := tt.s.Data(tt.args.r); (err != nil) != tt.wantErr {
			t.Errorf("%q. Session.Data() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}
