package rule

import (
	"fmt"
	"strconv"

	"github.com/antonmedv/expr/ast"
)

type FilterParser struct {
	Filters *[]Filter
	Err     *error
}

func (p FilterParser) Visit(node *ast.Node) {
	if node == nil {
		return
	}
	switch n := (*node).(type) {
	case *ast.BinaryNode:
		if n.Operator == "&&" {
			return
		}
		f, err := filterFromTreeNode(n)
		if err != nil {
			*p.Err = err
			return
		}
		*p.Filters = append(*p.Filters, f)
	default:
		return
	}
}

func filterFromTreeNode(n *ast.BinaryNode) (f Filter, err error) {
	// left node should be the field
	switch n.Left.(type) {
	case *ast.IdentifierNode:
	default:
		return Filter{}, fmt.Errorf("invalid filter string")
	}

	value := n.Right.String()
	valueType := valueTypeFromNode(n.Right)
	if valueType == StringType {
		value, err = strconv.Unquote(value)
		if err != nil {
			return Filter{}, fmt.Errorf("unquote/unescape filter value")
		}
	}

	return Filter{
		Field:     n.Left.String(),
		Operator:  n.Operator,
		Value:     value,
		ValueType: valueType,
	}, nil
}

func valueTypeFromNode(node ast.Node) ValueType {
	switch node.(type) {
	case *ast.StringNode:
		return StringType
	case *ast.IntegerNode, *ast.FloatNode:
		return NumberType
	case *ast.BoolNode:
		return BoolType
	default:
		return UnknownType
	}
}
