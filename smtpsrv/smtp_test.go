package smtpsrv

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/auth/authtoken"
	"github.com/target/goalert/integrationkey"
)

func init() {
	Handler.alerts = &alert.Store{}
	Handler.intKeys = &mockIntKeyStore{}
	Handler.cfg = &Config{
		Domain: "goalert.local",
		AllowedDomains: []string{
			"goalert.local",
		},
	}
}

// mock method Authorize on integrationkey.Store
type MockIntKeyStore interface {
	Authorize(ctx context.Context, tok authtoken.Token, t integrationkey.Type) (context.Context, error)
}

type mockIntKeyStore struct{}

func (m *mockIntKeyStore) Authorize(ctx context.Context, tok authtoken.Token, t integrationkey.Type) (context.Context, error) {
	return ctx, nil
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
	to := uuid.New().String() + "@goalert.local"
	plain := "text/plain; charset=utf-8"
	html := "text/html; charset=utf-8"
	mixed := "multipart/mixed; boundary=foobar"

	tests := map[string]struct {
		s       *Session
		args    args
		wantErr bool
	}{
		"plain": {
			s: &Session{},
			args: args{
				r: strings.NewReader(fmt.Sprintf(msg, timeHeader, from, to, "Test Case 0", plain, "Test Case 0 Body")),
			},
			wantErr: false,
		},
		"plain no header": {
			s: &Session{},
			args: args{
				r: strings.NewReader(fmt.Sprintf(msg, timeHeader, from, to, "Test Case 1", "", "Test Case 1 Body")),
			},
			wantErr: false,
		},
		"html": {
			s: &Session{},
			args: args{
				r: strings.NewReader(fmt.Sprintf(msg, timeHeader, from, to, "Test Case 2", html, "<h1>Test Case 2</h1> </br><b>Body</b>")),
			},
			wantErr: false,
		},
		"multipart valid": {
			s: &Session{},
			args: args{
				r: strings.NewReader(fmt.Sprintf(msg, timeHeader, from, to, "Test Case 3", mixed, "--foobar\r\nContent-Type: text/plain\r\n\r\nTest Case 3 Body\r\n--foobar\r\nContent-Type: text/html\r\n\r\n<h1>Test Case 3</h1> </br><b>Body</b>\r\n--foobar--\r\n")),
			},
			wantErr: false,
		},
		"multipart invalid boundary": {
			s: &Session{},
			args: args{
				r: strings.NewReader(fmt.Sprintf(msg, timeHeader, from, to, "Test Case 4", mixed, "--foo\r\nContent-Type: text/plain\r\n\r\nTest Case 4 Body\r\n--foo\r\nContent-Type: text/html\r\n\r\n<h1>Test Case 4</h1> </br><b>Body</b>\r\n--foo--\r\n")),
			},
			wantErr: true,
		},
		"multipart missing sub types": {
			s: &Session{},
			args: args{
				r: strings.NewReader(fmt.Sprintf(msg, timeHeader, from, to, "Test Case 5", mixed, "--foobar5\r\nTest Case 5 Body\r\n--foobar\r\n<h1>Test Case 5</h1> </br><b>Body</b>\r\n--foobar--\r\n")),
			},
			wantErr: true,
		},
		"multipart missing plain": {
			s: &Session{},
			args: args{
				r: strings.NewReader(fmt.Sprintf(msg, timeHeader, from, to, "Test Case 6", mixed, "--foobar\r\nContent-Type: text/html\r\n\r\n<h1>Test Case 6</h1> </br><b>Body</b>\r\n--foobar--\r\n")),
			},
			wantErr: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if err := tt.s.Data(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("%q. Session.Data() error = %v, wantErr %v", name, err, tt.wantErr)
			}
		})
	}
}
