package user

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/golang/groupcache"
	"github.com/google/uuid"
	"github.com/target/goalert/permission"
)

func timeKey(key string, dur time.Duration) string {
	return key + "\n" + time.Now().Round(dur).Format(time.RFC3339)
}
func keyName(keyWithTime string) string {
	return strings.SplitN(keyWithTime, "\n", 2)[0]
}

// ExistanceChecker allows checking if various users exist.
type ExistanceChecker interface {
	UserExistsString(id string) bool
	UserExistsUUID(id uuid.UUID) bool
}

type checker map[uuid.UUID]struct{}

func (c checker) UserExistsString(idStr string) bool {
	if idStr == "" {
		return false
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return false
	}

	return c.UserExistsUUID(id)
}
func (c checker) UserExistsUUID(id uuid.UUID) bool { _, ok := c[id]; return ok }

// UserExists returns an ExistanceChecker.
func (s *Store) UserExists(ctx context.Context) (ExistanceChecker, error) {
	err := permission.LimitCheckAny(ctx)
	if err != nil {
		return nil, err
	}
	m, err := s.userExistMap(ctx)
	if err != nil {
		return nil, err
	}
	return checker(m), nil
}

func (s *Store) userExistMap(ctx context.Context) (map[uuid.UUID]struct{}, error) {
	var data []byte
	tk := timeKey("userIDs", time.Minute)
	fmt.Println("tk: ", tk)
	sink := groupcache.AllocatingByteSliceSink(&data)
	fmt.Printf("sink: %v\n", sink)
	err := s.grp.Get(ctx, tk, sink)
	if err != nil {
		return nil, err
	}

	m := <-s.userExist
	if bytes.Equal(data[:sha256.Size], s.userExistHash) {
		s.userExist <- m
		return m, nil
	}

	ids := make([]uuid.UUID, (len(data)-sha256.Size)/16)
	err = binary.Read(bytes.NewReader(data[sha256.Size:]), binary.BigEndian, &ids)
	if err != nil {
		s.userExist <- m
		return nil, err
	}

	m = make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		m[id] = struct{}{}
	}
	s.userExistHash = data[:sha256.Size]
	s.userExist <- m
	return m, nil
}

func (s *Store) currentUserIDs(ctx context.Context) (result []byte, err error) {
	rows, err := s.ids.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []uuid.UUID
	sum := sha256.New()
	for rows.Next() {
		var id uuid.UUID
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		sum.Write(id[:])
		data = append(data, id)
	}

	buf := bytes.NewBuffer(nil)
	buf.Write(sum.Sum(nil))
	err = binary.Write(buf, binary.BigEndian, data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *Store) cacheGet(ctx context.Context, key string, dest groupcache.Sink) error {
	switch keyName(key) {
	case "userIDs":
		data, err := s.currentUserIDs(ctx)
		if err != nil {
			return err
		}
		return dest.SetBytes(data)
	}
	return nil
}
