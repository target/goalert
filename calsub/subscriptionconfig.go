package calsub

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// SubscriptionConfig is the configuration for a calendar subscription.
type SubscriptionConfig struct {
	ReminderMinutes []int
	FullSchedule    bool
}

var (
	_ = driver.Valuer(SubscriptionConfig{})
	_ = sql.Scanner(&SubscriptionConfig{})
)

func (scfg SubscriptionConfig) Value() (driver.Value, error) {
	data, err := json.Marshal(scfg)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (scfg *SubscriptionConfig) Scan(v interface{}) error {
	switch v := v.(type) {
	case []byte:
		return json.Unmarshal(v, scfg)
	case string:
		return json.Unmarshal([]byte(v), scfg)
	default:
		return fmt.Errorf("unsupported type %T", v)
	}
}
