package sysapiapp

import (
	context "context"

	uuid "github.com/satori/go.uuid"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/sysapi"
	"github.com/target/goalert/user"
)

type Server struct {
	UserStore user.Store
	sysapi.UnimplementedSysAPIServer
}

func (srv *Server) Echo(ctx context.Context, req *sysapi.EchoRequest) (*sysapi.EchoResponse, error) {
	return &sysapi.EchoResponse{Data: req.Data}, nil
}

func (srv *Server) AuthSubjects(ctx context.Context, req *sysapi.AuthSubjectsRequest) (*sysapi.AuthSubjectsResponse, error) {
	ctx = permission.SystemContext(ctx, "SystemAPI")

	sub, err := srv.UserStore.FindSomeAuthSubjectsForProvider(ctx, int(req.Limit), req.AfterSubjectId, req.ProviderId)
	if err != nil {
		return nil, err
	}

	var subs []*sysapi.AuthSubject
	for _, s := range sub {
		subs = append(subs, &sysapi.AuthSubject{
			ProviderId: s.ProviderID,
			SubjectId:  s.SubjectID,
			UserId:     uuid.FromStringOrNil(s.UserID).Bytes(),
		})
	}

	return &sysapi.AuthSubjectsResponse{Subjects: subs}, nil
}

func (srv *Server) ListUsers(ctx context.Context, req *sysapi.ListUsersRequest) (*sysapi.ListUsersResponse, error) {
	ctx = permission.SystemContext(ctx, "SystemAPI")

	users, err := srv.UserStore.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var resUsers []*sysapi.UserInfo
	for _, u := range users {
		resUsers = append(resUsers, &sysapi.UserInfo{
			Id:    u.ID,
			Name:  u.Name,
			Email: u.Email,
		})
	}

	return &sysapi.ListUsersResponse{Users: resUsers}, nil
}
