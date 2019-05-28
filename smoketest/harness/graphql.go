package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/target/goalert/auth"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/user"
)

func (h *Harness) insertGraphQLUser() {
	h.t.Helper()
	var err error
	permission.SudoContext(context.Background(), func(ctx context.Context) {
		_, err = h.usr.Insert(ctx, &user.User{
			Name: "GraphQL User",
			ID:   "bcefacc0-4764-012d-7bfb-002500d5decb",
			Role: permission.RoleAdmin,
		})
	})
	if err != nil {
		h.t.Fatal(errors.Wrap(err, "create GraphQL user"))
	}

	h.sessToken, _, err = h.authH.CreateSession(context.Background(), "goalert-smoketest", "bcefacc0-4764-012d-7bfb-002500d5decb")
	if err != nil {
		h.t.Fatal(errors.Wrap(err, "create auth session"))
	}
}

// GraphQLQuery will perform a GraphQL query against the backend, internally
// handling authentication. Queries are performed with Admin role.
func (h *Harness) GraphQLQuery(query string) *QLResponse {
	h.t.Helper()
	return h.GraphQLQueryT(h.t, query, "/v1/graphql")
}

// GraphQLQuery2 will perform a GraphQL2 query against the backend, internally
// handling authentication. Queries are performed with Admin role.
func (h *Harness) GraphQLQuery2(query string) *QLResponse {
	h.t.Helper()
	return h.GraphQLQueryT(h.t, query, "/v1/graphql2")
}

// GraphQLQueryT will perform a GraphQL query against the backend, internally
// handling authentication. Queries are performed with Admin role.
func (h *Harness) GraphQLQueryT(t *testing.T, query string, u string) *QLResponse {
	t.Helper()
	h.addGraphUser.Do(h.insertGraphQLUser)
	query = strings.Replace(query, "\t", "", -1)
	q := struct{ Query string }{Query: query}

	data, err := json.Marshal(q)
	if err != nil {
		h.t.Fatal("failed to marshal graphql query")
	}
	t.Log("Query:", query)

	url := h.URL() + u
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		t.Fatal("failed to make request:", err)
	}
	req.AddCookie(&http.Cookie{
		Name:  auth.CookieName,
		Value: h.sessToken,
	})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal("failed to make http request:", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatal("failed to make graphql request:", resp.Status)
	}

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
