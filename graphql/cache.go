package graphql

import (
	"context"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/escalation"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/service"
	"github.com/target/goalert/user"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/user/notificationrule"
	"github.com/target/goalert/util"
	"strconv"
)

func cachedConfig(c Config) Config {
	cache := util.NewContextCache()
	c.UserStore = &userCache{Store: c.UserStore, c: cache}
	c.AlertStore = &alertCache{Store: c.AlertStore, c: cache}
	c.ServiceStore = &serviceCache{Store: c.ServiceStore, c: cache}
	c.EscalationStore = &escalationCache{Store: c.EscalationStore, c: cache}
	c.ScheduleStore = &scheduleCache{Store: c.ScheduleStore, c: cache}
	c.RotationStore = &rotationCache{Store: c.RotationStore, c: cache}
	c.NRStore = &nrCache{Store: c.NRStore, c: cache}
	c.CMStore = &cmCache{Store: c.CMStore, c: cache}
	return c
}

type nrCache struct {
	notificationrule.Store
	c util.ContextCache
}

func (n *nrCache) FindOne(ctx context.Context, id string) (*notificationrule.NotificationRule, error) {
	v, err := n.c.LoadOrStore(ctx, "nr_find_one:"+id, func() (interface{}, error) {
		return n.Store.FindOne(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	return v.(*notificationrule.NotificationRule), nil
}
func (n *nrCache) FindAll(ctx context.Context, userID string) ([]notificationrule.NotificationRule, error) {
	v, err := n.c.LoadOrStore(ctx, "nr_find_all:"+userID, func() (interface{}, error) {
		return n.Store.FindAll(ctx, userID)
	})
	if err != nil {
		return nil, err
	}
	return v.([]notificationrule.NotificationRule), nil
}

type cmCache struct {
	contactmethod.Store
	c util.ContextCache
}

func (c *cmCache) FindOne(ctx context.Context, id string) (*contactmethod.ContactMethod, error) {
	v, err := c.c.LoadOrStore(ctx, "cm_find_one:"+id, func() (interface{}, error) {
		return c.Store.FindOne(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	return v.(*contactmethod.ContactMethod), nil
}
func (c *cmCache) FindAll(ctx context.Context, userID string) ([]contactmethod.ContactMethod, error) {
	v, err := c.c.LoadOrStore(ctx, "cm_find_all:"+userID, func() (interface{}, error) {
		return c.Store.FindAll(ctx, userID)
	})
	if err != nil {
		return nil, err
	}
	return v.([]contactmethod.ContactMethod), nil
}

type rotationCache struct {
	rotation.Store
	c util.ContextCache
}

func (r *rotationCache) FindAllRotations(ctx context.Context) ([]rotation.Rotation, error) {
	v, err := r.c.LoadOrStore(ctx, "rotation_find_all", func() (interface{}, error) {
		return r.Store.FindAllRotations(ctx)
	})
	if err != nil {
		return nil, err
	}
	return v.([]rotation.Rotation), nil
}
func (r *rotationCache) FindRotation(ctx context.Context, id string) (*rotation.Rotation, error) {
	v, err := r.c.LoadOrStore(ctx, "rotation_find_one:"+id, func() (interface{}, error) {
		return r.Store.FindRotation(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	return v.(*rotation.Rotation), nil
}
func (r *rotationCache) FindParticipant(ctx context.Context, id string) (*rotation.Participant, error) {
	v, err := r.c.LoadOrStore(ctx, "rotation_find_one_participant:"+id, func() (interface{}, error) {
		return r.Store.FindParticipant(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	return v.(*rotation.Participant), nil
}
func (r *rotationCache) FindAllParticipants(ctx context.Context, rotID string) ([]rotation.Participant, error) {
	v, err := r.c.LoadOrStore(ctx, "rotation_find_all_participants:"+rotID, func() (interface{}, error) {
		return r.Store.FindAllParticipants(ctx, rotID)
	})
	if err != nil {
		return nil, err
	}
	return v.([]rotation.Participant), nil
}

type scheduleCache struct {
	schedule.Store
	c util.ContextCache
}

func (s *scheduleCache) FindOne(ctx context.Context, id string) (*schedule.Schedule, error) {
	v, err := s.c.LoadOrStore(ctx, "schedule_find_one:"+id, func() (interface{}, error) {
		return s.Store.FindOne(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	return v.(*schedule.Schedule), nil
}

type escalationCache struct {
	escalation.Store
	c util.ContextCache
}

func (s *escalationCache) FindOnePolicy(ctx context.Context, id string) (*escalation.Policy, error) {
	v, err := s.c.LoadOrStore(ctx, "escalation_find_one_policy:"+id, func() (interface{}, error) {
		return s.Store.FindOnePolicy(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	return v.(*escalation.Policy), nil
}
func (s *escalationCache) FindOneStep(ctx context.Context, id string) (*escalation.Step, error) {
	v, err := s.c.LoadOrStore(ctx, "escalation_find_one_step:"+id, func() (interface{}, error) {
		return s.Store.FindOneStep(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	return v.(*escalation.Step), nil
}
func (s *escalationCache) FindAllSteps(ctx context.Context, id string) ([]escalation.Step, error) {
	v, err := s.c.LoadOrStore(ctx, "escalation_find_all_steps:"+id, func() (interface{}, error) {
		return s.Store.FindAllSteps(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	return v.([]escalation.Step), nil
}

type serviceCache struct {
	service.Store
	c util.ContextCache
}

func (s *serviceCache) FindOne(ctx context.Context, id string) (*service.Service, error) {
	v, err := s.c.LoadOrStore(ctx, "service_fine_one:"+id, func() (interface{}, error) {
		return s.Store.FindOne(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	return v.(*service.Service), nil
}

type alertCache struct {
	alert.Store
	c util.ContextCache
}

func (a *alertCache) FindOne(ctx context.Context, id int) (*alert.Alert, error) {
	v, err := a.c.LoadOrStore(ctx, "alert_find_one:"+strconv.Itoa(id), func() (interface{}, error) {
		return a.Store.FindOne(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	return v.(*alert.Alert), nil
}

type userCache struct {
	user.Store
	c util.ContextCache
}

func (u *userCache) FindOne(ctx context.Context, id string) (*user.User, error) {
	v, err := u.c.LoadOrStore(ctx, "user_find_one:"+id, func() (interface{}, error) { return u.Store.FindOne(ctx, id) })
	if err != nil {
		return nil, err
	}
	return v.(*user.User), nil
}
