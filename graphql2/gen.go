package graphql2

//go:generate rm -f mapconfig.go
//go:generate rm -f maplimit.go
//go:generate go tool gqlgen -config gqlgen.yml
//go:generate go tool configparams -out mapconfig.go
//go:generate go tool limitapigen -out maplimit.go
//go:generate go tool goimports -w mapconfig.go
//go:generate go tool goimports -w maplimit.go
