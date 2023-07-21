package graphql2

import (
	_ "embed"
)

//go:embed schema.graphql
var schema string

func Schema() string {
	return schema
}
