package alertlog

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
)

type Store struct {
	db *sql.DB

	insert        *sql.Stmt
	insertEP      *sql.Stmt
	insertSvc     *sql.Stmt
	findAll       *sql.Stmt
	findAllByType *sql.Stmt
	findOne       *sql.Stmt

	lookupCallbackType *sql.Stmt
	lookupIKeyType     *sql.Stmt
	lookupCMType       *sql.Stmt
	lookupNCTypeName   *sql.Stmt
	lookupHBInterval   *sql.Stmt
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
		lookupCMType: p.P(`
			select "type" from user_contact_methods where id = $1
		`),
		lookupNCTypeName: p.P(`
			select "type", name from notification_channels where id = $1
		`),
		lookupHBInterval: p.P(`
			select extract(epoch from heartbeat_interval)/60 from heartbeat_monitors where id = $1
		`),
		lookupIKeyType: p.P(`select "type" from integration_keys where id = $1`),
		insertEP: p.P(`
			insert into alert_logs (
				alert_id,
				event,
				sub_type,
				sub_user_id,
				sub_integration_key_id,
				sub_hb_monitor_id,
				sub_channel_id,
				sub_classifier,
				meta,
				message	
			)
			select
				a.id, $2, $3, $4, $5, $6, $7, $8, $9, $10
			from alerts a
			join services svc on svc.id = a.service_id and svc.escalation_policy_id = ANY ($1)
			where a.status != 'closed'
		`),
		insertSvc: p.P(`
			insert into alert_logs (
				alert_id,
				event,
				sub_type,
				sub_user_id,
				sub_integration_key_id,
				sub_hb_monitor_id,
				sub_channel_id,
				sub_classifier,
				meta,
				message	
			)
			select
				a.id, $2, $3, $4, $5, $6, $7, $8, $9, $10
			from alerts a
			where a.service_id = ANY ($1) and (
				($2 = 'closed'::enum_alert_log_event and a.status != 'closed') or
				($2::enum_alert_log_event in ('acknowledged', 'notification_sent') and a.status = 'triggered')
			)
		`),
		insert: p.P(`
			insert into alert_logs (
				alert_id,
				event,
				sub_type,
				sub_user_id,
				sub_integration_key_id,
				sub_hb_monitor_id,
				sub_channel_id,
				sub_classifier,
				meta,
				message
			)
			SELECT unnest, $2, $3, $4, $5, $6, $7, $8, $9, $10
			FROM unnest($1::int[])
		`),
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
	err := validate.UUID("EscalationPolicyID", epID)
	if err != nil {
		return err
	}
	return s.logAny(ctx, tx, s.insertEP, epID, _type, meta)
}

func (s *Store) LogServiceTx(ctx context.Context, tx *sql.Tx, serviceID string, _type Type, meta interface{}) error {
	err := validate.UUID("ServiceID", serviceID)
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
	return s.logAny(ctx, tx, s.insertSvc, serviceID, t, meta)
}

func (s *Store) LogManyTx(ctx context.Context, tx *sql.Tx, alertIDs []int, _type Type, meta interface{}) error {
	return s.logAny(ctx, tx, s.insert, alertIDs, _type, meta)
}

func (s *Store) LogTx(ctx context.Context, tx *sql.Tx, alertID int, _type Type, meta interface{}) error {
	return s.logAny(ctx, tx, s.insert, alertID, _type, meta)
}

func txWrap(ctx context.Context, tx *sql.Tx, stmt *sql.Stmt) *sql.Stmt {
	if tx == nil {
		return stmt
	}
	return tx.StmtContext(ctx, stmt)
}

func (s *Store) logAny(ctx context.Context, tx *sql.Tx, insertStmt *sql.Stmt, id interface{}, _type Type, meta interface{}) error {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return err
	}

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

	if meta != nil {
		r.meta, err = json.Marshal(meta)
		if err != nil {
			return err
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
				return errors.Wrap(err, "lookup contact method type for callback ID")
			}

			switch ncType {
			case notificationchannel.TypeSlackChan:
				r.subject.classifier = "Slack"
			case notificationchannel.TypeWebhook:
				r.subject.classifier = "Webhook"
			}
			r.subject.channelID.String = src.ID
			r.subject.channelID.Valid = true
		case permission.SourceTypeAuthProvider:
			r.subject.classifier = "Web"
			r.subject._type = SubjectTypeUser

			r.subject.userID.String = permission.UserID(ctx)
			if r.subject.userID.String != "" {
				r.subject.userID.Valid = true
			}
		case permission.SourceTypeContactMethod:
			r.subject._type = SubjectTypeUser
			r.subject.userID.String = permission.UserID(ctx)
			if r.subject.userID.String != "" {
				r.subject.userID.Valid = true
			}
			if _type == TypeNoNotificationSent {
				// no CMID for no notification sent
				r.subject.classifier = "no immediate rule"
				break
			}
			var cmType contactmethod.Type
			err = txWrap(ctx, tx, s.lookupCMType).QueryRowContext(ctx, src.ID).Scan(&cmType)
			if err != nil {
				return errors.Wrap(err, "lookup contact method type for callback ID")
			}
			switch cmType {
			case contactmethod.TypeVoice:
				r.subject.classifier = "Voice"
			case contactmethod.TypeSMS:
				r.subject.classifier = "SMS"
			case contactmethod.TypeEmail:
				r.subject.classifier = "Email"
			case contactmethod.TypeWebhook:
				r.subject.classifier = "Webhook"
			case contactmethod.TypeSlackDM:
				r.subject.classifier = "Slack"
			}

		case permission.SourceTypeNotificationCallback:
			r.subject._type = SubjectTypeUser
			var dt notification.ScannableDestType
			err = txWrap(ctx, tx, s.lookupCallbackType).QueryRowContext(ctx, src.ID).Scan(&dt.CM, &dt.NC)
			if err != nil {
				return errors.Wrap(err, "lookup notification type for callback ID")
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
			r.subject.userID.String = permission.UserID(ctx)
			if r.subject.userID.String != "" {
				r.subject.userID.Valid = true
			}

		case permission.SourceTypeHeartbeat:
			r.subject._type = SubjectTypeHeartbeatMonitor
			var minutes int
			err = txWrap(ctx, tx, s.lookupHBInterval).QueryRowContext(ctx, src.ID).Scan(&minutes)
			if err != nil {
				return errors.Wrap(err, "lookup heartbeat monitor interval by ID")
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
			r.subject.heartbeatMonitorID.String = src.ID
		case permission.SourceTypeIntegrationKey:
			r.subject._type = SubjectTypeIntegrationKey
			var ikeyType integrationkey.Type
			err = txWrap(ctx, tx, s.lookupIKeyType).QueryRowContext(ctx, src.ID).Scan(&ikeyType)
			if err != nil {
				return errors.Wrap(err, "lookup integration key type by ID")
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
			}
			r.subject.integrationKeyID.Valid = true
			r.subject.integrationKeyID.String = src.ID
		}
	}

	if r.subject.classifier != "" {
		classExtras = append([]string{r.subject.classifier}, classExtras...)
	}
	r.subject.classifier = strings.Join(classExtras, ", ")

	var idArg interface{}

	switch t := id.(type) {
	case string:
		idArg = sqlutil.UUIDArray{t}
	case int:
		idArg = sqlutil.IntArray{t}
	case []int:
		idArg = sqlutil.IntArray(t)
	default:
		return errors.Errorf("invalid id type %T", t)
	}

	_, err = txWrap(ctx, tx, insertStmt).ExecContext(ctx, idArg, _type, r.subject._type, r.subject.userID, r.subject.integrationKeyID, r.subject.heartbeatMonitorID, r.subject.channelID, r.subject.classifier, r.meta, r.String(ctx))

	return err
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
