package sysapiserver

import (
	"context"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/pkg/sysapi"
	"github.com/target/goalert/user"
)

type Server struct {
	UserStore *user.Store
	sysapi.UnimplementedSysAPIServer
}

func (srv *Server) UsersWithoutAuthProvider(req *sysapi.UsersWithoutAuthProviderRequest, rSrv sysapi.SysAPI_UsersWithoutAuthProviderServer) error {
	ctx := permission.SystemContext(rSrv.Context(), "SystemAPI")
	return srv.UserStore.WithoutAuthProviderFunc(ctx, req.ProviderId, func(u user.User) error {
		return rSrv.Send(&sysapi.UserInfo{
			Id:    u.ID,
			Name:  u.Name,
			Email: u.Email,
		})
	})
}

func (srv *Server) SetAuthSubject(ctx context.Context, req *sysapi.SetAuthSubjectRequest) (*sysapi.SetAuthSubjectResponse, error) {
	ctx = permission.SystemContext(ctx, "SystemAPI")

	return &sysapi.SetAuthSubjectResponse{}, srv.UserStore.SetAuthSubject(ctx, req.Subject.ProviderId, req.Subject.SubjectId, req.Subject.UserId)
}

func (srv *Server) AuthSubjects(req *sysapi.AuthSubjectsRequest, rSrv sysapi.SysAPI_AuthSubjectsServer) error {
	ctx := permission.SystemContext(rSrv.Context(), "SystemAPI")

	return srv.UserStore.AuthSubjectsFunc(ctx, req.ProviderId, func(s user.AuthSubject) error {
		return rSrv.Send(&sysapi.AuthSubject{
			ProviderId: s.ProviderID,
			SubjectId:  s.SubjectID,
			UserId:     s.UserID,
		})
	}, req.UserId)
}

func (srv *Server) DeleteUser(ctx context.Context, req *sysapi.DeleteUserRequest) (*sysapi.DeleteUserResponse, error) {
	ctx = permission.SystemContext(ctx, "SystemAPI")
	err := srv.UserStore.Delete(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &sysapi.DeleteUserResponse{}, nil
}
