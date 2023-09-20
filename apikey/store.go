package apikey

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

type Store struct {
	db  *sql.DB
	key keyring.Keyring

	polCache      *polCache
	lastUsedCache *lastUsedCache
}

func NewStore(ctx context.Context, db *sql.DB, key keyring.Keyring) (*Store, error) {
	s := &Store{
		db:  db,
		key: key,
	}

	s.polCache = newPolCache(polCacheConfig{
		FillFunc: s._fetchPolicyInfo,
		Verify:   s._verifyPolicyID,
		MaxSize:  1000,
	})

	s.lastUsedCache = newLastUsedCache(1000, s._updateLastUsed)

	return s, nil
}

func (s *Store) DeleteAdminGraphQLKey(ctx context.Context, id uuid.UUID) error {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return err
	}

	return gadb.New(s.db).APIKeyDelete(ctx, id)
}

func (s *Store) AuthorizeGraphQL(ctx context.Context, tok, ua, ip string) (context.Context, error) {
	var claims Claims
	_, err := s.key.VerifyJWT(tok, &claims, Issuer, Audience)
	if err != nil {
		return ctx, permission.Unauthorized()
	}
	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		log.Logf(ctx, "apikey: invalid subject: %v", err)
		return ctx, permission.Unauthorized()
	}

	info, valid, err := s.polCache.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !valid {
		// Successful negative cache lookup, we return Unauthorized because althought the token was validated, the key was revoked/removed.
		return ctx, permission.Unauthorized()
	}
	if !bytes.Equal(info.Hash, claims.PolicyHash) {
		// Successful cache lookup, but the policy has changed since the token was issued and so the token is no longer valid.
		s.polCache.Revoke(ctx, id)

		// We want to log this as a warning, because it is a potential security issue.
		log.Log(ctx, fmt.Errorf("apikey: policy hash mismatch for key %s", id))
		return ctx, permission.Unauthorized()
	}

	err = s.lastUsedCache.RecordUsage(ctx, id, ua, ip)
	if err != nil {
		// Recording usage is not critical, so we log the error and continue.
		log.Log(ctx, err)
	}

	ctx = permission.SourceContext(ctx, &permission.SourceInfo{
		ID:   id.String(),
		Type: permission.SourceTypeGQLAPIKey,
	})
	ctx = permission.UserContext(ctx, "", info.Policy.Role)

	ctx = ContextWithPolicy(ctx, &info.Policy)
	return ctx, nil
}

// NewAdminGQLKeyOpts is used to create a new GraphQL API key.
type NewAdminGQLKeyOpts struct {
	Name    string
	Desc    string
	Fields  []string
	Expires time.Time
	Role    permission.Role
}

// CreateAdminGraphQLKey will create a new GraphQL API key returning the ID and token.
func (s *Store) CreateAdminGraphQLKey(ctx context.Context, opt NewAdminGQLKeyOpts) (uuid.UUID, string, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return uuid.Nil, "", err
	}

	err = validate.Many(
		validate.IDName("Name", opt.Name),
		validate.Text("Description", opt.Desc, 0, 255),
		validate.Range("Fields", len(opt.Fields), 1, len(graphql2.SchemaFields())),
		validate.OneOf("Role", opt.Role, permission.RoleAdmin, permission.RoleUser),
	)
	if time.Until(opt.Expires) <= 0 {
		err = validate.Many(err, validation.NewFieldError("Expires", "must be in the future"))
	}
	for i, f := range opt.Fields {
		if slices.Contains(graphql2.SchemaFields(), f) {
			continue
		}

		err = validate.Many(err, validation.NewFieldError(fmt.Sprintf("Fields[%d]", i), "is not a valid field"))
	}
	if err != nil {
		return uuid.Nil, "", err
	}

	sort.Strings(opt.Fields)
	policyData, err := json.Marshal(GQLPolicy{
		Version:       1,
		AllowedFields: opt.Fields,
		Role:          opt.Role,
	})
	if err != nil {
		return uuid.Nil, "", err
	}

	var user uuid.NullUUID
	userID, err := uuid.Parse(permission.UserID(ctx))
	if err == nil {
		user = uuid.NullUUID{UUID: userID, Valid: true}
	}

	id := uuid.New()
	err = gadb.New(s.db).APIKeyInsert(ctx, gadb.APIKeyInsertParams{
		ID:          id,
		Name:        opt.Name,
		Description: opt.Desc,
		ExpiresAt:   opt.Expires,
		Policy:      policyData,
		CreatedBy:   user,
		UpdatedBy:   user,
	})
	if err != nil {
		return uuid.Nil, "", err
	}

	hash := sha256.Sum256([]byte(policyData))
	tok, err := s.key.SignJWT(NewGraphQLClaims(id, hash[:], opt.Expires))
	if err != nil {
		return uuid.Nil, "", err
	}

	return id, tok, nil
}
