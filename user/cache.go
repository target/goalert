package user

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"strings"
	"time"

	"github.com/golang/groupcache"
	uuid "github.com/satori/go.uuid"
	"github.com/target/goalert/permission"
)

func timeKey(key string, dur time.Duration) string {
	return key + "\n" + time.Now().Round(dur).Format(time.RFC3339)
}
func keyName(keyWithTime string) string {
	return strings.SplitN(keyWithTime, "\n", 2)[0]
}

// ExistanceChecker allows checking if various users exist. Done must be called when finished.
type ExistanceChecker interface {
	UserExistsString(id string) bool
	UserExistsUUID(id uuid.UUID) bool
	Done()
}

type checker struct {
	m  map[uuid.UUID]struct{}
	ch chan map[uuid.UUID]struct{}
}

func (c *checker) UserExistsString(id string) bool { return c.UserExistsUUID(uuid.FromStringOrNil(id)) }
func (c *checker) UserExistsUUID(id uuid.UUID) bool {
	_, ok := c.m[id]
	return ok
}
func (c *checker) Done() {
	if c.m == nil {
		return
	}
	c.ch <- c.m
	c.m = nil
}

// UserExists returns an ExistanceChecker.
func (db *DB) UserExists(ctx context.Context) (ExistanceChecker, error) {
	err := permission.LimitCheckAny(ctx)
	if err != nil {
		return nil, err
	}
	m, err := db.userExistMap(ctx)
	if err != nil {
		return nil, err
	}
	return &checker{
		m:  m,
		ch: db.userExist,
	}, nil
}

func (db *DB) userExistMap(ctx context.Context) (map[uuid.UUID]struct{}, error) {
	var data []byte
	err := db.grp.Get(ctx, timeKey("userIDs", time.Minute), groupcache.AllocatingByteSliceSink(&data))
	if err != nil {
		return nil, err
	}

	m := <-db.userExist
	if bytes.Equal(data[:sha256.Size], db.userExistHash) {
		return m, nil
	}

	var ids []uuid.UUID
	err = binary.Read(bytes.NewReader(data[sha256.Size:]), binary.BigEndian, &ids)
	if err != nil {
		db.userExist <- m
		return nil, err
	}

	for k := range m {
		delete(m, k)
	}

	for _, id := range ids {
		m[id] = struct{}{}
	}
	db.userExistHash = data[:sha256.Size]

	return m, nil
}

func (db *DB) currentUserIDs(ctx context.Context) (result []byte, err error) {
	rows, err := db.ids.QueryContext(ctx)
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

func (db *DB) cacheGet(ctx context.Context, key string, dest groupcache.Sink) error {
	switch keyName(key) {
	case "userIDs":
		data, err := db.currentUserIDs(ctx)
		if err != nil {
			return err
		}
		return dest.SetBytes(data)
	}
	return nil
}
