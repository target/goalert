package uik

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
)

// SampleFromRequest will extract the relevant data from an incoming request.
func SampleFromRequest(r *http.Request) (*gadb.UIKRequestDataV1, error) {
	var s gadb.UIKRequestDataV1
	err := json.NewDecoder(r.Body).Decode(&s.Body)
	if err != nil {
		return nil, err
	}

	s.Query = r.URL.Query()
	s.UserAgent = r.UserAgent()
	s.RemoteAddr = r.RemoteAddr

	return &s, nil
}

// EnvFromSample will convert a UIKRequestDataV1 into a map[string]any for use in an expr VM.
func EnvFromSample(s gadb.UIKRequestDataV1) map[string]any {
	flatQuery := make(map[string]string)
	for key := range s.Query {
		flatQuery[key] = s.Query.Get(key)
	}

	return map[string]any{
		"sprintf": fmt.Sprintf,
		"req": map[string]any{
			"body":   s.Body,
			"query":  flatQuery,
			"querya": map[string][]string(s.Query),
			"ua":     s.UserAgent,
			"ip":     s.RemoteAddr,
		},
	}
}

type limiter struct {
	limit  map[limitKey]time.Time
	cleanT <-chan time.Time
	sInit  sync.Once
	mx     sync.RWMutex
}
type limitKey struct {
	KeyID  uuid.UUID
	Failed bool
}

func (l *limiter) init() {
	l.sInit.Do(func() {
		l.limit = make(map[limitKey]time.Time)
		l.cleanT = time.Tick(10 * time.Minute)
	})
}

func (l *limiter) clean() {
	l.init()
	select {
	case <-l.cleanT:
	default:
		return
	}

	l.mx.RLock()
	size := len(l.limit)
	l.mx.RUnlock()
	if size < 1000 {
		// fast path, no need to clean
		return
	}

	l.mx.Lock()
	defer l.mx.Unlock()

	now := time.Now()
	for k, t := range l.limit {
		if now.Sub(t) >= time.Minute {
			delete(l.limit, k)
		}
	}
}

func (l *limiter) check(keyID uuid.UUID, failed bool) bool {
	l.clean()

	key := limitKey{KeyID: keyID, Failed: failed}
	l.mx.RLock()
	t, ok := l.limit[key]
	l.mx.RUnlock()

	now := time.Now()
	if ok && now.Sub(t) < time.Minute {
		return false
	}

	l.mx.Lock()
	if l.limit[key] != t {
		l.mx.Unlock()
		// another goroutine updated the limit
		return false
	}

	l.limit[key] = now
	l.mx.Unlock()

	return true
}

// Sample will record a sample of incoming data according to the following rules:
//
// - The 3 most recent failed and successful samples are stored.
// - Older ones are discarded.
// - Only one sample is stored per minute at most.
func (h *Handler) Sample(ctx context.Context, keyID uuid.UUID, s gadb.UIKRequestDataV1, failed bool) {
	if !h.sampleLimit.check(keyID, failed) {
		// already recorded a sample for this key in the last minute
		// return
	}

	err := h._Sample(ctx, keyID, s, failed)
	if err != nil {
		log.Log(ctx, fmt.Errorf("insert UIK sample: %w", err))
	}
}

const SampleLimit = 3

func (h *Handler) _Sample(ctx context.Context, keyID uuid.UUID, s gadb.UIKRequestDataV1, failed bool) error {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer sqlutil.Rollback(ctx, "insert UIK sample", tx)

	db := gadb.New(tx)
	err = gadb.New(h.db).UIKInsertSample(ctx, gadb.UIKInsertSampleParams{
		ID:          uuid.Must(uuid.NewV7()),
		KeyID:       keyID,
		Failed:      failed,
		RequestData: gadb.UIKRequestData{Version: 1, V1: s},
	})
	if err != nil {
		return fmt.Errorf("insert sample: %w", err)
	}

	err = db.UIKDeleteOldestSamples(ctx, gadb.UIKDeleteOldestSamplesParams{KeyID: keyID, Failed: failed, Offset: SampleLimit})
	if err != nil {
		return fmt.Errorf("delete oldest samples: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
