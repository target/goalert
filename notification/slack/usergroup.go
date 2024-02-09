package slack

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/validation"
)

type UserGroup struct {
	ID     string
	TeamID string
	Name   string
	Handle string
}

func (s *ChannelSender) ValidateUserGroup(ctx context.Context, fieldID, id string) error {
	ug, err := s._UserGroup(ctx, id)
	if err != nil {
		return err
	}

	if ug == nil {
		return validation.NewFieldError(fieldID, "user group not found")
	}

	return nil
}

// User will lookup a single Slack user group.
func (s *ChannelSender) UserGroup(ctx context.Context, id string) (*UserGroup, error) {
	ug, err := s._UserGroup(ctx, id)
	if err != nil {
		return nil, err
	}
	if ug == nil {
		return nil, validation.NewGenericError("invalid user group id")
	}

	return ug, nil
}

// User will lookup a single Slack user group.
func (s *ChannelSender) _UserGroup(ctx context.Context, id string) (*UserGroup, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.System)
	if err != nil {
		return nil, err
	}

	ug, ok := s.ugInfoCache.Get(id)
	if ok {
		return &ug, nil
	}

	groups, err := s.ListUserGroups(ctx)
	if err != nil {
		return nil, err
	}

	for _, g := range groups {
		if g.ID == id {
			s.ugInfoCache.Add(id, g)
			return &g, nil
		}
	}

	return nil, validation.NewGenericError("invalid user group id")
}

// ListUserGroups will return a list of all Slack user groups.
func (s *ChannelSender) ListUserGroups(ctx context.Context) ([]UserGroup, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.System)
	if err != nil {
		return nil, err
	}

	id, err := s.TeamID(ctx)
	if err != nil {
		return nil, err
	}

	groups, ok := s.ugCache.Get(id)
	if !ok {
		err = s.withClient(ctx, func(c *slack.Client) error {
			groups, err = c.GetUserGroupsContext(ctx)
			return err
		})
		if err != nil {
			return nil, fmt.Errorf("get user groups: %w", err)
		}
		s.ugCache.Add(id, groups)
	}

	res := make([]UserGroup, 0, len(groups))
	for _, g := range groups {
		grp := UserGroup{
			ID:     g.ID,
			Name:   g.Name,
			Handle: "@" + g.Handle,
			TeamID: g.TeamID,
		}
		res = append(res, grp)
		s.ugInfoCache.Add(g.ID, grp)
	}

	return res, nil
}
