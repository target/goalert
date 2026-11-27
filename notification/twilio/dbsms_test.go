package twilio

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification"
)

type fakeCallbackRow struct {
	phone      string
	callbackID string
	code       int
	alertID    sql.NullInt64
	serviceID  sql.NullString
	svcName    string
}

type fakeSMSDB struct {
	mu           sync.Mutex
	rows         []fakeCallbackRow
	serviceNames map[string]string
	lookupErr    error
	inserts      int
	commits      int
}

func (f *fakeSMSDB) openDB() *sql.DB {
	return sql.OpenDB(&fakeSMSConnector{db: f})
}

type fakeSMSConnector struct{ db *fakeSMSDB }

func (c *fakeSMSConnector) Connect(context.Context) (driver.Conn, error) {
	return &fakeSMSConn{db: c.db}, nil
}
func (c *fakeSMSConnector) Driver() driver.Driver { return nil }

type fakeSMSConn struct{ db *fakeSMSDB }

func (c *fakeSMSConn) Prepare(query string) (driver.Stmt, error) {
	return &fakeSMSStmt{db: c.db, query: query}, nil
}
func (c *fakeSMSConn) Close() error              { return nil }
func (c *fakeSMSConn) Begin() (driver.Tx, error) { return &fakeSMSTx{db: c.db}, nil }

type fakeSMSTx struct{ db *fakeSMSDB }

func (tx *fakeSMSTx) Commit() error {
	tx.db.mu.Lock()
	tx.db.commits++
	tx.db.mu.Unlock()
	return nil
}
func (tx *fakeSMSTx) Rollback() error { return nil }

type fakeSMSStmt struct {
	db    *fakeSMSDB
	query string
}

func (s *fakeSMSStmt) Close() error  { return nil }
func (s *fakeSMSStmt) NumInput() int { return -1 }

func (s *fakeSMSStmt) Exec(args []driver.Value) (driver.Result, error) {
	q := strings.TrimSpace(s.query)
	if strings.HasPrefix(q, "LOCK ") {
		return driver.RowsAffected(0), nil
	}
	if strings.HasPrefix(q, "INSERT INTO twilio_sms_callbacks") {
		phone := args[0].(string)
		cbID := args[1].(string)
		code := int(args[2].(int64))

		var aID sql.NullInt64
		if args[3] != nil {
			aID = sql.NullInt64{Int64: args[3].(int64), Valid: true}
		}
		var sID sql.NullString
		if args[4] != nil {
			sID = sql.NullString{String: args[4].(string), Valid: true}
		}

		s.db.mu.Lock()
		defer s.db.mu.Unlock()
		s.db.inserts++
		updated := false
		for i := range s.db.rows {
			if s.db.rows[i].phone == phone && s.db.rows[i].code == code {
				s.db.rows[i].callbackID = cbID
				s.db.rows[i].alertID = aID
				s.db.rows[i].serviceID = sID
				if sID.Valid {
					s.db.rows[i].svcName = s.db.serviceNames[sID.String]
				} else {
					s.db.rows[i].svcName = ""
				}
				updated = true
				break
			}
		}
		if !updated {
			row := fakeCallbackRow{
				phone:      phone,
				callbackID: cbID,
				code:       code,
				alertID:    aID,
				serviceID:  sID,
			}
			if sID.Valid {
				row.svcName = s.db.serviceNames[sID.String]
			}
			s.db.rows = append(s.db.rows, row)
		}
		return driver.RowsAffected(1), nil
	}
	return nil, errors.New("unexpected exec query: " + q)
}

func (s *fakeSMSStmt) Query(args []driver.Value) (driver.Rows, error) {
	q := strings.TrimSpace(s.query)
	s.db.mu.Lock()
	defer s.db.mu.Unlock()

	if strings.Contains(q, "FROM twilio_sms_callbacks cb") && strings.Contains(q, "service_id = $3") {
		phone := args[0].(string)
		for _, r := range s.db.rows {
			if r.phone != phone {
				continue
			}
			if args[2] != nil && r.serviceID.Valid && r.serviceID.String == args[2].(string) {
				return &fakeSMSRows{cols: []string{"code"}, vals: [][]driver.Value{{int64(r.code)}}}, nil
			}
			if args[1] != nil && r.alertID.Valid && r.alertID.Int64 == args[1].(int64) {
				return &fakeSMSRows{cols: []string{"code"}, vals: [][]driver.Value{{int64(r.code)}}}, nil
			}
		}
		return &fakeSMSRows{cols: []string{"code"}}, nil
	}

	if strings.Contains(q, "FROM twilio_sms_callbacks cb") && strings.Contains(q, "service_id NOTNULL") {
		phone := args[0].(string)
		var vals [][]driver.Value
		for _, r := range s.db.rows {
			if r.phone == phone {
				vals = append(vals, []driver.Value{int64(r.code)})
			}
		}
		return &fakeSMSRows{cols: []string{"code"}, vals: vals}, nil
	}

	if strings.Contains(q, "JOIN services svc") {
		if s.db.lookupErr != nil {
			return nil, s.db.lookupErr
		}
		phone := args[0].(string)
		code := int(args[1].(int64))
		for _, r := range s.db.rows {
			if r.phone == phone && r.code == code && r.serviceID.Valid {
				return &fakeSMSRows{
					cols: []string{"callback_id", "alert_id", "name"},
					vals: [][]driver.Value{{r.callbackID, nil, r.svcName}},
				}, nil
			}
		}
		return &fakeSMSRows{cols: []string{"callback_id", "alert_id", "name"}}, nil
	}

	return nil, errors.New("unexpected query: " + q)
}

type fakeSMSRows struct {
	cols []string
	vals [][]driver.Value
	idx  int
}

func (r *fakeSMSRows) Columns() []string { return r.cols }
func (r *fakeSMSRows) Close() error      { return nil }
func (r *fakeSMSRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.vals) {
		return io.EOF
	}
	copy(dest, r.vals[r.idx])
	r.idx++
	return nil
}

type fakeReceiver struct {
	notification.Receiver
	receiveFn func(ctx context.Context, callbackID string, result notification.Result) error
}

func (f *fakeReceiver) Receive(ctx context.Context, callbackID string, result notification.Result) error {
	return f.receiveFn(ctx, callbackID, result)
}

func (f *fakeReceiver) IsKnownDest(context.Context, gadb.DestV1) (bool, error) {
	return true, nil
}

func TestDBSMS_InsertDB_RefreshesExistingCallback(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fake := &fakeSMSDB{
		serviceNames: map[string]string{
			"svc-1": "Service One",
			"svc-2": "Service Two",
		},
	}
	sqlDB := fake.openDB()
	defer sqlDB.Close()

	b, err := newDB(ctx, sqlDB)
	require.NoError(t, err)

	// First bundle for svc-1 gets code 100.
	code, err := b.insertDB(ctx, "+15551234567", "cb-old", 0, "svc-1")
	require.NoError(t, err)
	assert.Equal(t, 100, code)

	// A subsequent bundle for the same service should reuse code 100 and update callback_id.
	code, err = b.insertDB(ctx, "+15551234567", "cb-new", 0, "svc-1")
	require.NoError(t, err)
	assert.Equal(t, 100, code)
	assert.Equal(t, 2, fake.inserts)
	assert.Equal(t, 2, fake.commits)

	info, err := b.LookupSvcByCode(ctx, "+15551234567", 100)
	require.NoError(t, err)
	assert.Equal(t, "cb-new", info.CallbackID)
	assert.Equal(t, "Service One", info.ServiceName)

	// A different service gets the next available code.
	code, err = b.insertDB(ctx, "+15551234567", "cb-svc2", 0, "svc-2")
	require.NoError(t, err)
	assert.Equal(t, 101, code)

	// Re-notifying an open alert also refreshes its callback_id on the same code.
	code, err = b.insertDB(ctx, "+15551234567", "cb-alert-1", 42, "")
	require.NoError(t, err)
	assert.Equal(t, 1, code)

	code, err = b.insertDB(ctx, "+15551234567", "cb-alert-2", 42, "")
	require.NoError(t, err)
	assert.Equal(t, 1, code)
	assert.Equal(t, 5, fake.inserts)
}

func TestSMS_ServeMessage_AckAllAndLookupErrors(t *testing.T) {
	t.Parallel()

	newSMSHandler := func(t *testing.T, fake *fakeSMSDB, recv notification.Receiver) (*SMS, func() []string) {
		t.Helper()
		var mu sync.Mutex
		var sent []string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mu.Lock()
			sent = append(sent, r.FormValue("Body"))
			mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"sid":"SM` + strings.Repeat("0", 32) + `","status":"sent"}`))
		}))
		t.Cleanup(srv.Close)

		sqlDB := fake.openDB()
		t.Cleanup(func() { sqlDB.Close() })

		sms, err := NewSMS(context.Background(), sqlDB, &Config{
			BaseURL: srv.URL,
			Client:  srv.Client(),
		})
		require.NoError(t, err)
		sms.SetReceiver(recv)
		return sms, func() []string {
			mu.Lock()
			defer mu.Unlock()
			cp := make([]string, len(sent))
			copy(cp, sent)
			return cp
		}
	}

	makeReq := func(body string) *http.Request {
		form := url.Values{}
		form.Set("From", "+16125551234")
		form.Set("To", "+16125559999")
		form.Set("Body", body)

		req := httptest.NewRequest(http.MethodPost, "/api/v2/twilio/message", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var cfg config.Config
		cfg.Twilio.Enable = true
		cfg.Twilio.AccountSID = "test-account-sid"
		cfg.Twilio.AuthToken = "test-token"
		cfg.Twilio.FromNumber = "+16125559999"
		return req.WithContext(cfg.Context(req.Context()))
	}

	t.Run("ack-all uses refreshed callback", func(t *testing.T) {
		fake := &fakeSMSDB{serviceNames: map[string]string{"svc-1": "My Service"}}
		var gotCallback string
		var gotResult notification.Result
		recv := &fakeReceiver{
			receiveFn: func(_ context.Context, callbackID string, result notification.Result) error {
				gotCallback = callbackID
				gotResult = result
				if callbackID == "cb-stale" {
					return sql.ErrNoRows
				}
				return nil
			},
		}

		sms, getSent := newSMSHandler(t, fake, recv)
		_, err := sms.b.insertDB(context.Background(), "+16125551234", "cb-stale", 0, "svc-1")
		require.NoError(t, err)
		_, err = sms.b.insertDB(context.Background(), "+16125551234", "cb-fresh", 0, "svc-1")
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		sms.ServeMessage(rec, makeReq("100aa"))

		assert.Equal(t, "cb-fresh", gotCallback)
		assert.Equal(t, notification.ResultAcknowledge, gotResult)
		sent := getSent()
		require.Len(t, sent, 1)
		assert.Equal(t, "Acknowledged all alerts for service 'My Service'", sent[0])
	})

	t.Run("non-ErrNoRows lookup error replies system error", func(t *testing.T) {
		fake := &fakeSMSDB{lookupErr: errors.New("connection reset")}
		recv := &fakeReceiver{
			receiveFn: func(context.Context, string, notification.Result) error {
				t.Fatal("Receive should not be called when lookup fails")
				return nil
			},
		}

		sms, getSent := newSMSHandler(t, fake, recv)
		rec := httptest.NewRecorder()
		sms.ServeMessage(rec, makeReq("100aa"))

		sent := getSent()
		require.Len(t, sent, 1)
		assert.Equal(t, "System error. Visit the dashboard to manage alerts.", sent[0])
	})

	t.Run("unknown code still replies unknown reply code", func(t *testing.T) {
		fake := &fakeSMSDB{}
		recv := &fakeReceiver{
			receiveFn: func(context.Context, string, notification.Result) error {
				t.Fatal("Receive should not be called for unknown code")
				return nil
			},
		}

		sms, getSent := newSMSHandler(t, fake, recv)
		rec := httptest.NewRecorder()
		sms.ServeMessage(rec, makeReq("999aa"))

		sent := getSent()
		require.Len(t, sent, 1)
		assert.Equal(t, "Unknown reply code for this action. Visit the dashboard to manage alerts.", sent[0])
	})
}
