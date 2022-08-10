package swoinfo

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/swo/swodb"
)

type ConnCount struct {
	Name   string
	IsNext bool
	Count  int
}

// ConnInfo provides information about the connections to both old and new databases.
func ConnInfo(ctx context.Context, oldConn, newConn *pgx.Conn) ([]ConnCount, error) {
	oldConns, err := swodb.New(oldConn).ConnectionInfo(ctx)
	if err != nil {
		return nil, err
	}
	newConns, err := swodb.New(newConn).ConnectionInfo(ctx)
	if err != nil {
		return nil, err
	}

	type connType struct {
		Name   string
		IsNext bool
	}
	counts := make(map[connType]int)
	for _, oldConn := range oldConns {
		counts[connType{Name: oldConn.Name.String}] += int(oldConn.Count)
	}
	for _, newConn := range newConns {
		counts[connType{Name: newConn.Name.String, IsNext: true}] += int(newConn.Count)
	}

	var result []ConnCount
	for t, count := range counts {
		result = append(result, ConnCount{Name: t.Name, IsNext: t.IsNext, Count: count})
	}

	return result, nil
}
