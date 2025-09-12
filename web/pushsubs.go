package web

import (
    "encoding/json"
    "sync"
)

// pushSubscription represents the minimal fields we care about for a Push API subscription.
// It accepts and preserves any additional fields via Raw for potential future use.
type pushSubscription struct {
    Endpoint string `json:"endpoint"`
    Raw      json.RawMessage
}

type pushStore struct {
    mx   sync.RWMutex
    subs map[string]json.RawMessage // key: endpoint
}

func newPushStore() *pushStore {
    return &pushStore{subs: make(map[string]json.RawMessage)}
}

func (s *pushStore) add(sub pushSubscription) {
    if sub.Endpoint == "" || len(sub.Raw) == 0 {
        return
    }
    s.mx.Lock()
    s.subs[sub.Endpoint] = sub.Raw
    s.mx.Unlock()
}

