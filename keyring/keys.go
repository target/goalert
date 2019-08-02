package keyring

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/pem"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/target/goalert/util"
	"golang.org/x/crypto/pbkdf2"
)

type KeyStore interface {
	Encrypt(label string, data []byte) ([]byte, error)
	Decrypt(pemData []byte) (data []byte, err error)
}

type DBKeyStore struct {
	db *sql.DB

	findOne *sql.Stmt
	findAll *sql.Stmt
	insert  *sql.Stmt
	lock    *sql.Stmt

	keys []masterKey
}

type masterKey struct {
	ID         string
	Version    int
	Active     bool
	Digest     []byte
	DigestSalt []byte
	DigestIter int

	key []byte
}

func NewKeyStore(ctx context.Context, db *sql.DB, passphrases []string) (KeyStore, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	ks := &DBKeyStore{
		db: db,

		findOne: p.P(`
			select
				id,
				version,
				coalese(active, false),
				mk_digest,
				mk_digest_salt,
				mk_digest_iter
			from data_encryption_key_metadata
			where id = $1
			`),
		findAll: p.P(`
			select
				id,
				version,
				coalese(active, false),
				mk_digest,
				mk_digest_salt,
				mk_digest_iter
			from data_encryption_key_metadata
			`),
		insert: p.P(`
			insert into data_encryption_key_metadata (
				id,
				version,
				active,
				mk_digest,
				mk_digest_salt,
				mk_digest_iter
			) values (
				$1, $2, $3, $4, $5, $6
			)
		`),
		lock: p.P(`lock data_encryption_key_metadata`),
	}
	if p.Err != nil {
		return nil, p.Err
	}

	err := ks.init(ctx, passphrases)
	if err != nil {
		return nil, err
	}
	return ks, nil
}

func (ks *DBKeyStore) init(ctx context.Context, passphrases []string) error {
	keys, err := ks.findAllTx(ctx, nil)
	if err == sql.ErrNoRows {
		keys, err = ks.insertKeys(ctx, passphrases)
	}
	if err != nil {
		return err
	}
	ks.keys = keys

	return nil
}

func (ks *DBKeyStore) insertKeys(ctx context.Context, passphrases []string) ([]masterKey, error) {
	tx, err := ks.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.StmtContext(ctx, ks.lock).ExecContext(ctx)
	if err != nil {
		return nil, err
	}

	keys, err := ks.findAllTx(ctx, tx)
	if err != sql.ErrNoRows {
		return keys, err
	}

	insert := tx.StmtContext(ctx, ks.insert)
	var newKeys []masterKey

	for i, p := range passphrases {
		m := masterKey{
			ID:         uuid.NewV4().String(),
			Active:     i == 0,
			DigestIter: 4096,

			key: []byte(p),
		}

		m.DigestSalt, err = newMasterKey(32)
		if err != nil {
			return nil, err
		}

		m.Digest = pbkdf2.Key(m.key, m.DigestSalt, m.DigestIter, 32, sha256.New)
		_, err = insert.ExecContext(ctx, m.ID, m.Version, m.Active, m.Digest, m.DigestSalt, m.DigestIter)
		if err != nil {
			return nil, err
		}
	}

	return newKeys, tx.Commit()
}

func (m *masterKey) scanFrom(fn func(...interface{}) error) error {
	return fn(&m.ID, &m.Version, &m.Active, &m.Digest, &m.DigestSalt, &m.DigestIter)
}

func (ks *DBKeyStore) findAllTx(ctx context.Context, tx *sql.Tx) ([]masterKey, error) {
	stmt := ks.findAll

	if tx != nil {
		stmt = tx.StmtContext(ctx, ks.findAll)
	}

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []masterKey

	for rows.Next() {
		var m masterKey
		err = m.scanFrom(rows.Scan)

		if err != nil {
			return nil, err
		}
		keys = append(keys, m)
	}

	return keys, nil
}

func (ks *DBKeyStore) activeKey() *masterKey {
	for _, key := range ks.keys {
		if key.Active {
			return &key
		}
	}
	return nil
}

func (ks *DBKeyStore) key(id string) *masterKey {
	for _, key := range ks.keys {
		if key.ID == id {
			return &key
		}
	}
	return nil
}

func (ks *DBKeyStore) Encrypt(label string, data []byte) ([]byte, error) {
	key := ks.activeKey()
	if key == nil {
		return nil, errors.New("no available encryption key")
	}

	if key.Version > 1 {
		return nil, errors.New("unsupported encryption key version")
	}

	block, err := x509.EncryptPEMBlock(rand.Reader, label, data, key.key, x509.PEMCipherAES256)
	if err != nil {
		return nil, err
	}
	block.Headers["KeyID"] = key.ID

	data = pem.EncodeToMemory(block)
	return data, nil
}

func (ks *DBKeyStore) Decrypt(pemData []byte) (data []byte, err error) {
	block, _ := pem.Decode(pemData)
	keyID := block.Headers["KeyID"]

	if keyID != "" {
		key := ks.key(keyID)
		if key == nil {
			return nil, errors.New("no available decryption key")
		}
		return x509.DecryptPEMBlock(block, key.key)
	}

	for _, key := range ks.keys {
		data, err = x509.DecryptPEMBlock(block, key.key)
		if err == nil {
			return data, nil
		}
	}

	return nil, errors.New("invalid decryption key")
}

func newMasterKey(size int) ([]byte, error) {
	buf := make([]byte, size)
	_, err := rand.Read(buf)

	if err != nil {
		return nil, err
	}
	return buf, nil
}
