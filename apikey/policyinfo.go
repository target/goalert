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

func parsePolicyInfo(data []byte) (*policyInfo, error) {
	var info policyInfo
	err := json.Unmarshal(data, &info.Policy)
	if err != nil {
		return nil, err
	}

	// re-encode policy to get a consistent hash
	data, err = json.Marshal(info.Policy)
	if err != nil {
		return nil, err
	}

	h := sha256.Sum256(data)
	info.Hash = h[:]

	return &info, nil
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

	info, err := parsePolicyInfo(polData)
	if err != nil {
		return nil, false, err
	}

	return info, true, nil
}
