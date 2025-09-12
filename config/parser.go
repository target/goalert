package config

import (
	"errors"
	"strconv"
	"strings"

	"github.com/target/goalert/validation"
)

type Parser struct {
	err error
}

func NewParser() *Parser     { return &Parser{} }
func (p *Parser) Err() error { return p.err }

func (p *Parser) StringList(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func (p *Parser) Integer(s string) int {
	if s == "" {
		return 0
	}
	val, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		p.err = errors.Join(p.err, validation.NewFieldError("Value", "integer value invalid: "+err.Error()))
		return 0
	}
	return int(val)
}

func (p *Parser) Bool(s string) bool {
	if s == "" {
		return false
	}
	val, err := strconv.ParseBool(s)
	if err != nil {
		p.err = errors.Join(p.err, validation.NewFieldError("Value", "boolean value invalid: "+err.Error()))
		return false
	}
	return val
}
