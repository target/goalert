package apikey

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation/validate"
)

type Store struct {
	db  *sql.DB
	key keyring.Keyring

	mx       sync.Mutex
	policies map[uuid.UUID]*policyInfo
}

type policyInfo struct {
	Hash   []byte
	Policy Policy
}

func NewStore(ctx context.Context, db *sql.DB, key keyring.Keyring) (*Store, error) {
	s := &Store{db: db, key: key, policies: make(map[uuid.UUID]*policyInfo)}

	return s, nil
}

const Issuer = "goalert"
const Audience = "apikey-v1/graphql-v1"

func (s *Store) AuthorizeGraphQL(ctx context.Context, tok string) (context.Context, error) {
	var claims Claims
	_, err := s.key.VerifyJWT(tok, &claims, Issuer, Audience)
	if err != nil {
		log.Logf(ctx, "apikey: verify failed: %v", err)
		return ctx, permission.Unauthorized()
	}
	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		log.Logf(ctx, "apikey: invalid subject: %v", err)
		return ctx, permission.Unauthorized()
	}

	key, err := gadb.New(s.db).APIKeyAuth(ctx, id)
	if err != nil {
		log.Logf(ctx, "apikey: lookup failed: %v", err)
		return ctx, permission.Unauthorized()
	}

	// TODO: cache policy hash by key ID when loading
	policyHash := sha256.Sum256(key.Policy)
	if !bytes.Equal(policyHash[:], claims.PolicyHash) {
		log.Logf(ctx, "apikey: policy hash mismatch")
	}

	var p Policy
	err = json.Unmarshal(key.Policy, &p)
	if err != nil {
		log.Logf(ctx, "apikey: invalid policy: %v", err)
		return ctx, permission.Unauthorized()
	}

	if p.Type != PolicyTypeGraphQLV1 || p.GraphQLV1 == nil {
		log.Logf(ctx, "apikey: invalid policy type: %v", p.Type)
		return ctx, permission.Unauthorized()
	}

	s.mx.Lock()
	s.policies[id] = &policyInfo{
		Hash:   policyHash[:],
		Policy: p,
	}
	s.mx.Unlock()

	ctx = permission.SourceContext(ctx, &permission.SourceInfo{
		ID:   id.String(),
		Type: permission.SourceTypeGQLAPIKey,
	})
	ctx = permission.UserContext(ctx, "", permission.RoleUnknown)
	return ctx, nil
}

type Key struct {
	ID        uuid.UUID
	Name      string
	Type      Type
	ExpiresAt time.Time
	Token     string
}

func (s *Store) CreateAdminGraphQLKey(ctx context.Context, name, query string, exp time.Time) (*Key, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return nil, err
	}
	err = validate.IDName("Name", name)
	if err != nil {
		return nil, err
	}
	err = validate.RequiredText("Query", query, 1, 8192)
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256([]byte(query))

	id := uuid.New()

	data, err := json.Marshal(V1{
		Type:      TypeGraphQLV1,
		GraphQLV1: &GraphQLV1{
			// Query:  query,
			// SHA256: hash,
		},
	})
	if err != nil {
		return nil, err
	}
	tok, err := s.key.SignJWT(NewGraphQLClaims(id, hash[:], exp))
	if err != nil {
		return nil, err
	}

	err = gadb.New(s.db).APIKeyInsert(ctx, gadb.APIKeyInsertParams{
		ID:        id,
		Name:      name,
		ExpiresAt: exp,
		Policy:    data,
	})
	if err != nil {
		return nil, err
	}

	return &Key{
		ID:        id,
		Name:      name,
		Type:      TypeGraphQLV1,
		ExpiresAt: exp,
		Token:     tok,
	}, nil
}
