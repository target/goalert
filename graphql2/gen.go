package graphql2

//go:generate rm -f mapconfig.go
//go:generate go run ../devtools/gqlgen/gqlgen.go -config gqlgen.yml
//go:generate go run ../devtools/configparams/main.go -out mapconfig.go
//go:generate go run golang.org/x/tools/cmd/goimports -w mapconfig.go
