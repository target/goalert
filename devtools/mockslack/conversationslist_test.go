package mockslack

import (
	"context"
	"sort"
	"testing"
)

func TestState_ConversationsList(t *testing.T) {
	st := newState()

	chans := []ChannelInfo{
		st.NewChannel("foo"),
		st.NewChannel("bar"),
		st.NewChannel("baz"),
	}
	sort.Slice(chans, func(i, j int) bool { return chans[i].ID < chans[j].ID })

	ctx := context.Background()
	ch, _, err := st.API().ConversationsList(ctx, ConversationsListOpts{Limit: 1})
	if err == nil {
		t.Error("got nil; expected permissions error")
	}

	ctx = WithToken(ctx, &AuthToken{Scopes: []string{"channels:read"}})

	check := func(idx int) {
		t.Helper()
		if len(ch) != 1 {
			t.Fatalf("got len=%d; want 1", len(ch))
		}
		if ch[0].ID != chans[idx].ID {
			t.Errorf("ID[%d]=%s; want %s", idx, ch[0].ID, chans[idx].ID)
		}
		if ch[0].Name != chans[idx].Name {
			t.Errorf("Name[%d]=%s; want %s", idx, ch[0].Name, chans[idx].Name)
		}
	}

	ch, cur, err := st.API().ConversationsList(ctx, ConversationsListOpts{Limit: 1})
	if err != nil {
		t.Errorf("err=%v; expected nil", err)
	}
	if cur == "" {
		t.Errorf("got empty cursor; expected next page")
	}
	check(0)

	ch, cur, err = st.API().ConversationsList(ctx, ConversationsListOpts{Limit: 1, Cursor: cur})
	if err != nil {
		t.Errorf("err=%v; expected nil", err)
	}
	if cur == "" {
		t.Errorf("got empty cursor; expected next page")
	}
	check(1)

	ch, cur, err = st.API().ConversationsList(ctx, ConversationsListOpts{Limit: 1, Cursor: cur})
	if err != nil {
		t.Errorf("err=%v; expected nil", err)
	}
	if cur != "" {
		t.Errorf("cursor=%s; expected empty", cur)
	}
	check(2)

}
