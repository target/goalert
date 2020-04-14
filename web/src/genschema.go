package src

//go:generate go run ../../devtools/gqltsgen/main.go -out ./schema.d.ts ../../graphql2/schema.graphql
//go:generate ./node_modules/.bin/prettier -l --write schema.d.ts
