package graphqlapp

import (
	context "context"
	"database/sql"
	"strconv"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/escalation"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/heartbeat"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/label"
	"github.com/target/goalert/notice"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/service"
	"github.com/target/goalert/service/rule"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

const tempUUID = "00000000-0000-0000-0000-000000000001"

type Service App

func (a *App) Service() graphql2.ServiceResolver { return (*Service)(a) }

func (q *Query) Service(ctx context.Context, id string) (*service.Service, error) {
	return (*App)(q).FindOneService(ctx, id)
}

func (q *Query) Services(ctx context.Context, opts *graphql2.ServiceSearchOptions) (conn *graphql2.ServiceConnection, err error) {
	if opts == nil {
		opts = &graphql2.ServiceSearchOptions{}
	}

	var searchOpts service.SearchOptions
	searchOpts.FavoritesUserID = permission.UserID(ctx)
	if opts.Search != nil {
		searchOpts.Search = *opts.Search
	}
	if opts.FavoritesOnly != nil {
		searchOpts.FavoritesOnly = *opts.FavoritesOnly
	}
	if opts.FavoritesFirst != nil {
		searchOpts.FavoritesFirst = *opts.FavoritesFirst
	}
	searchOpts.Omit = opts.Omit
	if opts.After != nil && *opts.After != "" {
		err = search.ParseCursor(*opts.After, &searchOpts)
		if err != nil {
			return nil, err
		}
	}
	if opts.First != nil {
		searchOpts.Limit = *opts.First
	}
	if searchOpts.Limit == 0 {
		searchOpts.Limit = 15
	}

	searchOpts.Limit++
	svcs, err := q.ServiceStore.Search(ctx, &searchOpts)
	if err != nil {
		return nil, err
	}
	conn = new(graphql2.ServiceConnection)
	conn.PageInfo = &graphql2.PageInfo{}
	if len(svcs) == searchOpts.Limit {
		svcs = svcs[:len(svcs)-1]
		conn.PageInfo.HasNextPage = true
	}
	if len(svcs) > 0 {
		last := svcs[len(svcs)-1]
		searchOpts.After.IsFavorite = last.IsUserFavorite()
		searchOpts.After.Name = last.Name

		cur, err := search.Cursor(searchOpts)
		if err != nil {
			return conn, err
		}
		conn.PageInfo.EndCursor = &cur
	}
	conn.Nodes = svcs
	return conn, err
}

func (s *Service) Notices(ctx context.Context, raw *service.Service) ([]notice.Notice, error) {
	return s.NoticeStore.FindAllServiceNotices(ctx, raw.ID)
}

func (s *Service) Labels(ctx context.Context, raw *service.Service) ([]label.Label, error) {
	return s.LabelStore.FindAllByService(ctx, raw.ID)
}

func (s *Service) EscalationPolicy(ctx context.Context, raw *service.Service) (*escalation.Policy, error) {
	return (*App)(s).FindOnePolicy(ctx, raw.EscalationPolicyID)
}

func (s *Service) IsFavorite(ctx context.Context, raw *service.Service) (bool, error) {
	return raw.IsUserFavorite(), nil
}

func (s *Service) OnCallUsers(ctx context.Context, raw *service.Service) ([]oncall.ServiceOnCallUser, error) {
	return s.OnCallStore.OnCallUsersByService(ctx, raw.ID)
}

func (s *Service) IntegrationKeys(ctx context.Context, raw *service.Service) ([]integrationkey.IntegrationKey, error) {
	return s.IntKeyStore.FindAllByService(ctx, raw.ID)
}

func (s *Service) HeartbeatMonitors(ctx context.Context, raw *service.Service) ([]heartbeat.Monitor, error) {
	return s.HeartbeatStore.FindAllByService(ctx, raw.ID)
}

func (s *Service) Rules(ctx context.Context, raw *service.Service) ([]rule.Rule, error) {
	return s.ServiceRuleStore.FindAllByService(ctx, raw.ID)
}

func (m *Mutation) CreateService(ctx context.Context, input graphql2.CreateServiceInput) (result *service.Service, err error) {
	if input.NewEscalationPolicy != nil && input.EscalationPolicyID != nil && *input.EscalationPolicyID != "" {
		return nil, validation.NewFieldError("newEscalationPolicy", "cannot be used with `escalationPolicyID`.")
	}

	err = withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		svc := &service.Service{
			Name: input.Name,
		}
		if input.EscalationPolicyID != nil {
			svc.EscalationPolicyID = *input.EscalationPolicyID
		}
		if input.Description != nil {
			svc.Description = *input.Description
		}
		if input.NewEscalationPolicy != nil {
			// Set tempUUID so that Normalize won't fail on the yet-to-be-created
			// escalation policy.
			//
			// We want to fail on service validation errors before attempting to
			// create the nested policy.
			svc.EscalationPolicyID = tempUUID
		}
		_, err := svc.Normalize()
		if err != nil {
			return err
		}

		if input.NewEscalationPolicy != nil {
			ep, err := m.CreateEscalationPolicy(ctx, *input.NewEscalationPolicy)
			if err != nil {
				return validation.AddPrefix("newEscalationPolicy.", err)
			}
			svc.EscalationPolicyID = ep.ID
		}

		result, err = m.ServiceStore.CreateServiceTx(ctx, tx, svc)
		if err != nil {
			return err
		}

		if input.Favorite != nil && *input.Favorite {
			err = m.FavoriteStore.SetTx(ctx, tx, permission.UserID(ctx), assignment.ServiceTarget(result.ID))
			if err != nil {
				return err
			}
		}

		err = validate.Many(
			validate.Range("NewIntegrationKeys", len(input.NewIntegrationKeys), 0, 5),
			validate.Range("Labels", len(input.Labels), 0, 5),
		)
		if err != nil {
			return err
		}

		for i, key := range input.NewIntegrationKeys {
			key.ServiceID = &result.ID
			_, err = m.CreateIntegrationKey(ctx, key)
			if err != nil {
				return validation.AddPrefix("newIntegrationKeys["+strconv.Itoa(i)+"].", err)
			}
		}

		for i, hb := range input.NewHeartbeatMonitors {
			hb.ServiceID = &result.ID
			_, err = m.CreateHeartbeatMonitor(ctx, hb)
			if err != nil {
				return validation.AddPrefix("newHeartbeatMonitors["+strconv.Itoa(i)+"].", err)
			}
		}

		for i, lbl := range input.Labels {
			lbl.Target = &assignment.RawTarget{Type: assignment.TargetTypeService, ID: result.ID}
			_, err = m.SetLabel(ctx, lbl)
			if err != nil {
				return validation.AddPrefix("labels["+strconv.Itoa(i)+"].", err)
			}
		}

		return err
	})

	return result, err
}

func (a *Mutation) UpdateService(ctx context.Context, input graphql2.UpdateServiceInput) (bool, error) {
	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer sqlutil.Rollback(ctx, "graphql: update service", tx)

	svc, err := a.ServiceStore.FindOneForUpdate(ctx, tx, input.ID)
	if err != nil {
		return false, err
	}

	if input.Name != nil {
		svc.Name = *input.Name
	}
	if input.Description != nil {
		svc.Description = *input.Description
	}
	if input.EscalationPolicyID != nil {
		svc.EscalationPolicyID = *input.EscalationPolicyID
	}

	if input.MaintenanceExpiresAt != nil {
		svc.MaintenanceExpiresAt = *input.MaintenanceExpiresAt
	}

	err = a.ServiceStore.UpdateTx(ctx, tx, svc)
	if err != nil {
		return false, err
	}

	err = tx.Commit()
	if err != nil {
		return false, err
	}

	return true, nil
}
