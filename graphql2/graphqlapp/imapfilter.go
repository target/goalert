package graphqlapp

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/service"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

// Query resolvers

func (q *Query) ImapFilterRules(ctx context.Context, serviceID string) ([]graphql2.IMAPFilterRule, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	svcID, err := validate.ParseUUID("ServiceID", serviceID)
	if err != nil {
		return nil, err
	}

	rules, err := gadb.New(q.DB).IMAPFilterRulesAll(ctx, svcID)
	if err != nil {
		return nil, err
	}

	result := make([]graphql2.IMAPFilterRule, len(rules))
	for i, r := range rules {
		result[i] = graphql2.IMAPFilterRule{
			ID:             r.ID.String(),
			ServiceID:      r.ServiceID.String(),
			Name:           r.Name,
			Enabled:        r.Enabled,
			FromPattern:    ptrIfValid(r.FromPattern),
			SubjectPattern: ptrIfValid(r.SubjectPattern),
			ToPattern:      ptrIfValid(r.ToPattern),
			MatchMode:      graphql2.IMAPFilterMatchMode(r.MatchMode),
			ExcludeReplies: r.ExcludeReplies,
			CreatedAt:      r.CreatedAt,
			UpdatedAt:      r.UpdatedAt,
		}
	}

	return result, nil
}

func ptrIfValid(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

// Mutation resolvers

func (m *Mutation) CreateIMAPFilterRule(ctx context.Context, input graphql2.CreateIMAPFilterRuleInput) (*graphql2.IMAPFilterRule, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	svcID, err := validate.ParseUUID("ServiceID", input.ServiceID)
	if err != nil {
		return nil, err
	}

	// Validate that at least one pattern is provided
	if input.FromPattern == nil && input.SubjectPattern == nil && input.ToPattern == nil {
		return nil, validation.NewFieldError("Input", "at least one pattern (fromPattern, subjectPattern, or toPattern) must be provided")
	}

	// Validate match mode
	var matchMode string
	switch input.MatchMode {
	case graphql2.IMAPFilterMatchModeExact:
		matchMode = "exact"
	case graphql2.IMAPFilterMatchModeContains:
		matchMode = "contains"
	case graphql2.IMAPFilterMatchModeRegex:
		matchMode = "regex"
	default:
		return nil, validation.NewFieldError("MatchMode", "invalid match mode")
	}

	var fromPattern, subjectPattern, toPattern sql.NullString
	if input.FromPattern != nil {
		fromPattern = sql.NullString{String: *input.FromPattern, Valid: true}
	}
	if input.SubjectPattern != nil {
		subjectPattern = sql.NullString{String: *input.SubjectPattern, Valid: true}
	}
	if input.ToPattern != nil {
		toPattern = sql.NullString{String: *input.ToPattern, Valid: true}
	}

	rule, err := gadb.New(m.DB).IMAPFilterRuleCreate(ctx, gadb.IMAPFilterRuleCreateParams{
		ServiceID:      svcID,
		Name:           input.Name,
		Enabled:        true, // New rules start enabled
		FromPattern:    fromPattern,
		SubjectPattern: subjectPattern,
		ToPattern:      toPattern,
		MatchMode:      matchMode,
		ExcludeReplies: input.ExcludeReplies,
	})
	if err != nil {
		return nil, err
	}

	return &graphql2.IMAPFilterRule{
		ID:             rule.ID.String(),
		ServiceID:      rule.ServiceID.String(),
		Name:           rule.Name,
		Enabled:        rule.Enabled,
		FromPattern:    ptrIfValid(rule.FromPattern),
		SubjectPattern: ptrIfValid(rule.SubjectPattern),
		ToPattern:      ptrIfValid(rule.ToPattern),
		MatchMode:      graphql2.IMAPFilterMatchMode(rule.MatchMode),
		ExcludeReplies: rule.ExcludeReplies,
		CreatedAt:      rule.CreatedAt,
		UpdatedAt:      rule.UpdatedAt,
	}, nil
}

func (m *Mutation) UpdateIMAPFilterRule(ctx context.Context, input graphql2.UpdateIMAPFilterRuleInput) (bool, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.User)
	if err != nil {
		return false, err
	}

	ruleID, err := validate.ParseUUID("ID", input.ID)
	if err != nil {
		return false, err
	}

	// Build update parameters
	var name, matchMode sql.NullString
	var enabled sql.NullBool
	var fromPattern, subjectPattern, toPattern sql.NullString
	var excludeReplies sql.NullBool

	if input.Name != nil {
		name = sql.NullString{String: *input.Name, Valid: true}
	}
	if input.Enabled != nil {
		enabled = sql.NullBool{Bool: *input.Enabled, Valid: true}
	}
	if input.FromPattern != nil {
		if *input.FromPattern == "" {
			// Empty string means clear the field (set to NULL)
			fromPattern = sql.NullString{Valid: true}
		} else {
			fromPattern = sql.NullString{String: *input.FromPattern, Valid: true}
		}
	}
	if input.SubjectPattern != nil {
		if *input.SubjectPattern == "" {
			// Empty string means clear the field (set to NULL)
			subjectPattern = sql.NullString{Valid: true}
		} else {
			subjectPattern = sql.NullString{String: *input.SubjectPattern, Valid: true}
		}
	}
	if input.ToPattern != nil {
		if *input.ToPattern == "" {
			// Empty string means clear the field (set to NULL)
			toPattern = sql.NullString{Valid: true}
		} else {
			toPattern = sql.NullString{String: *input.ToPattern, Valid: true}
		}
	}
	if input.MatchMode != nil {
		switch *input.MatchMode {
		case graphql2.IMAPFilterMatchModeExact:
			matchMode = sql.NullString{String: "exact", Valid: true}
		case graphql2.IMAPFilterMatchModeContains:
			matchMode = sql.NullString{String: "contains", Valid: true}
		case graphql2.IMAPFilterMatchModeRegex:
			matchMode = sql.NullString{String: "regex", Valid: true}
		default:
			return false, validation.NewFieldError("MatchMode", "invalid match mode")
		}
	}
	if input.ExcludeReplies != nil {
		excludeReplies = sql.NullBool{Bool: *input.ExcludeReplies, Valid: true}
	}

	err = gadb.New(m.DB).IMAPFilterRuleUpdate(ctx, gadb.IMAPFilterRuleUpdateParams{
		ID:             ruleID,
		Name:           name,
		Enabled:        enabled,
		FromPattern:    fromPattern,
		SubjectPattern: subjectPattern,
		ToPattern:      toPattern,
		MatchMode:      matchMode,
		ExcludeReplies: excludeReplies,
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

func (m *Mutation) DeleteIMAPFilterRule(ctx context.Context, id string) (bool, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.User)
	if err != nil {
		return false, err
	}

	ruleID, err := validate.ParseUUID("ID", id)
	if err != nil {
		return false, err
	}

	err = gadb.New(m.DB).IMAPFilterRuleDelete(ctx, ruleID)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ServiceIMAPConfig mutations

func (m *Mutation) CreateServiceIMAPConfig(ctx context.Context, input graphql2.CreateServiceIMAPConfigInput) (*graphql2.ServiceIMAPConfig, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	svcID, err := validate.ParseUUID("ServiceID", input.ServiceID)
	if err != nil {
		return nil, err
	}

	// Convert input values including OAuth credentials
	var oauthClientID, oauthClientSecret, oauthRefreshToken sql.NullString
	if input.OauthClientID != nil && *input.OauthClientID != "" {
		oauthClientID = sql.NullString{String: *input.OauthClientID, Valid: true}
	}
	if input.OauthClientSecret != nil && *input.OauthClientSecret != "" {
		oauthClientSecret = sql.NullString{String: *input.OauthClientSecret, Valid: true}
	}
	if input.OauthRefreshToken != nil && *input.OauthRefreshToken != "" {
		oauthRefreshToken = sql.NullString{String: *input.OauthRefreshToken, Valid: true}
	}

	config, err := gadb.New(m.DB).IMAPConfigCreate(ctx, gadb.IMAPConfigCreateParams{
		ServiceID:            svcID,
		Enabled:              input.Enabled,
		OauthClientID:        oauthClientID,
		OauthClientSecret:    oauthClientSecret,
		OauthRefreshToken:    oauthRefreshToken,
		Host:                 input.Host,
		Port:                 int32(input.Port),
		Username:             input.Username,
		UseTls:               input.UseTLS,
		Mailbox:              input.Mailbox,
		PollIntervalMinutes:  int32(input.PollIntervalMinutes),
		MarkAsRead:           input.MarkAsRead,
		DeleteAfter:          input.DeleteAfter,
		IncludeHeaders:       input.IncludeHeaders,
		IncludeFrom:          input.IncludeFrom,
		IncludeTo:            input.IncludeTo,
		IncludeSubject:       input.IncludeSubject,
		IncludeBody:          input.IncludeBody,
	})
	if err != nil {
		return nil, err
	}

	return &graphql2.ServiceIMAPConfig{
		ServiceID:           config.ServiceID.String(),
		Enabled:             config.Enabled,
		Host:                config.Host,
		Port:                int(config.Port),
		Username:            config.Username,
		UseTLS:              config.UseTls,
		Mailbox:             config.Mailbox,
		PollIntervalMinutes: int(config.PollIntervalMinutes),
		MarkAsRead:          config.MarkAsRead,
		DeleteAfter:         config.DeleteAfter,
		IncludeHeaders:      config.IncludeHeaders,
		IncludeFrom:         config.IncludeFrom,
		IncludeTo:           config.IncludeTo,
		IncludeSubject:      config.IncludeSubject,
		IncludeBody:         config.IncludeBody,
		CreatedAt:           config.CreatedAt,
		UpdatedAt:           config.UpdatedAt,
	}, nil
}

func (m *Mutation) UpdateServiceIMAPConfig(ctx context.Context, input graphql2.UpdateServiceIMAPConfigInput) (bool, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.User)
	if err != nil {
		return false, err
	}

	svcID, err := validate.ParseUUID("ServiceID", input.ServiceID)
	if err != nil {
		return false, err
	}

	// Build update parameters including OAuth credentials
	var enabled sql.NullBool
	var host, username, mailbox sql.NullString
	var oauthClientID, oauthClientSecret, oauthRefreshToken sql.NullString
	var port, pollIntervalMinutes sql.NullInt32
	var useTLS, markAsRead, deleteAfter sql.NullBool
	var includeHeaders, includeFrom, includeTo, includeSubject, includeBody sql.NullBool

	if input.Enabled != nil {
		enabled = sql.NullBool{Bool: *input.Enabled, Valid: true}
	}
	if input.Host != nil {
		host = sql.NullString{String: *input.Host, Valid: true}
	}
	if input.Port != nil {
		port = sql.NullInt32{Int32: int32(*input.Port), Valid: true}
	}
	if input.Username != nil {
		username = sql.NullString{String: *input.Username, Valid: true}
	}
	if input.UseTLS != nil {
		useTLS = sql.NullBool{Bool: *input.UseTLS, Valid: true}
	}
	if input.Mailbox != nil {
		mailbox = sql.NullString{String: *input.Mailbox, Valid: true}
	}
	if input.PollIntervalMinutes != nil {
		pollIntervalMinutes = sql.NullInt32{Int32: int32(*input.PollIntervalMinutes), Valid: true}
	}
	if input.MarkAsRead != nil {
		markAsRead = sql.NullBool{Bool: *input.MarkAsRead, Valid: true}
	}
	if input.DeleteAfter != nil {
		deleteAfter = sql.NullBool{Bool: *input.DeleteAfter, Valid: true}
	}
	if input.IncludeHeaders != nil {
		includeHeaders = sql.NullBool{Bool: *input.IncludeHeaders, Valid: true}
	}
	if input.IncludeFrom != nil {
		includeFrom = sql.NullBool{Bool: *input.IncludeFrom, Valid: true}
	}
	if input.IncludeTo != nil {
		includeTo = sql.NullBool{Bool: *input.IncludeTo, Valid: true}
	}
	if input.IncludeSubject != nil {
		includeSubject = sql.NullBool{Bool: *input.IncludeSubject, Valid: true}
	}
	if input.IncludeBody != nil {
		includeBody = sql.NullBool{Bool: *input.IncludeBody, Valid: true}
	}
	// OAuth credentials
	if input.OauthClientID != nil {
		if *input.OauthClientID != "" {
			oauthClientID = sql.NullString{String: *input.OauthClientID, Valid: true}
		} else {
			// Empty string means clear the value
			oauthClientID = sql.NullString{Valid: true}
		}
	}
	if input.OauthClientSecret != nil {
		if *input.OauthClientSecret != "" {
			oauthClientSecret = sql.NullString{String: *input.OauthClientSecret, Valid: true}
		} else {
			oauthClientSecret = sql.NullString{Valid: true}
		}
	}
	if input.OauthRefreshToken != nil {
		if *input.OauthRefreshToken != "" {
			oauthRefreshToken = sql.NullString{String: *input.OauthRefreshToken, Valid: true}
		} else {
			oauthRefreshToken = sql.NullString{Valid: true}
		}
	}

	err = gadb.New(m.DB).IMAPConfigUpdate(ctx, gadb.IMAPConfigUpdateParams{
		ServiceID:           svcID,
		Enabled:             enabled,
		Host:                host,
		Port:                port,
		Username:            username,
		UseTls:              useTLS,
		Mailbox:             mailbox,
		PollIntervalMinutes: pollIntervalMinutes,
		MarkAsRead:          markAsRead,
		DeleteAfter:         deleteAfter,
		IncludeHeaders:      includeHeaders,
		IncludeFrom:         includeFrom,
		IncludeTo:           includeTo,
		IncludeSubject:      includeSubject,
		IncludeBody:         includeBody,
		OauthClientID:       oauthClientID,
		OauthClientSecret:   oauthClientSecret,
		OauthRefreshToken:   oauthRefreshToken,
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

func (m *Mutation) DeleteServiceIMAPConfig(ctx context.Context, serviceID string) (bool, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.User)
	if err != nil {
		return false, err
	}

	svcID, err := validate.ParseUUID("ServiceID", serviceID)
	if err != nil {
		return false, err
	}

	err = gadb.New(m.DB).IMAPConfigDelete(ctx, svcID)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Service field resolvers

func (s *Service) ImapConfig(ctx context.Context, obj *service.Service) (*graphql2.ServiceIMAPConfig, error) {
	id, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, err
	}

	config, err := gadb.New(s.DB).IMAPConfigGet(ctx, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &graphql2.ServiceIMAPConfig{
		ServiceID:           config.ServiceID.String(),
		Enabled:             config.Enabled,
		Host:                config.Host,
		Port:                int(config.Port),
		Username:            config.Username,
		UseTLS:              config.UseTls,
		Mailbox:             config.Mailbox,
		PollIntervalMinutes: int(config.PollIntervalMinutes),
		MarkAsRead:          config.MarkAsRead,
		DeleteAfter:         config.DeleteAfter,
		IncludeHeaders:      config.IncludeHeaders,
		IncludeFrom:         config.IncludeFrom,
		IncludeTo:           config.IncludeTo,
		IncludeSubject:      config.IncludeSubject,
		IncludeBody:         config.IncludeBody,
		CreatedAt:           config.CreatedAt,
		UpdatedAt:           config.UpdatedAt,
	}, nil
}

func (s *Service) ImapFilterRules(ctx context.Context, obj *service.Service) ([]graphql2.IMAPFilterRule, error) {
	id, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, err
	}

	rules, err := gadb.New(s.DB).IMAPFilterRulesAll(ctx, id)
	if err != nil {
		return nil, err
	}

	result := make([]graphql2.IMAPFilterRule, len(rules))
	for i, r := range rules {
		result[i] = graphql2.IMAPFilterRule{
			ID:             r.ID.String(),
			ServiceID:      r.ServiceID.String(),
			Name:           r.Name,
			Enabled:        r.Enabled,
			FromPattern:    ptrIfValid(r.FromPattern),
			SubjectPattern: ptrIfValid(r.SubjectPattern),
			ToPattern:      ptrIfValid(r.ToPattern),
			MatchMode:      graphql2.IMAPFilterMatchMode(r.MatchMode),
			ExcludeReplies: r.ExcludeReplies,
			CreatedAt:      r.CreatedAt,
			UpdatedAt:      r.UpdatedAt,
		}
	}

	return result, nil
}


// TestIMAPConnection tests the IMAP connection with the provided configuration.
func (m *Mutation) TestIMAPConnection(ctx context.Context, input graphql2.CreateServiceIMAPConfigInput) (bool, error) {
	// Basic validation
	if input.Host == "" || input.Username == "" || input.Mailbox == "" {
		return false, fmt.Errorf("missing required fields: host, username, and mailbox are required")
	}

	// For now, just validate that the fields are present
	// A full implementation would actually connect to the IMAP server and test authentication
	// TODO: Implement actual IMAP connection test using the imapmanager package

	return true, nil
}
