package statusupdatemanager

import (
	"context"
	"fmt"

	"github.com/target/goalert/permission"
)

// UpdateAll will update all schedule rules.
func (db *DB) UpdateAll(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}

	err = db.updateUsers(ctx, true, nil)
	if err != nil {
		return err
	}
	fmt.Println("trying to update channels!")
	err = db.updateChannels(ctx, true, nil)
	if err != nil {
		fmt.Println("made it into updateChannels:", err)
		return err
	}

	return nil
}
