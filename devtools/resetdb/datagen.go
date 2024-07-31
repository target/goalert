package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/escalation"
	"github.com/target/goalert/heartbeat"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/label"
	"github.com/target/goalert/override"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/service"
	"github.com/target/goalert/user"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/user/notificationrule"
	"github.com/target/goalert/util/timeutil"
)

var (
	timeZones     = []string{"America/Chicago", "Europe/Berlin", "UTC"}
	rotationTypes = []rotation.Type{rotation.TypeDaily, rotation.TypeHourly, rotation.TypeWeekly, rotation.TypeMonthly}
)

type AlertLog struct {
	AlertID   int
	Timestamp time.Time
	Event     string
	Message   string
	UserID    string
	Class     string
	Meta      json.RawMessage
}

type AlertMsg struct {
	ID        string
	AlertID   int
	UserID    string
	ServiceID string
	CMID      string
	EPID      string
	Status    string
	CreatedAt time.Time
	SentAt    time.Time
}

type datagenConfig struct {
	Seed int64

	UserCount            int
	CMMax                int
	NRMax                int
	NRCMMax              int
	EPCount              int
	EPMaxStep            int
	EPMaxAssigned        int
	SvcCount             int
	RotationMaxPart      int
	ScheduleCount        int
	AlertClosedCount     int
	AlertActiveCount     int
	RotationCount        int
	IntegrationKeyMax    int
	ScheduleMaxRules     int
	ScheduleMaxOverrides int
	HeartbeatMonitorMax  int
	UserFavMax           int
	SvcLabelMax          int
	UniqueLabelKeys      int
	LabelValueMax        int
	MsgPerAlertMax       int

	AdminID string
}
type rotationPart struct {
	ID         string
	RotationID string
	UserID     string
	Pos        int
}
type stepAction struct {
	ID     string
	StepID string
	Tgt    assignment.Target
}
type userFavorite struct {
	UserID string
	Tgt    assignment.Target
}

type datagen struct {
	Users              []user.User
	ContactMethods     []contactmethod.ContactMethod
	NotificationRules  []notificationrule.NotificationRule
	Rotations          []rotation.Rotation
	RotationParts      []rotationPart
	Schedules          []schedule.Schedule
	ScheduleRules      []rule.Rule
	Overrides          []override.UserOverride
	EscalationPolicies []escalation.Policy
	EscalationSteps    []escalation.Step
	EscalationActions  []stepAction
	Services           []service.Service
	IntKeys            []integrationkey.IntegrationKey
	Monitors           []heartbeat.Monitor
	Alerts             []alert.Alert
	AlertLogs          []AlertLog
	AlertFeedback      []alert.Feedback
	Favorites          []userFavorite
	Labels             []label.Label
	AlertMessages      []AlertMsg

	ids          *uniqGen
	ints         *uniqIntGen
	alertDetails []string
	labelKeyVal  map[string][]string
	labelKeys    []string

	*gofakeit.Faker
}

func (d *datagen) genPhone() string {
	return fmt.Sprintf("+17633%06d", d.Intn(1000000))
}

// NewUser generates a new user.User and adds it to the Users slice.
func (d *datagen) NewUser() {
	u := user.User{
		ID:    d.UUID(),
		Name:  d.ids.Gen(d.Name, "user"),
		Role:  permission.RoleUser,
		Email: d.ids.Gen(d.Email, "user"),
	}
	d.Users = append(d.Users, u)
}

// NewCM will generate a contact method for the given UserID.
func (d *datagen) NewCM(userID string) {
	cm := contactmethod.ContactMethod{
		ID:       uuid.MustParse(d.UUID()),
		Type:     contactmethod.TypeSMS,
		Name:     d.ids.Gen(d.FirstName, userID),
		Disabled: true,
		UserID:   userID,
		Pending:  false,
	}
	if d.Bool() {
		cm.Type = contactmethod.TypeVoice
	}

	cm.Value = d.ids.Gen(d.genPhone, string(cm.Type))
	d.ContactMethods = append(d.ContactMethods, cm)
}

// NewNR will generate a notification rule for the user/contact method provided.
func (d *datagen) NewNR(userID, cmID string) {
	nr := notificationrule.NotificationRule{
		ID:              d.UUID(),
		UserID:          userID,
		ContactMethodID: uuid.MustParse(cmID),
		DelayMinutes:    d.ints.Gen(600, cmID),
	}
	d.NotificationRules = append(d.NotificationRules, nr)
}

// NewRotation will generate a rotation.
func (d *datagen) NewRotation() {
	r := rotation.Rotation{
		ID:          d.UUID(),
		Name:        d.ids.Gen(idName(d.Faker, "Rotation")),
		Description: d.LoremIpsumSentence(d.Intn(10) + 3),
		Type:        rotationTypes[d.Intn(len(rotationTypes))],
		Start:       d.DateRange(time.Now().AddDate(-3, 0, 0), time.Now()).In(time.FixedZone(d.RandomString(timeZones), 0)),
		ShiftLength: d.Intn(14) + 1,
	}

	d.Rotations = append(d.Rotations, r)
}

func (d *datagen) Intn(n int) int { return d.IntRange(0, n-1) }

func (d *datagen) Int63n(n int64) int64 { return int64(d.IntRange(0, int(n))) }

// NewRotationParticipant will create a new rotation participant for the given rotation and position.
func (d *datagen) NewRotationParticipant(rotID string, pos int) {
	d.RotationParts = append(d.RotationParts, rotationPart{
		ID:         d.UUID(),
		RotationID: rotID,
		UserID:     d.Users[d.Intn(len(d.Users))].ID,
		Pos:        pos,
	})
}

// NewSchedule will generate a new random schedule.
func (d *datagen) NewSchedule() {
	d.Schedules = append(d.Schedules, schedule.Schedule{
		ID:          d.UUID(),
		Name:        d.ids.Gen(idName(d.Faker, "Schedule")),
		Description: d.LoremIpsumSentence(d.Intn(10) + 3),
		TimeZone:    time.FixedZone(d.RandomString(timeZones), 0),
	})
}

// NewScheduleRule will generate a random schedule rule for the provided schedule ID.
func (d *datagen) NewScheduleRule(scheduleID string) {
	var filter timeutil.WeekdayFilter
	for i := range filter {
		filter.SetDay(time.Weekday(i), d.Bool())
	}
	var tgt assignment.Target
	if d.Bool() {
		tgt = assignment.RotationTarget(d.Rotations[d.Intn(len(d.Rotations))].ID)
	} else {
		tgt = assignment.UserTarget(d.Users[d.Intn(len(d.Users))].ID)
	}
	d.ScheduleRules = append(d.ScheduleRules, rule.Rule{
		ID:            d.UUID(),
		ScheduleID:    scheduleID,
		WeekdayFilter: filter,
		Start:         timeutil.Clock(d.Int63n(int64(24 * time.Hour))),
		End:           timeutil.Clock(d.Int63n(int64(24 * time.Hour))),
		Target:        tgt,
	})
}

// NewScheduleOverride well generate a random override for the provided schedule ID.
func (d *datagen) NewScheduleOverride(scheduleID string) {
	end := d.DateRange(time.Now().Add(time.Hour), time.Now().Add(30*24*time.Hour))
	start := d.DateRange(end.Add(-30*24*time.Hour), end.Add(-time.Hour))
	o := override.UserOverride{
		ID:     d.UUID(),
		Target: assignment.ScheduleTarget(scheduleID),
		Start:  start,
		End:    end,
	}
	n := d.Intn(3)
	if n < 2 {
		o.AddUserID = d.ids.Gen(func() string { return d.Users[d.Intn(len(d.Users))].ID }, scheduleID)
	}
	if n > 0 {
		o.RemoveUserID = d.ids.Gen(func() string { return d.Users[d.Intn(len(d.Users))].ID }, scheduleID)
	}
	d.Overrides = append(d.Overrides, o)
}

// NewEP will generate a new escalation policy.
func (d *datagen) NewEP() {
	d.EscalationPolicies = append(d.EscalationPolicies, escalation.Policy{
		ID:          d.UUID(),
		Name:        d.ids.Gen(idName(d.Faker, "Policy")),
		Description: d.LoremIpsumSentence(d.Intn(10) + 3),
		Repeat:      d.Intn(5),
	})
}

// NewEPStep will generate a random escalation policy step for the provided policy.
func (d *datagen) NewEPStep(epID string, n int) {
	d.EscalationSteps = append(d.EscalationSteps, escalation.Step{
		ID:           uuid.MustParse(d.UUID()),
		PolicyID:     epID,
		DelayMinutes: d.Intn(25) + 5,
		StepNumber:   n,
	})
}

// NewEPStepAction will generate a new action for the provided step ID.
func (d *datagen) NewEPStepAction(stepID string) {
	var tgt assignment.Target
	switch d.Intn(3) {
	case 0:
		tgt = assignment.UserTarget(d.ids.Gen(func() string { return d.Users[d.Intn(len(d.Users))].ID }, stepID))
	case 1:
		tgt = assignment.RotationTarget(d.ids.Gen(func() string { return d.Rotations[d.Intn(len(d.Rotations))].ID }, stepID))
	case 2:
		tgt = assignment.ScheduleTarget(d.ids.Gen(func() string { return d.Schedules[d.Intn(len(d.Schedules))].ID }, stepID))
	}
	d.EscalationActions = append(d.EscalationActions, stepAction{
		ID:     d.UUID(),
		StepID: stepID,
		Tgt:    tgt,
	})
}

// NewService will generate a random service.
func (d *datagen) NewService() {
	d.Services = append(d.Services, service.Service{
		ID:                 d.UUID(),
		Name:               d.ids.Gen(idName(d.Faker, "Service")),
		Description:        d.LoremIpsumSentence(d.Intn(10) + 3),
		EscalationPolicyID: d.EscalationPolicies[d.Intn(len(d.EscalationPolicies))].ID,
	})
}

// NewIntKey will generate a random integration key for the given service ID.
func (d *datagen) NewIntKey(svcID string) {
	var typ integrationkey.Type
	switch d.Intn(4) {
	case 0:
		typ = integrationkey.TypeEmail
	case 1:
		typ = integrationkey.TypeGeneric
	case 2:
		typ = integrationkey.TypeGrafana
	case 3:
		typ = integrationkey.TypeSite24x7
	}
	d.IntKeys = append(d.IntKeys, integrationkey.IntegrationKey{
		ID:        d.UUID(),
		Name:      d.ids.Gen(idName(d.Faker, "Key")),
		Type:      typ,
		ServiceID: svcID,
	})
}

// NewLabel will generate a random label for the provided service ID.
func (d *datagen) NewLabel(svcID string) {
	key := d.ids.Gen(func() string {
		return d.RandomString(d.labelKeys)
	}, "labelKey", svcID)

	d.Labels = append(d.Labels, label.Label{
		Key:    key,
		Value:  d.RandomString(d.labelKeyVal[key]),
		Target: assignment.ServiceTarget(svcID),
	})
}

// NewMonitor will generate a random heartbreat monitor for the provided service ID.
func (d *datagen) NewMonitor(svcID string) {
	d.Monitors = append(d.Monitors, heartbeat.Monitor{
		ID:        d.UUID(),
		Name:      d.ids.Gen(idName(d.Faker, "Monitor")),
		ServiceID: svcID,
		Timeout:   5*time.Minute + time.Duration(d.Int63n(int64(60*time.Hour))),
	})
}

// NewAlert will generate an alert with the provided status.
func (d *datagen) NewAlert(status alert.Status) {
	var details string
	if d.Bool() {
		details = d.RandomString(d.alertDetails)
	}
	var src alert.Source
	switch d.Intn(5) {
	case 0:
		src = alert.SourceEmail
	case 1:
		src = alert.SourceGeneric
	case 2:
		src = alert.SourceGrafana
	case 3:
		src = alert.SourceManual
	case 4:
		src = alert.SourceSite24x7
	}
	var serviceID string
	if status == alert.StatusTriggered {
		serviceID = d.ids.GenN(200, func() string { return d.Services[d.Intn(len(d.Services))].ID }, "active-alerts")
	} else {
		// unlimited closed alerts
		serviceID = d.Services[d.Intn(len(d.Services))].ID
	}
	d.Alerts = append(d.Alerts, alert.Alert{
		ID:        len(d.Alerts) + 1,
		CreatedAt: d.DateRange(time.Now().Add(-180*24*time.Hour), time.Now().Add(-1*time.Hour)),
		Status:    status,
		ServiceID: serviceID,
		Summary:   d.ids.Gen(func() string { return d.LoremIpsumSentence(d.Intn(10) + 3) }, serviceID),
		Details:   details,
		Source:    src,
	})
}

func (d *datagen) NewAlertMessages(a alert.Alert, max int) {
	getEPID := func(svcID string) string {
		idx := sort.Search(len(d.Services), func(n int) bool {
			return d.Services[n].ID >= svcID
		})
		if idx == len(d.Services) {
			panic("service not found: " + svcID)
		}
		return d.Services[idx].EscalationPolicyID
	}

	for i := 0; i < d.Intn(max); i++ {
		cm := d.ContactMethods[d.Intn(len(d.ContactMethods))]
		ts := d.DateRange(a.CreatedAt, time.Now())
		id := d.UUID()
		d.AlertMessages = append(d.AlertMessages, AlertMsg{
			ID:        id,
			AlertID:   a.ID,
			ServiceID: a.ServiceID,
			Status:    "delivered",
			UserID:    cm.UserID,
			EPID:      getEPID(a.ServiceID),
			CMID:      cm.ID.String(),
			SentAt:    ts,
			CreatedAt: d.DateRange(ts.Add(-time.Minute), ts),
		})
		var meta struct {
			MessageID string
		}
		meta.MessageID = id
		data, err := json.Marshal(meta)
		if err != nil {
			panic(err)
		}
		d.AlertLogs = append(d.AlertLogs, AlertLog{
			AlertID:   a.ID,
			Timestamp: ts,
			Event:     "notification_sent",
			UserID:    cm.UserID,
			Meta:      data,
			Class:     string(cm.Type),
		})
	}
}

func (d *datagen) NewAlertFeedback(a alert.Alert) {
	if d.Bool() {
		// no feedback
		return
	}

	var reasons []string
	if d.Bool() {
		reasons = append(reasons, "False positive")
	}
	if d.Bool() {
		reasons = append(reasons, "Not actionable")
	}
	if d.Bool() {
		reasons = append(reasons, "Poor details")
	}
	if d.Bool() {
		reasons = append(reasons, d.Sentence(3))
	}

	d.AlertFeedback = append(d.AlertFeedback, alert.Feedback{
		ID:          a.ID,
		NoiseReason: strings.Join(reasons, "|"),
	})
}

// NewAlertLog will generate an alert log for the provided alert.
func (d *datagen) NewAlertLogs(a alert.Alert) {
	t := a.CreatedAt
	addEvent := func(event string) {
		t = d.DateRange(t, t.Add(30*time.Minute))
		d.AlertLogs = append(d.AlertLogs, AlertLog{
			AlertID:   a.ID,
			Timestamp: t,
			Event:     event,
		})
	}

	// initial creation and escalation
	addEvent("created")
	addEvent("escalated")

	switch a.Status {
	case alert.StatusTriggered:
		if d.Bool() {
			addEvent("escalated")
		}
	case alert.StatusActive:
		if d.Bool() {
			addEvent("escalated")
		}
		addEvent("acknowledged")
	case alert.StatusClosed:
		if d.Bool() {
			addEvent("escalated")
		}
		if d.Bool() {
			addEvent("acknowledged")
		}
		addEvent("closed")
	}
}

// NewFavorite will generate a new favorite for the provided user ID.
func (d *datagen) NewFavorite(userID string) {
	var tgt assignment.Target
	switch d.Intn(5) {
	case 0:
		tgt = assignment.ServiceTarget(d.ids.Gen(func() string { return d.Services[d.Intn(len(d.Services))].ID }, "favSvc", userID))
	case 1:
		tgt = assignment.RotationTarget(d.ids.Gen(func() string { return d.Rotations[d.Intn(len(d.Rotations))].ID }, "favRot", userID))
	case 2:
		tgt = assignment.ScheduleTarget(d.ids.Gen(func() string { return d.Schedules[d.Intn(len(d.Schedules))].ID }, "favSched", userID))
	case 3:
		tgt = assignment.EscalationPolicyTarget(d.ids.Gen(func() string { return d.EscalationPolicies[d.Intn(len(d.EscalationPolicies))].ID }, "favEP", userID))
	case 4:
		tgt = assignment.UserTarget(d.ids.Gen(func() string { return d.Users[d.Intn(len(d.Users))].ID }, "favUsr", userID))
	}

	d.Favorites = append(d.Favorites, userFavorite{
		UserID: userID,
		Tgt:    tgt,
	})
}

func (cfg *datagenConfig) SetDefaults() {
	setDefault := func(val *int, def int) {
		if *val != 0 {
			return
		}

		*val = def
	}
	setDefault(&cfg.UserCount, UserCount)
	setDefault(&cfg.CMMax, CMMax)
	setDefault(&cfg.NRMax, NRMax)
	setDefault(&cfg.RotationCount, RotationCount)
	setDefault(&cfg.RotationMaxPart, RotationMaxPart)
	setDefault(&cfg.ScheduleCount, ScheduleCount)
	setDefault(&cfg.ScheduleMaxRules, ScheduleMaxRules)
	setDefault(&cfg.ScheduleMaxOverrides, ScheduleMaxOverrides)
	setDefault(&cfg.EPCount, EPCount)
	setDefault(&cfg.EPMaxStep, EPMaxStep)
	setDefault(&cfg.EPMaxAssigned, EPMaxAssigned)
	setDefault(&cfg.SvcCount, SvcCount)
	setDefault(&cfg.IntegrationKeyMax, IntegrationKeyMax)
	setDefault(&cfg.AlertClosedCount, AlertClosedCount)
	setDefault(&cfg.AlertActiveCount, AlertActiveCount)
	setDefault(&cfg.HeartbeatMonitorMax, HeartbeatMonitorMax)
	setDefault(&cfg.UserFavMax, UserFavMax)
	setDefault(&cfg.SvcLabelMax, SvcLabelMax)
	setDefault(&cfg.UniqueLabelKeys, UniqueLabelKeys)
	setDefault(&cfg.LabelValueMax, LabelValueMax)
	setDefault(&cfg.MsgPerAlertMax, MsgPerAlertMax)
}

// Multiply will multiply the following counts:
// - UserCount
// - RotationCount
// - ScheduleCount
// - EPCount
// - SvcCount
// - AlertClosedCount
// - AlertActiveCount
func (cfg *datagenConfig) Multiply(n float64) {
	cfg.UserCount = int(float64(cfg.UserCount) * n)
	cfg.RotationCount = int(float64(cfg.RotationCount) * n)
	cfg.ScheduleCount = int(float64(cfg.ScheduleCount) * n)
	cfg.EPCount = int(float64(cfg.EPCount) * n)
	cfg.SvcCount = int(float64(cfg.SvcCount) * n)
	cfg.AlertClosedCount = int(float64(cfg.AlertClosedCount) * n)
	cfg.AlertActiveCount = int(float64(cfg.AlertActiveCount) * n)
}

// Generate will produce a full random dataset based on the configuration.
func (cfg datagenConfig) Generate() datagen {
	f := gofakeit.New(cfg.Seed)
	d := datagen{
		Faker: f,

		ids:         newGen(f),
		ints:        newUniqIntGen(f),
		labelKeyVal: make(map[string][]string),
		Alerts:      make([]alert.Alert, 0, cfg.AlertClosedCount+cfg.AlertActiveCount),
	}

	run := func(times int, fn func()) int {
		for i := 0; i < times; i++ {
			fn()
		}
		return times
	}

	if cfg.AdminID != "" {
		d.Users = append(d.Users, user.User{
			ID:    cfg.AdminID,
			Name:  "Admin McAdminFace",
			Role:  permission.RoleAdmin,
			Email: "admin@example.com",
		})
	}
	run(cfg.UserCount, d.NewUser)
	for _, u := range d.Users {
		n := run(d.Intn(cfg.CMMax), func() { d.NewCM(u.ID) })
		cmMethods := d.ContactMethods[len(d.ContactMethods)-n:]
		if len(cmMethods) == 0 {
			continue
		}
		run(d.Intn(cfg.NRMax), func() { d.NewNR(u.ID, cmMethods[d.Intn(len(cmMethods))].ID.String()) })
	}

	run(cfg.RotationCount, d.NewRotation)
	for _, r := range d.Rotations {
		var pos int
		run(d.Intn(cfg.RotationMaxPart), func() { d.NewRotationParticipant(r.ID, pos); pos++ })
	}

	run(cfg.ScheduleCount, d.NewSchedule)
	for _, sched := range d.Schedules {
		run(d.Intn(cfg.ScheduleMaxRules), func() { d.NewScheduleRule(sched.ID) })
		run(d.Intn(cfg.ScheduleMaxOverrides), func() { d.NewScheduleOverride(sched.ID) })
	}

	run(cfg.EPCount, d.NewEP)
	for _, ep := range d.EscalationPolicies {
		var stepNum int
		run(d.Intn(cfg.EPMaxStep), func() {
			d.NewEPStep(ep.ID, stepNum)
			stepNum++
		})
	}
	for _, step := range d.EscalationSteps {
		run(d.Intn(cfg.EPMaxAssigned), func() { d.NewEPStepAction(step.ID.String()) })
	}

	d.labelKeys = make([]string, cfg.UniqueLabelKeys)
	for i := range d.labelKeys {
		d.labelKeys[i] = d.ids.Gen(func() string {
			return d.DomainName() + "/" + d.HackerNoun()
		}, "labelKey")
	}

	for _, key := range d.labelKeys {
		run(d.Intn(cfg.LabelValueMax)+1, func() {
			d.labelKeyVal[key] = append(d.labelKeyVal[key], d.ids.Gen(func() string {
				return d.HackerAdjective() + " " + d.HackerVerb()
			}, "labelKeyVal", key))
		})
	}

	run(cfg.SvcCount, d.NewService)
	for _, svc := range d.Services {
		run(d.Intn(cfg.IntegrationKeyMax), func() { d.NewIntKey(svc.ID) })
		run(d.Intn(cfg.HeartbeatMonitorMax), func() { d.NewMonitor(svc.ID) })
		run(d.Intn(cfg.SvcLabelMax), func() { d.NewLabel(svc.ID) })
	}
	sort.Slice(d.Services, func(i, j int) bool { return d.Services[i].ID < d.Services[j].ID })

	for _, usr := range d.Users {
		run(d.Intn(cfg.UserFavMax), func() { d.NewFavorite(usr.ID) })
	}

	d.alertDetails = make([]string, 20)
	for i := range d.alertDetails {
		d.alertDetails[i] = d.LoremIpsumParagraph(2, 4, 10, "\n\n")
	}

	run(cfg.AlertClosedCount, func() { d.NewAlert(alert.StatusClosed) })
	run(cfg.AlertActiveCount, func() { d.NewAlert(alert.StatusActive) })

	for _, alert := range d.Alerts {
		d.NewAlertLogs(alert)
		d.NewAlertMessages(alert, cfg.MsgPerAlertMax)
		d.NewAlertFeedback(alert)
	}

	return d
}
