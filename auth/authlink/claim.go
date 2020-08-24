package authlink

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/target/goalert/util/log"
)

const (
	issuer   = "GoAlert AuthLink"
	audience = "auth-token"
)

// ClaimRefreshWait will refresh the available claims.
func (s *Store) ClaimRefreshWait(ctx context.Context) error {
	s.mx.Lock()
	startCh := s.startRefresh
	waitCh := s.waitRefresh
	s.mx.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-startCh:
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-waitCh:
	}

	return nil
}

func (s *Store) claimUpdates() {
	defer close(s.waitRefresh)
	defer close(s.startRefresh)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t := time.NewTicker(MaxClaimQueryFrequency)
	defer t.Stop()
	for {
		select {
		case <-s.closeCh:
			return
		case <-t.C:
			// min wait time
		}

		select {
		case <-s.closeCh:
			return
		case s.startRefresh <- struct{}{}:
			// got start request
		}

		s.refreshClaims(ctx)

		s.mx.Lock()
		close(s.startRefresh)
		close(s.waitRefresh)
		s.startRefresh = make(chan struct{})
		s.waitRefresh = make(chan struct{})
		s.mx.Unlock()
	}
}

func (s *Store) refreshClaims(ctx context.Context) {
	rows, err := s.unclaimed.QueryContext(ctx)
	if err != nil {
		log.Log(ctx, err)
		return
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			log.Log(ctx, err)
			return
		}
		ids = append(ids, id)
	}
	s.wl.Set(ids)
}

type ClaimResponse struct {
	ExpiresAt  time.Time
	VerifyCode string
	AuthToken  string
}

func (s *Store) Claim(ctx context.Context, claimCode string, refresh bool) (*ClaimResponse, error) {
	var idStr, verifyCode string
	var claimedAt, expiresAt time.Time
	err := s.wl.LockRemove(ctx, claimCode, func(ctx context.Context) error {
		err := s.claim.QueryRowContext(ctx, claimCode).Scan(&idStr, &claimedAt, &verifyCode, &expiresAt)
		if err == sql.ErrNoRows {
			return ErrBadID
		}
		return err
	})
	if refresh && errors.Is(err, ErrBadID) {
		err = s.ClaimRefreshWait(ctx)
		if err != nil {
			return nil, err
		}
		return s.Claim(ctx, claimCode, false)
	}
	if err != nil {
		return nil, err
	}

	tok, err := s.keys.SignJWT(jwt.StandardClaims{
		Subject:   idStr,
		ExpiresAt: expiresAt.Unix(),
		NotBefore: claimedAt.Unix(),
		IssuedAt:  time.Now().Unix(),
		Issuer:    issuer,
		Audience:  audience,
	})
	if err != nil {
		return nil, err
	}

	return &ClaimResponse{
		ExpiresAt:  expiresAt,
		VerifyCode: verifyCode,
		AuthToken:  tok,
	}, nil
}
