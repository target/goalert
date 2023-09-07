package apikey

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
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
	Policy GQLPolicy
}

func NewStore(ctx context.Context, db *sql.DB, key keyring.Keyring) (*Store, error) {
	s := &Store{db: db, key: key, policies: make(map[uuid.UUID]*policyInfo)}

	return s, nil
}

const Issuer = "goalert"
const Audience = "apikey-v1/graphql-v1"

type APIKeyInfo struct {
	ID            uuid.UUID
	Name          string
	Description   string
	ExpiresAt     time.Time
	LastUsed      *APIKeyUsage
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CreatedBy     *uuid.UUID
	UpdatedBy     *uuid.UUID
	AllowedFields []string
}

func (s *Store) FindAllAdminGraphQLKeys(ctx context.Context) ([]APIKeyInfo, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return nil, err
	}

	keys, err := gadb.New(s.db).APIKeyList(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]APIKeyInfo, 0, len(keys))
	for _, k := range keys {
		k := k

		var p GQLPolicy
		err = json.Unmarshal(k.Policy, &p)
		if err != nil {
			log.Log(ctx, fmt.Errorf("invalid policy for key %s: %w", k.ID, err))
			continue
		}
		if p.Version != 1 {
			log.Log(ctx, fmt.Errorf("unknown policy version for key %s: %d", k.ID, p.Version))
			continue
		}

		var lastUsed *APIKeyUsage
		if k.LastUsedAt.Valid {
			var ip string
			if k.LastIpAddress.Valid {
				ip = k.LastIpAddress.IPNet.IP.String()
			}
			lastUsed = &APIKeyUsage{
				UserAgent: k.LastUserAgent.String,
				IP:        ip,
				Time:      k.LastUsedAt.Time,
			}
		}

		res = append(res, APIKeyInfo{
			ID:            k.ID,
			Name:          k.Name,
			Description:   k.Description,
			ExpiresAt:     k.ExpiresAt,
			LastUsed:      lastUsed,
			CreatedAt:     k.CreatedAt,
			UpdatedAt:     k.UpdatedAt,
			CreatedBy:     &k.CreatedBy.UUID,
			UpdatedBy:     &k.UpdatedBy.UUID,
			AllowedFields: p.AllowedFields,
		})
	}

	return res, nil
}

type APIKeyUsage struct {
	UserAgent string
	IP        string
	Time      time.Time
}

type UpdateKey struct {
	ID          uuid.UUID
	Name        string
	Description string
}

func (s *Store) UpdateAdminGraphQLKey(ctx context.Context, id uuid.UUID, name, desc *string) error {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return err
	}

	if name != nil {
		err = validate.IDName("Name", *name)
	}
	if desc != nil {
		err = validate.Many(err, validate.Text("Description", *desc, 0, 255))
	}
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer sqlutil.Rollback(ctx, "UpdateAdminGraphQLKey", tx)

	key, err := gadb.New(tx).APIKeyForUpdate(ctx, id)
	if err != nil {
		return err
	}
	if name != nil {
		key.Name = *name
	}
	if desc != nil {
		key.Description = *desc
	}

	var user uuid.NullUUID
	if u, err := uuid.Parse(permission.UserID(ctx)); err == nil {
		user = uuid.NullUUID{UUID: u, Valid: true}
	}

	err = gadb.New(tx).APIKeyUpdate(ctx, gadb.APIKeyUpdateParams{
		ID:          id,
		Name:        key.Name,
		Description: key.Description,
		UpdatedBy:   user,
	})
	if err != nil {
		return err
	}

	return tx.Commit()
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

	polData, err := gadb.New(s.db).APIKeyAuthPolicy(ctx, id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Log(ctx, err)
		}
		return ctx, permission.Unauthorized()
	}
	var buf bytes.Buffer
	err = json.Compact(&buf, polData)
	if err != nil {
		log.Logf(ctx, "apikey: invalid policy: %v", err)
		return ctx, permission.Unauthorized()
	}

	// TODO: cache policy hash by key ID when loading
	policyHash := sha256.Sum256(buf.Bytes())
	if !bytes.Equal(policyHash[:], claims.PolicyHash) {
		log.Logf(ctx, "apikey: policy hash mismatch")
		return ctx, permission.Unauthorized()
	}

	var p GQLPolicy
	err = json.Unmarshal(polData, &p)
	if err != nil || p.Version != 1 {
		log.Logf(ctx, "apikey: invalid policy: %v", err)
		return ctx, permission.Unauthorized()
	}

	ua = validate.SanitizeText(ua, 1024)
	ip, _, _ = net.SplitHostPort(ip)
	ip = validate.SanitizeText(ip, 255)
	params := gadb.APIKeyRecordUsageParams{
		KeyID:     id,
		UserAgent: ua,
	}
	params.IpAddress.IPNet.IP = net.ParseIP(ip)
	params.IpAddress.IPNet.Mask = net.CIDRMask(32, 32)
	if params.IpAddress.IPNet.IP != nil {
		params.IpAddress.Valid = true
	}
	err = gadb.New(s.db).APIKeyRecordUsage(ctx, params)
	if err != nil {
		log.Log(ctx, err)
		// don't fail authorization if we can't record usage
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

	ctx = ContextWithPolicy(ctx, &p)
	return ctx, nil
}

type NewAdminGQLKeyOpts struct {
	Name    string
	Desc    string
	Fields  []string
	Expires time.Time
}

type GQLPolicy struct {
	Version       int
	AllowedFields []string
}

func (s *Store) CreateAdminGraphQLKey(ctx context.Context, opt NewAdminGQLKeyOpts) (uuid.UUID, string, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return uuid.Nil, "", err
	}

	err = validate.Many(
		validate.IDName("Name", opt.Name),
		validate.Text("Description", opt.Desc, 0, 255),
		validate.Range("Fields", len(opt.Fields), 1, len(graphql2.SchemaFields())),
	)
	for i, f := range opt.Fields {
		err = validate.Many(err, validate.OneOf(fmt.Sprintf("Fields[%d]", i), f, graphql2.SchemaFields()...))
	}
	if time.Until(opt.Expires) <= 0 {
		err = validate.Many(err, validation.NewFieldError("Expires", "must be in the future"))
	}
	if err != nil {
		return uuid.Nil, "", err
	}

	sort.Strings(opt.Fields)
	policyData, err := json.Marshal(GQLPolicy{
		Version:       1,
		AllowedFields: opt.Fields,
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
