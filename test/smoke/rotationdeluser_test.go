package smoke

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/target/goalert/test/smoke/harness"
)

// RotationDelUser tests that rotations preserve the active user when a user is deleted.
func TestRotationDelUser(t *testing.T) {
	t.Parallel()

	const sql = `
	insert into users (id, name, email)
	values
		({{uuid "uid1"}}, 'bob', 'joe'),
		({{uuid "uid2"}}, 'ben', 'frank'),
		({{uuid "uid3"}}, 'joe', 'bob');

	insert into rotations (id, name, type, start_time, shift_length, time_zone)
	values
		({{uuid "rot1"}}, 'default rotation', 'daily', now(), 1, 'UTC');

	insert into rotation_participants (rotation_id, user_id, position)
	values
		({{uuid "rot1"}}, {{uuid "uid1"}}, 0),
		({{uuid "rot1"}}, {{uuid "uid2"}}, 1),
		({{uuid "rot1"}}, {{uuid "uid3"}}, 2);
	`
	h := harness.NewHarness(t, sql, "add-daily-alert-metrics")
	defer h.Close()

	h.GraphQLQuery2(fmt.Sprintf(
		`mutation{updateRotation(input:{id:"%s", activeUserIndex: 1})}`,
		h.UUID("rot1"),
	))

	h.GraphQLQuery2(fmt.Sprintf(
		`mutation{deleteAll(input:[{id:"%s", type: user}])}`,
		h.UUID("uid1"),
	))

	resp := h.GraphQLQuery2(fmt.Sprintf(
		`query{rotation(id:"%s"){activeUserIndex}}`,
		h.UUID("rot1"),
	))
	var data struct {
		Rotation struct {
			ActiveUserIndex int
		}
	}
	err := json.Unmarshal(resp.Data, &data)
	if err != nil {
		t.Fatal(err)
		return
	}
	// 2nd user is now first, so index should be zero
	if data.Rotation.ActiveUserIndex != 0 {
		t.Errorf("expected activeUserIndex to be 0, got %d", data.Rotation.ActiveUserIndex)
	}
}
