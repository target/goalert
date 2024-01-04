package alertlog

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
)

type Store struct {
	db *sql.DB

	findAll       *sql.Stmt
	findAllByType *sql.Stmt
	findOne       *sql.Stmt

	lookupCallbackType *sql.Stmt
	lookupIKeyType     *sql.Stmt

	lookupNCTypeName *sql.Stmt
}

func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &Store{
		db: db,
		lookupCallbackType: p.P(`
			select cm."type", ch."type"
			from outgoing_messages log
			left join user_contact_methods cm on cm.id = log.contact_method_id
			left join notification_channels ch on ch.id = log.channel_id
			where log.id = $1
		`),

		lookupNCTypeName: p.P(`
			select "type", name from notification_channels where id = $1
		`),

		lookupIKeyType: p.P(`select "type" from integration_keys where id = $1`),
		findOne: p.P(`
			select
				log.id,
				log.alert_id,
				log.timestamp,
				log.event,
				log.message,
				log.sub_type,
				log.sub_user_id,
				usr.name,
				log.sub_integration_key_id,
				ikey.name,
				log.sub_hb_monitor_id,
				hb.name,
				log.sub_channel_id,
				nc.name,
				log.sub_classifier,
				log.meta
			from alert_logs log
			left join users usr on usr.id = log.sub_user_id
			left join integration_keys ikey on ikey.id = log.sub_integration_key_id
			left join heartbeat_monitors hb on hb.id = log.sub_hb_monitor_id
			left join notification_channels nc on nc.id = log.sub_channel_id
			where log.id = $1
		`),
		findAll: p.P(`
			select
				log.id,
				log.alert_id,
				log.timestamp,
				log.event,
				log.message,
				log.sub_type,
				log.sub_user_id,
				usr.name,
				log.sub_integration_key_id,
				ikey.name,
				log.sub_hb_monitor_id,
				hb.name,
				log.sub_channel_id,
				nc.name,
				log.sub_classifier,
				log.meta
			from alert_logs log
			left join users usr on usr.id = log.sub_user_id
			left join integration_keys ikey on ikey.id = log.sub_integration_key_id
			left join heartbeat_monitors hb on hb.id = log.sub_hb_monitor_id
			left join notification_channels nc on nc.id = log.sub_channel_id
			where log.alert_id = $1
			order by id
		`),
		findAllByType: p.P(`
			select
				log.id,
				log.alert_id,
				log.timestamp,
				log.event,
				log.message,
				log.sub_type,
				log.sub_user_id,
				usr.name,
				log.sub_integration_key_id,
				ikey.name,
				log.sub_hb_monitor_id,
				hb.name,
				log.sub_channel_id,
				nc.name,
				log.sub_classifier,
				log.meta
			from alert_logs log
			left join users usr on usr.id = log.sub_user_id
			left join integration_keys ikey on ikey.id = log.sub_integration_key_id
			left join heartbeat_monitors hb on hb.id = log.sub_hb_monitor_id
			left join notification_channels nc on nc.id = log.sub_channel_id
			where log.alert_id = $1 and log.event = $2
			order by id DESC
			limit 1
		`),
	}, p.Err
}

func (s *Store) MustLog(ctx context.Context, alertID int, _type Type, meta interface{}) {
	s.MustLogTx(ctx, nil, alertID, _type, meta)
}

func (s *Store) MustLogTx(ctx context.Context, tx *sql.Tx, alertID int, _type Type, meta interface{}) {
	err := s.LogTx(ctx, tx, alertID, _type, meta)
	if err != nil {
		log.Log(ctx, errors.Wrap(err, "append alert log"))
	}
}

func (s *Store) LogEPTx(ctx context.Context, tx *sql.Tx, epID string, _type Type, meta *EscalationMetaData) error {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return err
	}

	err = validate.UUID("EscalationPolicyID", epID)
	if err != nil {
		return err
	}

	e, err := s.logEntry(ctx, tx, _type, meta)
	if err != nil {
		return err
	}
	params := gadb.AlertLogInsertEPParams{
		EscalationPolicyID:  uuid.MustParse(epID),
		Event:               gadb.EnumAlertLogEvent(e._type),
		SubType:             gadb.NullEnumAlertLogSubjectType{Valid: e.subject._type != SubjectTypeNone, EnumAlertLogSubjectType: gadb.EnumAlertLogSubjectType(e.subject._type)},
		SubUserID:           e.subject.userID,
		SubIntegrationKeyID: e.subject.integrationKeyID,
		SubHbMonitorID:      e.subject.heartbeatMonitorID,
		SubChannelID:        e.subject.channelID,
		SubClassifier:       e.subject.classifier,
		Meta:                pqtype.NullRawMessage{Valid: e.meta != nil, RawMessage: json.RawMessage(e.meta)},
		Message:             e.message,
	}

	return s.queries(tx).AlertLogInsertEP(ctx, params)
}

func (s *Store) LogServiceTx(ctx context.Context, tx *sql.Tx, serviceID string, _type Type, meta interface{}) error {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return err
	}

	err = validate.UUID("ServiceID", serviceID)
	if err != nil {
		return err
	}
	t := _type
	switch _type {
	case TypeAcknowledged:
		t = _TypeAcknowledgeAll
	case TypeClosed:
		t = _TypeCloseAll
	}
	e, err := s.logEntry(ctx, tx, t, meta)
	if err != nil {
		return err
	}

	params := gadb.AlertLogInsertSvcParams{
		ServiceID:           uuid.NullUUID{Valid: true, UUID: uuid.MustParse(serviceID)},
		Event:               gadb.EnumAlertLogEvent(e._type),
		SubType:             gadb.NullEnumAlertLogSubjectType{Valid: e.subject._type != SubjectTypeNone, EnumAlertLogSubjectType: gadb.EnumAlertLogSubjectType(e.subject._type)},
		SubUserID:           e.subject.userID,
		SubIntegrationKeyID: e.subject.integrationKeyID,
		SubHbMonitorID:      e.subject.heartbeatMonitorID,
		SubChannelID:        e.subject.channelID,
		SubClassifier:       e.subject.classifier,
		Meta:                pqtype.NullRawMessage{Valid: e.meta != nil, RawMessage: json.RawMessage(e.meta)},
		Message:             e.message,
	}

	return s.queries(tx).AlertLogInsertSvc(ctx, params)
}

func (s *Store) LogManyTx(ctx context.Context, tx *sql.Tx, alertIDs []int, _type Type, meta interface{}) error {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return err
	}

	e, err := s.logEntry(ctx, tx, _type, meta)
	if err != nil {
		return err
	}

	var ids []int64
	for _, id := range alertIDs {
		ids = append(ids, int64(id))
	}

	params := gadb.AlertLogInsertManyParams{
		Column1:             ids,
		Event:               gadb.EnumAlertLogEvent(e._type),
		SubType:             gadb.NullEnumAlertLogSubjectType{Valid: e.subject._type != SubjectTypeNone, EnumAlertLogSubjectType: gadb.EnumAlertLogSubjectType(e.subject._type)},
		SubUserID:           e.subject.userID,
		SubIntegrationKeyID: e.subject.integrationKeyID,
		SubHbMonitorID:      e.subject.heartbeatMonitorID,
		SubChannelID:        e.subject.channelID,
		SubClassifier:       e.subject.classifier,
		Meta:                pqtype.NullRawMessage{Valid: e.meta != nil, RawMessage: json.RawMessage(e.meta)},
		Message:             e.message,
	}

	return s.queries(tx).AlertLogInsertMany(ctx, params)
}

func (s *Store) queries(tx *sql.Tx) *gadb.Queries {
	if tx != nil {
		return gadb.New(tx)
	}
	return gadb.New(s.db)
}

func (s *Store) LogTx(ctx context.Context, tx *sql.Tx, alertID int, _type Type, meta interface{}) error {
	return s.LogManyTx(ctx, tx, []int{alertID}, _type, meta)
}

func txWrap(ctx context.Context, tx *sql.Tx, stmt *sql.Stmt) *sql.Stmt {
	if tx == nil {
		return stmt
	}
	return tx.StmtContext(ctx, stmt)
}

func (s *Store) logEntry(ctx context.Context, tx *sql.Tx, _type Type, meta interface{}) (*Entry, error) {
	var classExtras []string
	switch _type {
	case _TypeAcknowledgeAll:
		classExtras = append(classExtras, "Ack-All")
		_type = TypeAcknowledged
	case _TypeCloseAll:
		classExtras = append(classExtras, "Close-All")
		_type = TypeClosed
	}

	var r Entry
	r._type = _type
	var err error
	if meta != nil {
		r.meta, err = json.Marshal(meta)
		if err != nil {
			return nil, err
		}
	}

	src := permission.Source(ctx)
	if src != nil {
		switch src.Type {
		case permission.SourceTypeNotificationChannel:
			r.subject._type = SubjectTypeChannel
			var ncType notificationchannel.Type
			var name string
			err = txWrap(ctx, tx, s.lookupNCTypeName).QueryRowContext(ctx, src.ID).Scan(&ncType, &name)
			if err != nil {
				return nil, errors.Wrap(err, "lookup contact method type for callback ID")
			}

			switch ncType {
			case notificationchannel.TypeSlackChan:
				r.subject.classifier = "Slack"
			case notificationchannel.TypeWebhook:
				r.subject.classifier = "Webhook"
			}
			r.subject.channelID.UUID = uuid.MustParse(src.ID)
			r.subject.channelID.Valid = true
		case permission.SourceTypeAuthProvider:
			r.subject.classifier = "Web"
			r.subject._type = SubjectTypeUser

			if permission.UserID(ctx) != "" {
				r.subject.userID.UUID = uuid.MustParse(permission.UserID(ctx))
				r.subject.userID.Valid = true
			}
		case permission.SourceTypeContactMethod:
			r.subject._type = SubjectTypeUser
			if permission.UserID(ctx) != "" {
				r.subject.userID.UUID = uuid.MustParse(permission.UserID(ctx))
				r.subject.userID.Valid = true
			}
			if _type == TypeNoNotificationSent {
				// no CMID for no notification sent
				r.subject.classifier = "no immediate rule"
				break
			}
			cmType, err := s.queries(tx).AlertLogLookupCMType(ctx, uuid.MustParse(src.ID))
			if err != nil {
				return nil, errors.Wrap(err, "lookup contact method type for callback ID")
			}
			switch cmType {
			case gadb.EnumUserContactMethodTypeVOICE:
				r.subject.classifier = "Voice"
			case gadb.EnumUserContactMethodTypeSMS:
				r.subject.classifier = "SMS"
			case gadb.EnumUserContactMethodTypeEMAIL:
				r.subject.classifier = "Email"
			case gadb.EnumUserContactMethodTypeWEBHOOK:
				r.subject.classifier = "Webhook"
			case gadb.EnumUserContactMethodTypeSLACKDM:
				r.subject.classifier = "Slack"
			}

		case permission.SourceTypeNotificationCallback:
			r.subject._type = SubjectTypeUser
			var dt notification.ScannableDestType
			err = txWrap(ctx, tx, s.lookupCallbackType).QueryRowContext(ctx, src.ID).Scan(&dt.CM, &dt.NC)
			if err != nil {
				return nil, errors.Wrap(err, "lookup notification type for callback ID")
			}
			switch dt.DestType() {
			case notification.DestTypeVoice:
				r.subject.classifier = "Voice"
			case notification.DestTypeSMS:
				r.subject.classifier = "SMS"
			case notification.DestTypeUserEmail:
				r.subject.classifier = "Email"
			case notification.DestTypeChanWebhook:
				fallthrough
			case notification.DestTypeUserWebhook:
				r.subject.classifier = "Webhook"
			case notification.DestTypeSlackChannel:
				r.subject.classifier = "Slack"
			}
			if permission.UserID(ctx) != "" {
				r.subject.userID.UUID = uuid.MustParse(permission.UserID(ctx))
				r.subject.userID.Valid = true
			}

		case permission.SourceTypeHeartbeat:
			r.subject._type = SubjectTypeHeartbeatMonitor
			minutes, err := s.queries(tx).AlertLogHBIntervalMinutes(ctx, uuid.MustParse(src.ID))
			if err != nil {
				return nil, errors.Wrap(err, "lookup heartbeat monitor interval by ID")
			}
			if r.Type() == TypeCreated {
				s := "s"
				if minutes == 1 {
					s = ""
				}
				r.subject.classifier = fmt.Sprintf("expired after %d minute"+s, minutes)
			} else if r.Type() == TypeClosed {
				r.subject.classifier = "healthy"
			}

			r.subject.heartbeatMonitorID.Valid = true
			r.subject.heartbeatMonitorID.UUID = uuid.MustParse(src.ID)
		case permission.SourceTypeIntegrationKey:
			r.subject._type = SubjectTypeIntegrationKey
			var ikeyType integrationkey.Type
			err = txWrap(ctx, tx, s.lookupIKeyType).QueryRowContext(ctx, src.ID).Scan(&ikeyType)
			if err != nil {
				return nil, errors.Wrap(err, "lookup integration key type by ID")
			}
			switch ikeyType {
			case integrationkey.TypeGeneric:
				r.subject.classifier = "Generic API"
			case integrationkey.TypeGrafana:
				r.subject.classifier = "Grafana"
			case integrationkey.TypeSite24x7:
				r.subject.classifier = "Site24x7"
			case integrationkey.TypeEmail:
				r.subject.classifier = "Email"
			case integrationkey.TypeUniversal:
				r.subject.classifier = "Universal"
			}
			r.subject.integrationKeyID.Valid = true
			r.subject.integrationKeyID.UUID = uuid.MustParse(src.ID)
		}
	}

	if r.subject.classifier != "" {
		classExtras = append([]string{r.subject.classifier}, classExtras...)
	}
	r.subject.classifier = strings.Join(classExtras, ", ")

	return &r, nil
}

func (s *Store) FindOne(ctx context.Context, logID int) (*Entry, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	var e Entry
	row := s.findOne.QueryRowContext(ctx, logID)
	err = e.scanWith(row.Scan)
	if err != nil {
		return nil, err
	}

	return &e, nil
}

// FindLatestByType returns the latest Log Entry given alertID and status type
func (s *Store) FindLatestByType(ctx context.Context, alertID int, status Type) (*Entry, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	var e Entry
	row := s.findAllByType.QueryRowContext(ctx, alertID, status)
	err = e.scanWith(row.Scan)
	if err != nil {
		return nil, err
	}
	return &e, nil
}
