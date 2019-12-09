package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/brianvoe/gofakeit"
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
)

var timeZones = []string{"America/Chicago", "Europe/Berlin", "UTC"}
var rotationTypes = []rotation.Type{rotation.TypeDaily, rotation.TypeHourly, rotation.TypeWeekly}

type datagenConfig struct {
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
	Favorites          []userFavorite
	Labels             []label.Label

	ids          *uniqGen
	ints         *uniqIntGen
	alertDetails []string
	labelKeyVal  map[string][]string
	labelKeys    []string
}

func (d *datagen) genPhone() string {
	return fmt.Sprintf("+17633%06d", rand.Intn(999999))
}

// NewUser generates a new user.User and adds it to the Users slice.
func (d *datagen) NewUser() {
	u := user.User{
		ID:    gofakeit.UUID(),
		Name:  d.ids.Gen(gofakeit.Name, "user"),
		Role:  permission.RoleUser,
		Email: d.ids.Gen(gofakeit.Email, "user"),
	}
	d.Users = append(d.Users, u)
}

// NewCM will generate a contact method for the given UserID.
func (d *datagen) NewCM(userID string) {
	cm := contactmethod.ContactMethod{
		ID:       gofakeit.UUID(),
		Type:     contactmethod.TypeSMS,
		Name:     d.ids.Gen(gofakeit.FirstName, userID),
		Disabled: true,
		UserID:   userID,
	}
	if gofakeit.Bool() {
		cm.Type = contactmethod.TypeVoice
	}

	cm.Value = d.ids.Gen(d.genPhone, cm.Type.DestType().String())
	d.ContactMethods = append(d.ContactMethods, cm)
}

// NewNR will generate a notification rule for the user/contact method provided.
func (d *datagen) NewNR(userID, cmID string) {
	nr := notificationrule.NotificationRule{
		ID:              gofakeit.UUID(),
		UserID:          userID,
		ContactMethodID: cmID,
		DelayMinutes:    d.ints.Gen(60, cmID),
	}
	d.NotificationRules = append(d.NotificationRules, nr)
}

// NewRotation will generate a rotation.
func (d *datagen) NewRotation() {
	r := rotation.Rotation{
		ID:          gofakeit.UUID(),
		Name:        d.ids.Gen(idName("Rotation")),
		Description: gofakeit.Sentence(rand.Intn(10) + 3),
		Type:        rotationTypes[rand.Intn(len(rotationTypes))],
		Start:       gofakeit.DateRange(time.Now().AddDate(-3, 0, 0), time.Now()).In(time.FixedZone(gofakeit.RandString(timeZones), 0)),
		ShiftLength: rand.Intn(14) + 1,
	}

	d.Rotations = append(d.Rotations, r)
}

// NewRotationParticipant will create a new rotation participant for the given rotation and position.
func (d *datagen) NewRotationParticipant(rotID string, pos int) {
	d.RotationParts = append(d.RotationParts, rotationPart{
		ID:         gofakeit.UUID(),
		RotationID: rotID,
		UserID:     d.Users[rand.Intn(len(d.Users))].ID,
		Pos:        pos,
	})
}

// NewSchedule will generate a new random schedule.
func (d *datagen) NewSchedule() {
	d.Schedules = append(d.Schedules, schedule.Schedule{
		ID:          gofakeit.UUID(),
		Name:        d.ids.Gen(idName("Schedule")),
		Description: gofakeit.Sentence(rand.Intn(10) + 3),
		TimeZone:    time.FixedZone(gofakeit.RandString(timeZones), 0),
	})
}

// NewScheduleRule will generate a random schedule rule for the provided schedule ID.
func (d *datagen) NewScheduleRule(scheduleID string) {
	var filter rule.WeekdayFilter
	for i := range filter {
		filter.SetDay(time.Weekday(i), gofakeit.Bool())
	}
	var tgt assignment.Target
	if gofakeit.Bool() {
		tgt = assignment.RotationTarget(d.Rotations[rand.Intn(len(d.Rotations))].ID)
	} else {
		tgt = assignment.UserTarget(d.Users[rand.Intn(len(d.Users))].ID)
	}
	d.ScheduleRules = append(d.ScheduleRules, rule.Rule{
		ID:            gofakeit.UUID(),
		ScheduleID:    scheduleID,
		WeekdayFilter: filter,
		Start:         rule.Clock(rand.Int63n(int64(24 * time.Hour))),
		End:           rule.Clock(rand.Int63n(int64(24 * time.Hour))),
		Target:        tgt,
	})
}

// NewScheduleOverride well generate a random override for the provided schedule ID.
func (d *datagen) NewScheduleOverride(scheduleID string) {
	end := gofakeit.DateRange(time.Now().Add(time.Hour), time.Now().Add(30*24*time.Hour))
	start := gofakeit.DateRange(end.Add(-30*24*time.Hour), end.Add(-time.Hour))
	o := override.UserOverride{
		ID:     gofakeit.UUID(),
		Target: assignment.ScheduleTarget(scheduleID),
		Start:  start,
		End:    end,
	}
	n := rand.Intn(3)
	if n < 2 {
		o.AddUserID = d.ids.Gen(func() string { return d.Users[rand.Intn(len(d.Users))].ID }, scheduleID)
	}
	if n > 0 {
		o.RemoveUserID = d.ids.Gen(func() string { return d.Users[rand.Intn(len(d.Users))].ID }, scheduleID)
	}
	d.Overrides = append(d.Overrides, o)
}

// NewEP will generate a new escalation policy.
func (d *datagen) NewEP() {
	d.EscalationPolicies = append(d.EscalationPolicies, escalation.Policy{
		ID:          gofakeit.UUID(),
		Name:        d.ids.Gen(idName("Policy")),
		Description: gofakeit.Sentence(rand.Intn(10) + 3),
		Repeat:      rand.Intn(5),
	})
}

// NewEPStep will generate a random escalation policy step for the provided policy.
func (d *datagen) NewEPStep(epID string, n int) {
	d.EscalationSteps = append(d.EscalationSteps, escalation.Step{
		ID:           gofakeit.UUID(),
		PolicyID:     epID,
		DelayMinutes: rand.Intn(25) + 5,
		StepNumber:   n,
	})
}

// NewEPStepAction will generate a new action for the provided step ID.
func (d *datagen) NewEPStepAction(stepID string) {
	var tgt assignment.Target
	switch rand.Intn(3) {
	case 0:
		tgt = assignment.UserTarget(d.ids.Gen(func() string { return d.Users[rand.Intn(len(d.Users))].ID }, stepID))
	case 1:
		tgt = assignment.RotationTarget(d.ids.Gen(func() string { return d.Rotations[rand.Intn(len(d.Rotations))].ID }, stepID))
	case 2:
		tgt = assignment.ScheduleTarget(d.ids.Gen(func() string { return d.Schedules[rand.Intn(len(d.Schedules))].ID }, stepID))
	}
	d.EscalationActions = append(d.EscalationActions, stepAction{
		ID:     gofakeit.UUID(),
		StepID: stepID,
		Tgt:    tgt,
	})
}

// NewService will generate a random service.
func (d *datagen) NewService() {
	d.Services = append(d.Services, service.Service{
		ID:                 gofakeit.UUID(),
		Name:               d.ids.Gen(idName("Service")),
		Description:        gofakeit.Sentence(rand.Intn(10) + 3),
		EscalationPolicyID: d.EscalationPolicies[rand.Intn(len(d.EscalationPolicies))].ID,
	})
}

// NewIntKey will generate a random integration key for the given service ID.
func (d *datagen) NewIntKey(svcID string) {
	var typ integrationkey.Type
	switch rand.Intn(4) {
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
		ID:        gofakeit.UUID(),
		Name:      d.ids.Gen(idName("Key")),
		Type:      typ,
		ServiceID: svcID,
	})
}

// NewLabel will generate a random label for the provided service ID.
func (d *datagen) NewLabel(svcID string) {
	key := d.ids.Gen(func() string {
		return gofakeit.RandString(d.labelKeys)
	}, "labelKey", svcID)

	d.Labels = append(d.Labels, label.Label{
		Key:    key,
		Value:  gofakeit.RandString(d.labelKeyVal[key]),
		Target: assignment.ServiceTarget(svcID),
	})
}

// NewMonitor will generate a random heartbreat monitor for the provided service ID.
func (d *datagen) NewMonitor(svcID string) {
	d.Monitors = append(d.Monitors, heartbeat.Monitor{
		ID:        gofakeit.UUID(),
		Name:      d.ids.Gen(idName("Monitor")),
		ServiceID: svcID,
		Timeout:   5*time.Minute + time.Duration(rand.Int63n(int64(60*time.Hour))),
	})
}

// NewAlert will generate an alert with the provided status.
func (d *datagen) NewAlert(status alert.Status) {
	var details string
	if gofakeit.Bool() {
		details = gofakeit.RandString(d.alertDetails)
	}
	var src alert.Source
	switch rand.Intn(5) {
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
		serviceID = d.ids.GenN(200, func() string { return d.Services[rand.Intn(len(d.Services))].ID }, "active-alerts")
	} else {
		// unlimited closed alerts
		serviceID = d.Services[rand.Intn(len(d.Services))].ID
	}
	d.Alerts = append(d.Alerts, alert.Alert{
		Status:    status,
		ServiceID: serviceID,
		Summary:   d.ids.Gen(func() string { return gofakeit.Sentence(rand.Intn(10) + 3) }, serviceID),
		Details:   details,
		Source:    src,
	})
}

// NewFavorite will generate a new favorite for the provided user ID.
func (d *datagen) NewFavorite(userID string) {
	var tgt assignment.Target
	switch rand.Intn(3) {
	case 0:
		tgt = assignment.ServiceTarget(d.ids.Gen(func() string { return d.Services[rand.Intn(len(d.Services))].ID }, "favSvc", userID))
	case 1:
		tgt = assignment.RotationTarget(d.ids.Gen(func() string { return d.Rotations[rand.Intn(len(d.Rotations))].ID }, "favRot", userID))
	case 2:
		tgt = assignment.ScheduleTarget(d.ids.Gen(func() string { return d.Schedules[rand.Intn(len(d.Schedules))].ID }, "favSched", userID))
	}

	d.Favorites = append(d.Favorites, userFavorite{
		UserID: userID,
		Tgt:    tgt,
	})
}

// Generate will produce a full random dataset based on the configuration.
func (cfg datagenConfig) Generate() datagen {

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

	d := datagen{
		ids:         newGen(),
		ints:        newUniqIntGen(),
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
		n := run(rand.Intn(cfg.CMMax), func() { d.NewCM(u.ID) })
		cmMethods := d.ContactMethods[len(d.ContactMethods)-n:]
		if len(cmMethods) == 0 {
			continue
		}
		run(rand.Intn(cfg.NRMax), func() { d.NewNR(u.ID, cmMethods[rand.Intn(len(cmMethods))].ID) })
	}

	run(cfg.RotationCount, d.NewRotation)
	for _, r := range d.Rotations {
		var pos int
		run(rand.Intn(cfg.RotationMaxPart), func() { d.NewRotationParticipant(r.ID, pos); pos++ })
	}

	run(cfg.ScheduleCount, d.NewSchedule)
	for _, sched := range d.Schedules {
		run(rand.Intn(cfg.ScheduleMaxRules), func() { d.NewScheduleRule(sched.ID) })
		run(rand.Intn(cfg.ScheduleMaxOverrides), func() { d.NewScheduleOverride(sched.ID) })
	}

	run(cfg.EPCount, d.NewEP)
	for _, ep := range d.EscalationPolicies {
		var stepNum int
		run(rand.Intn(cfg.EPMaxStep), func() {
			d.NewEPStep(ep.ID, stepNum)
			stepNum++
		})
	}
	for _, step := range d.EscalationSteps {
		run(rand.Intn(cfg.EPMaxAssigned), func() { d.NewEPStepAction(step.ID) })
	}

	d.labelKeys = make([]string, cfg.UniqueLabelKeys)
	for i := range d.labelKeys {
		d.labelKeys[i] = d.ids.Gen(func() string {
			return gofakeit.DomainName() + "/" + gofakeit.HackerNoun()
		}, "labelKey")
	}

	for _, key := range d.labelKeys {
		run(rand.Intn(cfg.LabelValueMax)+1, func() {
			d.labelKeyVal[key] = append(d.labelKeyVal[key], d.ids.Gen(func() string {
				return gofakeit.HackerAdjective() + " " + gofakeit.HackerVerb()
			}, "labelKeyVal", key))
		})
	}

	run(cfg.SvcCount, d.NewService)
	for _, svc := range d.Services {
		run(rand.Intn(cfg.IntegrationKeyMax), func() { d.NewIntKey(svc.ID) })
		run(rand.Intn(cfg.HeartbeatMonitorMax), func() { d.NewMonitor(svc.ID) })
		run(rand.Intn(cfg.SvcLabelMax), func() { d.NewLabel(svc.ID) })
	}

	for _, usr := range d.Users {
		run(rand.Intn(cfg.UserFavMax), func() { d.NewFavorite(usr.ID) })
	}

	d.alertDetails = make([]string, 20)
	for i := range d.alertDetails {
		d.alertDetails[i] = gofakeit.Paragraph(2, 4, 10, "\n\n")
	}

	run(cfg.AlertClosedCount, func() { d.NewAlert(alert.StatusClosed) })
	run(cfg.AlertActiveCount, func() { d.NewAlert(alert.StatusActive) })

	return d
}
