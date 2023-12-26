package graphql2

import (
	_ "embed"
	"sort"

	"github.com/target/goalert/validation"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/validator"
)

var schemaFields []string
var astSchema *ast.Schema

func init() {
	astSchema = NewExecutableSchema(Config{}).Schema()

	for _, typ := range astSchema.Types {
		if typ.Kind != ast.Object {
			continue
		}
		for _, f := range typ.Fields {
			schemaFields = append(schemaFields, typ.Name+"."+f.Name)
		}
	}
	sort.Strings(schemaFields)
}

// SchemaFields will return a list of all fields in the schema.
func SchemaFields() []string { return schemaFields }

// QueryFields will return a list of all fields that the given query references.
func QueryFields(query string) ([]string, error) {
	qDoc, qErr := gqlparser.LoadQuery(astSchema, query)
	if len(qErr) > 0 {
		return nil, validation.NewFieldError("Query", qErr.Error())
	}

	var fields []string
	var e validator.Events
	e.OnField(func(w *validator.Walker, field *ast.Field) {
		fields = append(fields, field.ObjectDefinition.Name+"."+field.Name)
	})
	validator.Walk(astSchema, qDoc, &e)

	sort.Strings(fields)
	return fields, nil
}
