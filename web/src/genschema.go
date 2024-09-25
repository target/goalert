package src

//go:generate go run ../../devtools/gqltsgen/. -out ./schema.d.ts ../../graphql2/schema.graphql ../../graphql2/graph/*.graphqls
//go:generate go run ../../devtools/configtsidgen/. -out ./schema.d.ts
//go:generate go run ../../expflag/cmd/tsgen/. -out ./expflag.d.ts
//go:generate ../../bin/tools/bun run prettier -l --write *.d.ts
