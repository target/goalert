package errutil

import (
	"strings"

	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
)

// MapDBError will map known DB errors (like unique names) to a valiation error
func MapDBError(err error) error {
	if err == nil {
		return nil
	}

	dbErr := sqlutil.MapError(err)
	if dbErr == nil {
		return err
	}

	switch dbErr.Code {
	case "23503": // fkey constraint
		switch dbErr.ConstraintName {
		case "schedule_rules_tgt_rotation_id_fkey":
			return validation.NewFieldError("RotationID", "rotation does not exist")
		case "user_calendar_subscriptions_user_id_fkey":
			return validation.NewFieldError("UserID", "user does not exist")
		case "user_calendar_subscriptions_schedule_id_fkey", "schedule_data_schedule_id_fkey":
			return validation.NewFieldError("ScheduleID", "schedule does not exist")
		case "user_overrides_add_user_id_fkey":
			return validation.NewFieldError("AddUserID", "user does not exist")
		case "user_overrides_remove_user_id_fkey":
			return validation.NewFieldError("RemoveUserID", "user does not exist")
		case "user_overrides_tgt_schedule_id_fkey":
			return validation.NewFieldError("TargetID", "schedule does not exist")
		case "alerts_services_id_fkey":
			return validation.NewFieldError("ServiceID", "service does not exist")
		case "schedule_rules_tgt_user_id_fkey":
			return validation.NewFieldError("TargetID", "user does not exist")
		case "rotation_participants_user_id_fkey":
			return validation.NewFieldError("UserID", "user does not exist")
		case "auth_basic_users_user_id_fkey":
			return validation.NewFieldError("UserID", "user does not exist")
		}
	case "23505": // unique constraint
		if dbErr.ConstraintName == "auth_basic_users_username_key" {
			return validation.NewFieldError("Username", "already in use")
		}
		if strings.HasPrefix(dbErr.ConstraintName, dbErr.TableName+"_name") {
			return validation.NewFieldError("Name", "already in use")
		}
		if dbErr.ConstraintName == "user_contact_methods_type_value_key" {
			return validation.NewFieldError("Value", "contact method already exists for that type and value")
		}
		if dbErr.ConstraintName == "user_notification_rules_contact_method_id_delay_minutes_key" {
			return validation.NewFieldError("DelayMinutes", "notification rule already exists for that delay and contact method")
		}
		if dbErr.ConstraintName == "heartbeat_monitor_name_service_id" {
			return validation.NewFieldError("Name", "heartbeat monitor already exists with that name")
		}
		if dbErr.ConstraintName == "idx_no_alert_duplicates" {
			return validation.NewFieldError("", "duplicate alert already exists")
		}
		if dbErr.ConstraintName == "auth_basic_users_pkey" {
			return validation.NewFieldError("UserID", "already has a basic auth username configured")
		}
	case "23514": // check constraint
		newErr := mapLimitError(dbErr)
		if newErr != nil {
			return newErr
		}
		switch dbErr.ConstraintName {
		case "user_overrides_check2":
			return validation.NewFieldError("AddUserID", "cannot be the same as the user being replaced")
		case "user_override_no_conflict_allowed":
			return validation.NewFieldError("UserID", "cannot override the same user twice at the same time, check existing overrides; "+dbErr.Hint)
		case "alert_status_user_id_match":
			return validation.NewFieldError("AlertStatusCMID", "contact method is for wrong user")
		case "notification_rule_user_id_match":
			return validation.NewFieldError("UserID", "contact method is for wrong user")
		}
	}

	switch dbErr.ConstraintName {
	case "services_escalation_policy_id_fkey":
		if strings.Contains(dbErr.Detail, "is still referenced") {
			return validation.NewFieldError("EscalationPolicyID", "is currently in use")
		}
		if strings.Contains(dbErr.Detail, "is not present") {
			return validation.NewFieldError("EscalationPolicyID", "does not exist")
		}
	}

	return err
}
