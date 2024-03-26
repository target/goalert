package alert

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/sqlc-dev/pqtype"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

const metaV1 = "alert_meta_v1"

type metadataDBFormat struct {
	Type        string
	AlertMetaV1 map[string]string
}

// Metadata returns the metadata for a single alert.
func (s *Store) Metadata(ctx context.Context, db gadb.DBTX, alertID int) (map[string]string, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}

	md, err := gadb.New(db).AlertMetadata(ctx, int64(alertID))
	if errors.Is(err, sql.ErrNoRows) || !md.Valid {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}

	var doc metadataDBFormat
	err = json.Unmarshal(md.RawMessage, &doc)
	if err != nil {
		return nil, err
	}

	if doc.Type != metaV1 || doc.AlertMetaV1 == nil {
		return nil, errors.New("unsupported metadata type")
	}

	return doc.AlertMetaV1, nil
}

type MetadataAlertID struct {
	// AlertID is the ID of the alert.
	ID   int64
	Meta map[string]string
}

func (s Store) FindManyMetadata(ctx context.Context, db gadb.DBTX, alertIDs []int) ([]MetadataAlertID, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}

	err = validate.Range("AlertIDs", len(alertIDs), 1, maxBatch)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, len(alertIDs))
	for i, id := range alertIDs {
		ids[i] = int64(id)
	}

	rows, err := gadb.New(db).AlertManyMetadata(ctx, ids)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	res := make([]MetadataAlertID, len(rows))
	for i, r := range rows {
		var doc metadataDBFormat
		err = json.Unmarshal(r.Metadata.RawMessage, &doc)
		if err != nil {
			return nil, err
		}

		if doc.Type != metaV1 || doc.AlertMetaV1 == nil {
			return nil, errors.New("unsupported metadata type")
		}

		res[i] = MetadataAlertID{
			ID:   r.AlertID,
			Meta: doc.AlertMetaV1,
		}
	}

	return res, nil
}

func (s Store) SetMetadataTx(ctx context.Context, db gadb.DBTX, alertID int, meta map[string]string) error {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Service)
	if err != nil {
		return err
	}

	err = ValidateMetadata(meta)
	if err != nil {
		return err
	}

	var doc metadataDBFormat
	doc.Type = metaV1
	doc.AlertMetaV1 = meta

	md, err := json.Marshal(&doc)
	if err != nil {
		return err
	}

	rowCount, err := gadb.New(db).AlertSetMetadata(ctx, gadb.AlertSetMetadataParams{
		ID:        int64(alertID),
		ServiceID: permission.ServiceNullUUID(ctx), // only provide service_id restriction if request is from a service
		Metadata:  pqtype.NullRawMessage{Valid: true, RawMessage: json.RawMessage(md)},
	})
	if err != nil {
		return err
	}

	if rowCount == 0 {
		// shouldn't happen, but just in case
		return permission.NewAccessDenied("alert closed, invalid, or wrong service")
	}

	return nil
}
