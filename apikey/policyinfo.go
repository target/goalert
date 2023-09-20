package apikey

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
)

type policyInfo struct {
	Hash   []byte
	Policy GQLPolicy
}

// _fetchPolicyInfo will fetch the policyInfo for the given key.
func (s *Store) _fetchPolicyInfo(ctx context.Context, id uuid.UUID) (*policyInfo, bool, error) {
	polData, err := gadb.New(s.db).APIKeyAuthPolicy(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var info policyInfo
	err = json.Unmarshal(polData, &info.Policy)
	if err != nil {
		return nil, false, err
	}

	h := sha256.Sum256(polData)
	info.Hash = h[:]

	return &info, true, nil
}
