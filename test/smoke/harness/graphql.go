package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/auth"
	"github.com/target/goalert/limit"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/user"
)

// DefaultGraphQLAdminUserID is the UserID created & used for GraphQL calls by default.
const DefaultGraphQLAdminUserID = "00000000-0000-0000-0000-000000000002"

func (h *Harness) createGraphQLUser(userID string) {
	h.t.Helper()
	permission.SudoContext(context.Background(), func(ctx context.Context) {
		h.t.Helper()
		_, err := h.backend.UserStore.Insert(ctx, &user.User{
			Name: "GraphQL User",
			ID:   userID,
			Role: permission.RoleAdmin,
		})
		if err != nil {
			h.t.Fatal(errors.Wrap(err, "create GraphQL user"))
		}
	})
}

func (h *Harness) createGraphQLSession(userID string) string {
	h.t.Helper()
	tok, err := h.backend.AuthHandler.CreateSession(context.Background(), "goalert-smoketest", userID)
	if err != nil {
		h.t.Fatal(errors.Wrap(err, "create auth session"))
	}

	h.gqlSessions[userID], err = tok.Encode(h.backend.SessionKeyring.Sign)
	if err != nil {
		h.t.Fatal(errors.Wrap(err, "sign auth session"))
	}

	return h.gqlSessions[userID]
}

// GraphQLQuery2 will perform a GraphQL2 query against the backend, internally
// handling authentication. Queries are performed with Admin role.
func (h *Harness) GraphQLQuery2(query string) *QLResponse {
	h.t.Helper()
	return h.GraphQLQueryT(h.t, query)
}

// SetConfigValue will update the config value id (e.g. `General.PublicURL`) to the provided value.
func (h *Harness) SetConfigValue(id, value string) {
	h.t.Helper()
	res := h.GraphQLQuery2(fmt.Sprintf(`mutation{setConfig(input:[{id: %s, value: %s}])}`, strconv.Quote(id), strconv.Quote(value)))
	assert.Empty(h.t, res.Errors)

	// wait for engine cycle to complete to ensure next action
	// uses new config only
	h.Trigger()
}

// SetSystemLimit will update the value of a system limit given an id (e.g. `RulesPerSchedule`).
// TODO repalce SetSystemLimit with new mutation (work anticipated to be done with Admin Config view)
func (h *Harness) SetSystemLimit(id limit.ID, value int) {
	h.t.Helper()
	h.execQuery(fmt.Sprintf(`
		UPDATE config_limits
		SET max = %d
		WHERE id='%s'; 
	`, value, id), nil)
}

// GraphQLQueryT will perform a GraphQL query against the backend, internally
// handling authentication. Queries are performed with Admin role.
func (h *Harness) GraphQLQueryT(t *testing.T, query string) *QLResponse {
	t.Helper()
	return h.GraphQLQueryUserT(t, DefaultGraphQLAdminUserID, query)
}

func (h *Harness) GraphQLToken(userID string) string {
	h.mx.Lock()
	tok := h.gqlSessions[userID]
	if tok == "" {
		if userID == DefaultGraphQLAdminUserID {
			h.createGraphQLUser(userID)
		}
		tok = h.createGraphQLSession(userID)
	}
	h.mx.Unlock()
	return tok
}

// GraphQLQueryUserT will perform a GraphQL query against the backend, internally
// handling authentication. Queries are performed with the provided UserID.
func (h *Harness) GraphQLQueryUserT(t *testing.T, userID, query string) *QLResponse {
	t.Helper()
	return h.GraphQLQueryUserVarsT(t, userID, query, "", nil)
}

// GraphQLQueryUserT will perform a GraphQL query against the backend, internally
// handling authentication. Queries are performed with the provided UserID.
func (h *Harness) GraphQLQueryUserVarsT(t *testing.T, userID, query, opName string, vars any) *QLResponse {
	t.Helper()
	retry := 1
	var err error
	var resp *http.Response

	tok := h.GraphQLToken(userID)

	for {
		query = strings.Replace(query, "\t", "", -1)
		q := struct {
			Query         string
			Variables     any    `json:"variables,omitempty"`
			OperationName string `json:"operationName,omitempty"`
		}{Query: query, Variables: vars, OperationName: opName}

		data, err := json.Marshal(q)
		if err != nil {
			h.t.Fatal("failed to marshal graphql query")
		}
		t.Log("Query:", query)

		url := h.URL() + "/api/graphql"
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
		if err != nil {
			t.Fatal("failed to make request:", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  auth.CookieName,
			Value: tok,
		})
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal("failed to make http request:", err)
		}
		if resp.StatusCode == 401 && retry > 0 {
			h.t.Logf("GraphQL request failed with 401, retrying with new session (%d left)", retry)
			h.mx.Lock()
			tok = h.createGraphQLSession(userID)
			h.mx.Unlock()
			resp.Body.Close()
			retry--
			continue
		}
		if resp.StatusCode != 200 {
			data, _ := io.ReadAll(resp.Body)
			t.Fatal("failed to make graphql request:", resp.Status, string(data))
		}
		break
	}
	defer resp.Body.Close()
	var r QLResponse
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Fatal("failed to parse GraphQL response:", err)
	}
	return &r
}

// QLResponse is a generic GraphQL response.
type QLResponse struct {
	Data   json.RawMessage
	Errors []struct{ Message string }
}
