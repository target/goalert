package app

import (
	"context"
	"database/sql"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/target/goalert/config"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
)

func getSetConfig(setCfg bool, data []byte) error {
	if viper.GetBool("verbose") {
		log.EnableVerbose()
	}

	err := viper.ReadInConfig()
	// ignore file not found error
	if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && !ok {
		return errors.Wrap(err, "read config")
	}

	c, err := getConfig()
	if err != nil {
		return err
	}
	db, err := sql.Open("postgres", c.DBURL)
	if err != nil {
		return errors.Wrap(err, "connect to postgres")
	}
	defer db.Close()
	ctx := context.Background()
	ctx = permission.SystemContext(ctx, "SetConfig")
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "start transaction")
	}
	defer tx.Rollback()

	s, err := config.NewStore(ctx, db, c.EncryptionKeys, "")
	if err != nil {
		return errors.Wrap(err, "init config store")
	}
	if setCfg {
		id, err := s.SetConfigData(ctx, tx, data)
		if err != nil {
			return errors.Wrap(err, "save config")
		}

		err = tx.Commit()
		if err != nil {
			return errors.Wrap(err, "commit changes")
		}
		log.Logf(ctx, "Saved config version %d", id)
		return nil
	}

	_, _, data, err = s.ConfigData(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "read config")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "commit")
	}

	_, err = os.Stdout.Write(data)
	return err
}
