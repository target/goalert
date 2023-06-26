package mockslack

import (
	"context"
	"net/http"
	"sort"
)

func (st *API) UserGroupList(ctx context.Context) ([]UserGroup, error) {
	err := checkPermission(ctx, "bot", "usergroups:read")
	if err != nil {
		return nil, err
	}

	st.mx.Lock()
	defer st.mx.Unlock()

	ids := make([]string, 0, len(st.usergroups))
	for id := range st.usergroups {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	result := make([]UserGroup, 0, len(ids))
	for _, id := range ids {
		ch := st.usergroups[id]
		result = append(result, ch.UserGroup)
	}

	return result, nil
}

func (s *Server) ServeUserGroupList(w http.ResponseWriter, req *http.Request) {
	ug, err := s.API().UserGroupList(req.Context())
	if respondErr(w, err) {
		return
	}

	var resp struct {
		response
		UserGroups []UserGroup `json:"usergroups"`
	}

	resp.UserGroups = ug

	respondWith(w, resp)
}
