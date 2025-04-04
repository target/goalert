package src

//go:generate go tool gqltsgen -out ./schema.d.ts ../../graphql2/schema.graphql ../../graphql2/graph/*.graphqls
//go:generate go tool configtsidgen -out ./schema.d.ts
//go:generate go tool expflagtsgen -out ./expflag.d.ts
//go:generate make -C ../.. node_modules
//go:generate ../../bin/tools/bun run prettier -l --write *.d.ts
