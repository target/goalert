package message

import (
	"context"
	"fmt"
	"time"

	"github.com/target/goalert/notification"
	"github.com/target/goalert/util/sqlutil"
)

// GlobalCMThrottle represents the rate limits for each notification type.
var GlobalCMThrottle ThrottleConfig = ThrottleRules{{Count: 5, Per: 5 * time.Second}}

// PerCMThrottle configures rate limits for individual contact methods.
var PerCMThrottle ThrottleConfig

func init() {

}

func (db *DB) Init(ctx context.Context) error {

	tx, err := db.lock.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	ids := sqlutil.StringArray{"max_sms_per_15_minutes", "max_sms_per_hour", "max_sms_per_3_hours", "max_voice_per_15_minutes", "max_voice_per_hour", "max_voice_per_3_hours", "max_all_for_alert_status_per_3_minutes", "max_all_for_alert_status_per_2_hours", "max_all_for_alert_status_per_20_minutes"}

	rows, err := tx.StmtContext(ctx, db.getAllLimits).QueryContext(ctx, ids)
	if err != nil {
		return fmt.Errorf("set timeout: %w", err)
	}
	type MaxLimits struct {
		MaxHours int
	}
	var a []MaxLimits
	for rows.Next() {
		var al MaxLimits
		err = rows.Scan(&al.MaxHours)
		if err != nil {
			fmt.Println("Error fetching limits", err)
		}
		a = append(a, al)
	}

	var perCM ThrottleConfigBuilder

	// Rate limit sms, voice and email types
	perCM.
		WithDestTypes(notification.DestTypeVoice, notification.DestTypeSMS, notification.DestTypeUserEmail).
		AddRules([]ThrottleRule{{Count: 1, Per: time.Minute}})

	// On-Call Status Notifications
	perCM.
		WithMsgTypes(notification.MessageTypeScheduleOnCallUsers).
		AddRules([]ThrottleRule{
			{Count: 2, Per: 15 * time.Minute},
			{Count: 4, Per: 1 * time.Hour, Smooth: true},
		})

	// status notifications
	perCM.
		WithMsgTypes(notification.MessageTypeAlertStatus).
		WithDestTypes(notification.DestTypeVoice, notification.DestTypeSMS, notification.DestTypeUserEmail).
		AddRules([]ThrottleRule{
			{Count: a[6].MaxHours, Per: 3 * time.Minute},
			{Count: a[8].MaxHours, Per: 20 * time.Minute},
			{Count: a[7].MaxHours, Per: 120 * time.Minute, Smooth: true},
		})
	// alert notifications
	alertMessages := perCM.WithMsgTypes(notification.MessageTypeAlert, notification.MessageTypeAlertBundle)

	alertMessages.
		WithDestTypes(notification.DestTypeVoice).
		AddRules([]ThrottleRule{
			{Count: a[3].MaxHours, Per: 15 * time.Minute},
			{Count: a[4].MaxHours, Per: time.Hour, Smooth: true},
			{Count: a[5].MaxHours, Per: 3 * time.Hour, Smooth: true},
		})

	alertMessages.
		WithDestTypes(notification.DestTypeSMS).
		AddRules([]ThrottleRule{
			{Count: a[0].MaxHours, Per: 15 * time.Minute},
			{Count: a[1].MaxHours, Per: time.Hour, Smooth: true},
			{Count: a[2].MaxHours, Per: 3 * time.Hour, Smooth: true},
		})

	PerCMThrottle = perCM.Config()
	return nil
}
