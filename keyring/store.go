package keyring

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"crypto/x509"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"math/big"
	"sync"
	"time"

	"github.com/target/goalert/util"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

func init() {
	jwt.RegisterSigningMethod("ES224", func() jwt.SigningMethod {
		return &jwt.SigningMethodECDSA{
			Name:      "ES224",
			Hash:      crypto.SHA512_224,
			KeySize:   28,
			CurveBits: 224,
		}
	})
}

// A Keyring allows signing and verifying messages.
type Keyring interface {
	RotateKeys(ctx context.Context) error

	Sign(p []byte) ([]byte, error)
	Verify(p []byte, signature []byte) (valid, oldKey bool)

	SignJWT(jwt.Claims) (string, error)
	VerifyJWT(token string, c jwt.Claims, iss, aud string) (bool, error)

	Shutdown(context.Context) error
}

var _ Keyring = &DB{}

type header struct {
	Version  byte
	KeyIndex byte
}

type v1Signature struct {
	RLen, SLen byte
	R          [28]byte
	S          [28]byte
}

// Config allows specifying operational parameters of a keyring.
type Config struct {
	// Name is the unique identifier of this keyring.
	Name string

	// RotationDays is the number of days between automatic rotations. If zero, automatic rotation is disabled.
	RotationDays int

	// MaxOldKeys determines how many old keys (1-254) are kept for validation. This value, multiplied by RotationDays
	// determines the minimum amount of time a signature remains valid.
	MaxOldKeys int

	// Keys specifies a set of keys to use for encrypting and decrypting the private key.
	Keys Keys
}

// DB implements a Keyring using postgres as the datastore.
type DB struct {
	logger *log.Logger

	db *sql.DB

	cfg Config

	verificationKeys map[byte]ecdsa.PublicKey
	signingKey       *ecdsa.PrivateKey
	rotationCount    int

	mx          sync.RWMutex
	shutdown    chan context.Context
	forceRotate chan chan error

	fetchKeys  *sql.Stmt
	setKeys    *sql.Stmt
	txTime     *sql.Stmt
	insertKeys *sql.Stmt
}

func marshalVerificationKeys(keys map[byte]ecdsa.PublicKey) ([]byte, error) {
	m := make(map[byte][]byte, len(keys))
	var err error
	for id, key := range keys {
		m[id], err = x509.MarshalPKIXPublicKey(&key)
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal(m)
}

func parseVerificationKeys(data []byte) (map[byte]ecdsa.PublicKey, error) {
	var m map[byte][]byte
	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	res := make(map[byte]ecdsa.PublicKey, len(m))
	for id, data := range m {
		key, err := x509.ParsePKIXPublicKey(data)
		if err != nil {
			// ignore broken keys for verification
			continue
		}
		if k, ok := key.(*ecdsa.PublicKey); ok {
			res[id] = *k
		}
	}

	return res, nil
}

// NewDB creates a new postgres-backed keyring.
func NewDB(ctx context.Context, logger *log.Logger, db *sql.DB, cfg *Config) (*DB, error) {
	if cfg == nil {
		cfg = &Config{Name: "default"}
	}
	if cfg.MaxOldKeys == 0 {
		cfg.MaxOldKeys = 1
	}
	err := validate.Many(
		validate.IDName("Name", cfg.Name),

		// keyspace is 256 (1 byte); need 1 for current key, and 1 for next key leaving 254 possible slots for old ones
		validate.Range("MaxOldKeys", cfg.MaxOldKeys, 1, 254),

		validate.Range("RotationDays", cfg.RotationDays, 0, 9000),
	)

	if err != nil {
		return nil, err
	}
	p := &util.Prepare{DB: db, Ctx: ctx}
	d := &DB{
		db:  db,
		cfg: *cfg,

		logger: logger,

		forceRotate: make(chan chan error),
		shutdown:    make(chan context.Context),

		txTime: p.P(`select now()`),
		insertKeys: p.P(`
			insert into keyring (
				id,
				verification_keys,
				signing_key,
				next_key,
				next_rotation,
				rotation_count
			) values (
				$1, $2, $3, $4, $5, 0
			)
			on conflict do nothing
		`),
		fetchKeys: p.P(`
			select
				verification_keys,
				signing_key,
				next_key,
				now(),
				next_rotation,
				rotation_count
			from keyring
			where id = $1
			for update
		`),
		setKeys: p.P(`
			update keyring
			set
				verification_keys = $2,
				signing_key = $3,
				next_key = $4,
				next_rotation = $5,
				rotation_count = $6
			where id = $1
		`),
	}

	if p.Err != nil {
		return nil, p.Err
	}

	err = d.refreshAndRotateKeys(ctx, false)
	if err != nil {
		return nil, err
	}

	go d.loop()
	return d, nil
}

// Shutdown allows gracefully shutting down the keyring (e.g. auto rotations) after
// finishing any in-progress rotations.
func (db *DB) Shutdown(ctx context.Context) error {
	if db == nil {
		return nil
	}
	db.shutdown <- ctx

	// wait for it to complete
	<-db.shutdown
	return nil
}

func (db *DB) loop() {
	t := time.NewTicker(12 * time.Hour)
	var shutdownCtx context.Context
	defer close(db.shutdown)
mainLoop:
	for {
		select {
		case <-t.C:
			ctx, cancel := context.WithTimeout(db.logger.BackgroundContext(), time.Minute)
			err := db.refreshAndRotateKeys(ctx, false)
			cancel()
			if err != nil {
				log.Log(ctx, err)
			}
		case shutdownCtx = <-db.shutdown:
			break mainLoop
		case ch := <-db.forceRotate:
			ctx, cancel := context.WithTimeout(db.logger.BackgroundContext(), time.Minute)
			ch <- db.refreshAndRotateKeys(ctx, true)
			cancel()
		}
	}

	// respond to any pending force rotation calls
	close(db.forceRotate)
	for ch := range db.forceRotate {
		ctx, cancel := context.WithTimeout(shutdownCtx, time.Minute)
		ch <- db.refreshAndRotateKeys(ctx, true)
		cancel()
	}
}

func (db *DB) newKey() (*ecdsa.PrivateKey, []byte, error) {
	key, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	data, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, nil, err
	}
	data, err = db.cfg.Keys.Encrypt("ECDSA PRIVATE KEY", data)
	if err != nil {
		return nil, nil, err
	}
	return key, data, nil
}

func (db *DB) loadKey(encData []byte) (*ecdsa.PrivateKey, error) {
	data, _, err := db.cfg.Keys.Decrypt(encData)
	if err != nil {
		return nil, err
	}

	key, err := x509.ParseECPrivateKey(data)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (db *DB) commitNewKeyring(ctx context.Context, tx *sql.Tx) error {
	var t time.Time
	err := tx.Stmt(db.txTime).QueryRowContext(ctx).Scan(&t)
	if err != nil {
		return err
	}
	signKey, signData, err := db.newKey()
	if err != nil {
		return err
	}
	nextKey, nextData, err := db.newKey()
	if err != nil {
		return err
	}

	v := map[byte]ecdsa.PublicKey{
		0: signKey.PublicKey,
		1: nextKey.PublicKey,
	}

	vData, err := marshalVerificationKeys(v)
	if err != nil {
		return err
	}

	var nextRotTime interface{}
	if db.cfg.RotationDays > 0 {
		// We want to wait an explicit amount of time, rather than rotating by date.
		//
		// Specifically, if multiple instances of GoAlert happen to run on systems of differing
		// timezones, they should be able to agree on handoff times.
		nextRotTime = t.Add(time.Hour * 24 * time.Duration(db.cfg.RotationDays))
	}

	res, err := tx.Stmt(db.insertKeys).ExecContext(ctx, db.cfg.Name, vData, signData, nextData, nextRotTime)
	if err != nil {
		return err
	}
	rowCount, err := res.RowsAffected()
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	var rotationCount int

	if rowCount == 0 {
		// failed to insert the new data, so scan old & refresh
		var vKeysData, signKeyData, nextKeyData []byte
		var rotateT sql.NullTime
		err = db.fetchKeys.QueryRowContext(ctx, db.cfg.Name).Scan(&vKeysData, &signKeyData, &nextKeyData, &t, &rotateT, &rotationCount)
		if err != nil {
			return err
		}

		v, err = parseVerificationKeys(vKeysData)
		if err != nil {
			return err
		}

		signKey, err = db.loadKey(signKeyData)
		if err != nil {
			// if we can't get the sign key -- we will at least move forward with the verification keys
			log.Log(ctx, errors.Wrap(err, "load signing key"))
		}
	}

	db.mx.Lock()
	defer db.mx.Unlock()

	db.verificationKeys = v
	db.signingKey = signKey
	db.rotationCount = rotationCount

	return nil
}

func (db *DB) rotateVerificationKeys(m map[byte]ecdsa.PublicKey, n int, newKey ecdsa.PublicKey) map[byte]ecdsa.PublicKey {
	newM := make(map[byte]ecdsa.PublicKey, len(m)+1)
	for i := n - db.cfg.MaxOldKeys; i <= n; i++ {
		if key, ok := m[byte(i)]; ok {
			newM[byte(i)] = key
		}
	}
	newM[byte(n+1)] = newKey
	return newM
}

// RotateKeys will force a key rotation.
func (db *DB) RotateKeys(ctx context.Context) error {
	ch := make(chan error)
	db.forceRotate <- ch
	return <-ch
}

// refreshAndRotateKeys will perform a key rotation, and cleanup expired keys when appropriate. If forceRotation
// is true, a rotation will always happen -- even if RotationDays is zero (disabled). It also
// ensures the current key configuration is up-to-date.
//
// When a key is rotated, a new key is generated and inserted.
func (db *DB) refreshAndRotateKeys(ctx context.Context, forceRotation bool) error {
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer sqlutil.Rollback(ctx, "keyring: rotate keys", tx)

	row := tx.Stmt(db.fetchKeys).QueryRowContext(ctx, db.cfg.Name)

	var verificationKeys map[byte]ecdsa.PublicKey

	var vKeysData, signKeyData, nextKeyData []byte
	var t time.Time
	var rotateT sql.NullTime
	var count int
	err = row.Scan(&vKeysData, &signKeyData, &nextKeyData, &t, &rotateT, &count)
	if errors.Is(err, sql.ErrNoRows) {
		return db.commitNewKeyring(ctx, tx)
	}
	if err != nil {
		return err
	}

	verificationKeys, err = parseVerificationKeys(vKeysData)
	if err != nil {
		return errors.Wrap(err, "unmarshal verification keys")
	}

	if forceRotation || (rotateT.Valid && !t.Before(rotateT.Time)) {
		// perform a key rotation
		signKeyData = nextKeyData
		var nextKey *ecdsa.PrivateKey
		nextKey, nextKeyData, err = db.newKey()
		if err != nil {
			return err
		}
		count++
		verificationKeys = db.rotateVerificationKeys(verificationKeys, count, nextKey.PublicKey)
		vKeysData, err = marshalVerificationKeys(verificationKeys)
		if err != nil {
			return err
		}
		var nextRotTime interface{}
		if db.cfg.RotationDays > 0 {
			// We want to wait an explicit amount of time, rather than rotating by date.
			//
			// Specifically, if multiple instances of GoAlert happen to run on systems of differing
			// timezones, they should be able to agree on handoff times.
			nextRotTime = t.Add(time.Hour * 24 * time.Duration(db.cfg.RotationDays))
		}
		_, err := tx.Stmt(db.setKeys).ExecContext(ctx, db.cfg.Name, vKeysData, signKeyData, nextKeyData, nextRotTime, count)
		if err != nil {
			return err
		}
		err = tx.Commit()
		if err != nil {
			return err
		}
	}

	signKey, err := db.loadKey(signKeyData)
	if err != nil {
		log.Log(ctx, errors.Wrap(err, "load signing key"))
	}

	db.mx.Lock()
	defer db.mx.Unlock()

	db.verificationKeys = verificationKeys
	db.signingKey = signKey
	db.rotationCount = count

	return nil
}

func (db *DB) SignJWT(c jwt.Claims) (string, error) {
	db.mx.RLock()
	defer db.mx.RUnlock()

	if db.signingKey == nil {
		return "", errors.New("signing key unavailable")
	}

	tok := jwt.NewWithClaims(jwt.GetSigningMethod("ES224"), c)
	tok.Header["key"] = byte(db.rotationCount % 256)

	return tok.SignedString(db.signingKey)
}

// Sign will sign a message and return the signature.
func (db *DB) Sign(p []byte) ([]byte, error) {
	db.mx.RLock()
	defer db.mx.RUnlock()

	if db.signingKey == nil {
		return nil, errors.New("signing key unavailable")
	}

	hdr := header{
		Version:  1, // v1 is latest
		KeyIndex: byte(db.rotationCount % 256),
	}

	sum := sha512.Sum512_224(p)
	r, s, err := ecdsa.Sign(rand.Reader, db.signingKey, sum[:])
	if err != nil {
		return nil, err
	}
	var v1sig v1Signature
	v1sig.RLen = byte(len(r.Bytes()))
	v1sig.SLen = byte(len(s.Bytes()))
	copy(v1sig.R[:], r.Bytes())
	copy(v1sig.S[:], s.Bytes())

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, hdr)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.BigEndian, v1sig)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (db *DB) VerifyJWT(s string, c jwt.Claims, iss, aud string) (bool, error) {
	db.mx.RLock()
	defer db.mx.RUnlock()

	var currentKey bool
	_, err := jwt.ParseWithClaims(s, c, func(tok *jwt.Token) (interface{}, error) {
		keyIndex, ok := tok.Header["key"].(float64)
		if !ok {
			return nil, errors.New("invalid key index")
		}
		key, ok := db.verificationKeys[byte(keyIndex)]
		if !ok {
			return nil, errors.New("invalid key")
		}

		currentKey = byte(keyIndex) == byte(db.rotationCount) || byte(keyIndex) == byte(db.rotationCount+1)
		return &key, nil
	},
		jwt.WithValidMethods([]string{"ES224"}),
		jwt.WithIssuer(iss),
		jwt.WithAudience(aud),
	)
	if err != nil {
		return false, err
	}

	return currentKey, nil
}

// Verify will validate the signature and metadata, and optionally length, of a message.
func (db *DB) Verify(p []byte, signature []byte) (valid, oldKey bool) {
	db.mx.RLock()
	defer db.mx.RUnlock()

	buf := bytes.NewBuffer(signature)
	var hdr header
	err := binary.Read(buf, binary.BigEndian, &hdr)
	// The only error here for the bytes.Buffer is if it's too short
	// which just means it's an invalid message.
	if err != nil {
		return false, false
	}

	// only v1 is supported currently
	if hdr.Version != 1 {
		return false, false
	}

	var v1sig v1Signature
	err = binary.Read(buf, binary.BigEndian, &v1sig)
	if err != nil {
		return false, false
	}

	// signature should not include any trailing/extra data
	if buf.Len() != 0 {
		return false, false
	}

	if v1sig.RLen > 28 || v1sig.SLen > 28 {
		return false, false
	}

	key, ok := db.verificationKeys[hdr.KeyIndex]
	if !ok {
		return false, false
	}
	// ensure key exists
	r := big.NewInt(0)
	s := big.NewInt(0)
	r.SetBytes(v1sig.R[:v1sig.RLen])
	s.SetBytes(v1sig.S[:v1sig.SLen])

	sum := sha512.Sum512_224(p)
	valid = ecdsa.Verify(&key, sum[:], r, s)
	if !valid {
		return false, false
	}

	output := make([]byte, buf.Len())
	copy(output, buf.Bytes())
	oldKey = hdr.KeyIndex != byte(db.rotationCount) && hdr.KeyIndex != byte(db.rotationCount+1)
	return true, oldKey
}
