package keyring

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/sqlutil"
)

// ReEncryptAll will re-encrypt all encrypted values in the database to the first key in the provided list, using the remaining keys as alternates
// for decryption. This function will return an error if any values fail to re-encrypt.
func ReEncryptAll(ctx context.Context, db *sql.DB, keys Keys) error {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return fmt.Errorf("no keys provided")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer sqlutil.Rollback(ctx, "re-encrypt", tx)

	gdb := gadb.New(tx)

	err = gdb.Keyring_LockKeyrings(ctx)
	if err != nil {
		return fmt.Errorf("lock keyrings: %w", err)
	}

	rings, err := gdb.Keyring_GetKeyringSecrets(ctx)
	if err != nil {
		return fmt.Errorf("get keyring secrets: %w", err)
	}

	for _, ring := range rings {
		sign, signLabel, err := keys.Decrypt(ring.SigningKey)
		if err != nil {
			return fmt.Errorf("decrypt signing key for '%s': %w", ring.ID, err)
		}
		next, nextLabel, err := keys.Decrypt(ring.NextKey)
		if err != nil {
			return fmt.Errorf("decrypt next key for '%s': %w", ring.ID, err)
		}
		encSign, err := keys.Encrypt(signLabel, sign)
		if err != nil {
			return fmt.Errorf("encrypt signing key for '%s': %w", ring.ID, err)
		}
		encNext, err := keys.Encrypt(nextLabel, next)
		if err != nil {
			return fmt.Errorf("encrypt next key for '%s': %w", ring.ID, err)
		}
		err = gdb.Keyring_UpdateKeyringSecrets(ctx, gadb.Keyring_UpdateKeyringSecretsParams{
			ID:         ring.ID,
			SigningKey: encSign,
			NextKey:    encNext,
		})
		if err != nil {
			return fmt.Errorf("update keyring secrets for '%s': %w", ring.ID, err)
		}
	}

	err = gdb.Keyring_LockConfig(ctx)
	if err != nil {
		return fmt.Errorf("lock config: %w", err)
	}

	cfgs, err := gdb.Keyring_GetConfigPayloads(ctx)
	if err != nil {
		return fmt.Errorf("get config payloads: %w", err)
	}

	for _, cfg := range cfgs {
		dec, label, err := keys.Decrypt(cfg.Data)
		if err != nil {
			return fmt.Errorf("decrypt config payload for config #%d: %w", cfg.ID, err)
		}
		enc, err := keys.Encrypt(label, dec)
		if err != nil {
			return fmt.Errorf("encrypt config payload for #%d: %w", cfg.ID, err)
		}
		err = gdb.Keyring_UpdateConfigPayload(ctx, gadb.Keyring_UpdateConfigPayloadParams{
			ID:   cfg.ID,
			Data: enc,
		})
		if err != nil {
			return fmt.Errorf("update config payload for #%d: %w", cfg.ID, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
