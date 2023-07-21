package gqlauth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/target/goalert/graphql2"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/validator"
)

var s = gqlparser.MustLoadSchema(&ast.Source{Input: graphql2.Schema()})

type Query struct {
	doc *ast.QueryDocument

	allowedFields map[Field]struct{}
	allowedArgs   map[FieldArg][]AllowedArg
}
type Field struct {
	Name   string
	Object string
	Kind   ast.DefinitionKind
}

type FieldArg struct {
	Field
	Arg string
}

type AllowedArg struct {
	Raw      string
	Variable bool
}

func NewQuery(query string) (*Query, error) {
	doc, err := gqlparser.LoadQuery(s, query)
	if err != nil {
		return nil, err
	}

	q := &Query{
		doc:           doc,
		allowedFields: make(map[Field]struct{}),
		allowedArgs:   make(map[FieldArg][]AllowedArg),
	}

	var errs []error
	var e validator.Events
	e.OnField(func(w *validator.Walker, d *ast.Field) {
		f := Field{
			Name:   d.Definition.Name,
			Object: d.ObjectDefinition.Name,
			Kind:   d.ObjectDefinition.Kind,
		}
		q.allowedFields[f] = struct{}{}

		for _, arg := range d.Arguments {
			a := FieldArg{
				Field: f,
				Arg:   arg.Name,
			}

			fmt.Println(arg.Name, arg.Value.Kind)
			q.allowedArgs[a] = append(q.allowedArgs[a], AllowedArg{
				Raw:      arg.Value.Raw,
				Variable: arg.Value.Kind == ast.Variable,
			})
		}
	})
	validator.Walk(s, doc, &e)

	return q, errors.Join(errs...)
}

func (q *Query) ValidateField(af *ast.Field) error {
	f := Field{
		Name:   af.Definition.Name,
		Object: af.ObjectDefinition.Name,
		Kind:   af.ObjectDefinition.Kind,
	}
	if strings.HasPrefix(f.Name, "__") {
		return nil
	}
	if strings.HasPrefix(f.Object, "__") {
		return nil
	}

	if _, ok := q.allowedFields[f]; !ok {
		return fmt.Errorf("field %s.%s not allowed", f.Object, f.Name)
	}
argCheck:
	for _, arg := range af.Arguments {
		fmt.Println(arg.Name, q.allowedArgs)
		a := FieldArg{
			Field: f,
			Arg:   arg.Name,
		}
		if aaList, ok := q.allowedArgs[a]; ok {
			for _, aa := range aaList {

				if aa.Variable {
					continue argCheck
				}
				if aa.Raw == arg.Value.Raw {
					continue argCheck
				}
			}
		}

		return fmt.Errorf("field %s.%s arg %s not allowed or has a forbidden value", f.Object, f.Name, arg.Name)
	}

	return nil
}

// IsSubset checks if a is a subset of b.
func (q *Query) IsSubset(query string) (bool, error) {
	doc, err := gqlparser.LoadQuery(s, query)
	if err != nil {
		return false, err
	}

	var invalid bool
	var e validator.Events
	e.OnField(func(w *validator.Walker, d *ast.Field) {
		if q.ValidateField(d) != nil {
			invalid = true
		}
	})
	validator.Walk(s, doc, &e)

	return !invalid, nil
}
