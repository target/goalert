package apikey

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
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

	mx      sync.Mutex
	queries map[string]string
}

func NewStore(ctx context.Context, db *sql.DB, key keyring.Keyring) (*Store, error) {
	s := &Store{db: db, key: key, queries: make(map[string]string)}

	return s, nil
}

func (s *Store) ContextQuery(ctx context.Context) (string, error) {
	src := permission.Source(ctx)
	if src == nil {
		return "", errors.New("no permission source")
	}
	if src.Type != permission.SourceTypeGQLAPIKey {
		return "", errors.New("permission source is not a GQLAPIKey")
	}
	s.mx.Lock()
	q := s.queries[src.ID]
	s.mx.Unlock()
	if q == "" {
		return "", errors.New("no query found for key")
	}
	return q, nil
}

func (s *Store) AuthorizeGraphQL(ctx context.Context, tok string) (context.Context, error) {
	var claims GraphQLClaims
	_, err := s.key.VerifyJWT(tok, &claims, "goalert", "apikey-v1/graphql-v1")
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
	if key.Version != 1 {
		log.Logf(ctx, "apikey: invalid version: %v", key.Version)
		return ctx, permission.Unauthorized()
	}

	var v1 V1
	err = json.Unmarshal(key.Data, &v1)
	if err != nil {
		log.Logf(ctx, "apikey: invalid data: %v", err)
		return ctx, permission.Unauthorized()
	}

	if v1.Type != TypeGraphQLV1 || v1.GraphQLV1 == nil {
		log.Logf(ctx, "apikey: invalid type: %v", v1.Type)
		return ctx, permission.Unauthorized()
	}

	if v1.GraphQLV1.SHA256 != claims.AuthHash {
		log.Log(log.WithField(ctx, "key_id", id.String()), errors.New("apikey: query hash mismatch (claims)"))
		return ctx, permission.Unauthorized()
	}

	hash := sha256.Sum256([]byte(v1.GraphQLV1.Query))
	if hash != claims.AuthHash {
		log.Log(log.WithField(ctx, "key_id", id.String()), errors.New("apikey: query hash mismatch (key)"))
		return ctx, permission.Unauthorized()
	}

	s.mx.Lock()
	s.queries[id.String()] = v1.GraphQLV1.Query
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
		Type: TypeGraphQLV1,
		GraphQLV1: &GraphQLV1{
			Query:  query,
			SHA256: hash,
		},
	})
	if err != nil {
		return nil, err
	}
	tok, err := s.key.SignJWT(NewGraphQLClaims(id, hash, exp))
	if err != nil {
		return nil, err
	}

	err = gadb.New(s.db).APIKeyInsert(ctx, gadb.APIKeyInsertParams{
		ID:        id,
		Name:      name,
		Version:   1,
		ExpiresAt: exp,
		Data:      data,
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
