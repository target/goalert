package smoketest

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

func TestListAlerts(t *testing.T) {
	t.Parallel()

	sql := `
	insert into escalation_policies (id, name)
	values
		({{uuid "eid"}}, 'esc policy');
`
	for i := 0; i < 160; i++ {
		name := "s" + strconv.Itoa(i)
		sql += `
			insert into services (id, name, escalation_policy_id) values ({{uuid "` + name + `"}}, '` + name + `', {{uuid "eid"}});
			insert into alerts (service_id, description) values ({{uuid "` + name + `"}}, 'hi');
		`
	}

	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	resp := h.GraphQLQuery(`
		query {
			alerts {
				id: _id
				service_id
				service {id, name}
			}
		}
	`)
	for _, err := range resp.Errors {
		t.Error("fetch alerts:", err)
	}

	var res struct {
		Alerts []struct {
			ID      int
			Service struct{ Name string }
		}
	}

	err := json.Unmarshal(resp.Data, &res)
	if err != nil {
		t.Fatal("failed to parse response:", err)
	}

	if len(res.Alerts) == 0 {
		t.Error("got 0 alerts; expected at least 1")
	}
	for _, a := range res.Alerts {
		name := "s" + strconv.Itoa(a.ID-1)
		if a.Service.Name != name {
			t.Errorf("Alert[%d].Service.Name = %s; want %s", a.ID, a.Service.Name, name)
		}
	}
}
