package swoinfo

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/swo/swodb"
)

type ConnCount struct {
	Name  string
	Count int
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

	counts := make(map[string]int)
	for _, oldConn := range oldConns {
		counts[oldConn.Name.String] += int(oldConn.Count)
	}
	for _, newConn := range newConns {
		counts[newConn.Name.String] += int(newConn.Count)
	}

	var result []ConnCount
	for name, count := range counts {
		result = append(result, ConnCount{Name: name, Count: count})
	}

	return result, nil
}
