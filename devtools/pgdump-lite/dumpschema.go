package pgdump

import (
	"context"
	"fmt"
	"strings"

	"github.com/target/goalert/devtools/pgdump-lite/pgd"
)

type Schema struct {
	Extensions []Extension
	Functions  []Function
	Tables     []Table
	Enums      []Enum
	Sequences  []Sequence
}

func (s Schema) String() string {
	var b strings.Builder
	b.WriteString("-- Extensions\n\n")
	for _, e := range s.Extensions {
		b.WriteString(e.String())
		b.WriteString("\n\n")
	}

	b.WriteString("-- Enums\n\n")
	for _, e := range s.Enums {
		b.WriteString(e.String())
		b.WriteString("\n\n")
	}

	b.WriteString("-- Functions\n\n")
	for _, e := range s.Functions {
		b.WriteString(e.String())
		b.WriteString("\n\n")
	}

	b.WriteString("-- Tables\n\n")
	for _, e := range s.Tables {
		b.WriteString(e.String())
		b.WriteString("\n\n")
	}

	b.WriteString("-- Sequences\n\n")
	for _, e := range s.Sequences {
		b.WriteString(e.String())
		b.WriteString("\n\n")
	}

	return b.String()
}

type Index struct {
	Name string
	Def  string
}

func (idx Index) EntityName() string { return idx.Name }
func (idx Index) String() string     { return idx.Def + ";" }

type Trigger struct {
	Name string
	Def  string
}

func (t Trigger) EntityName() string { return t.Name }
func (t Trigger) String() string     { return t.Def + ";" }

type Sequence struct {
	Name       string
	StartValue int64
	Increment  int64
	MinValue   int64
	MaxValue   int64
	Cache      int64

	OwnedBy string
}

func (s Sequence) EntityName() string { return s.Name }
func (s Sequence) String() string {
	def := fmt.Sprintf("CREATE SEQUENCE %s\n\tSTART WITH %d\n\tINCREMENT BY %d\n\tMINVALUE %d\n\tMAXVALUE %d\n\tCACHE %d",
		s.Name, s.StartValue, s.Increment, s.MinValue, s.MaxValue, s.Cache)

	if s.OwnedBy == "" {
		return def + ";"
	}

	return fmt.Sprintf("%s\n\tOWNED BY%s;",
		def,
		s.OwnedBy,
	)
}

type Extension struct {
	Name string
}

func (e Extension) EntityName() string { return e.Name }
func (e Extension) String() string {
	return fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %s;", e.Name)
}

type Function struct {
	Name string
	Def  string
}

func (f Function) EntityName() string { return f.Name }
func (f Function) String() string     { return f.Def + ";" }

type Enum struct {
	Name   string
	Values []string
}

func (e Enum) EntityName() string { return e.Name }
func (e Enum) String() string {
	return fmt.Sprintf("CREATE TYPE %s AS ENUM (\n\t'%s'\n);", e.Name, strings.Join(e.Values, "',\n\t'"))
}

type Table struct {
	Name string

	Columns     []Column
	Constraints []Constraint
	Indexes     []Index
	Triggers    []Trigger
	Sequences   []Sequence
}

func (t Table) EntityName() string { return t.Name }
func (t Table) String() string {
	var lines []string
	for _, c := range t.Columns {
		lines = append(lines, c.String())
	}
	for _, c := range t.Constraints {
		lines = append(lines, c.String())
	}

	var b strings.Builder
	fmt.Fprintf(&b, "CREATE TABLE %s (\n\t%s\n);\n", t.Name, strings.Join(lines, ",\n\t"))

	if len(t.Indexes) > 0 {
		b.WriteString("\n")
	}
	for _, idx := range t.Indexes {
		b.WriteString(idx.String())
		b.WriteString("\n")
	}

	if len(t.Triggers) > 0 {
		b.WriteString("\n")
	}
	for _, trg := range t.Triggers {
		b.WriteString(trg.String())
		b.WriteString("\n")
	}
	return b.String()
}

type Constraint struct {
	Name string
	Def  string
}

func (c Constraint) EntityName() string { return c.Name }
func (c Constraint) String() string {
	return fmt.Sprintf("CONSTRAINT %s %s", c.Name, c.Def)
}

type Column struct {
	Name         string
	Type         string
	NotNull      bool
	DefaultValue string
}

func (c Column) EntityName() string { return c.Name }
func (c Column) String() string {
	var def string
	if c.DefaultValue != "" {
		def = fmt.Sprintf(" DEFAULT %s", c.DefaultValue)
	}
	if c.NotNull {
		def += " NOT NULL"
	}
	return fmt.Sprintf("%s %s%s", c.Name, c.Type, def)
}

func DumpSchema(ctx context.Context, conn pgd.DBTX) (*Schema, error) {
	db := pgd.New(conn)

	var s Schema

	// list extensions
	exts, err := db.ListExtensions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list extensions: %w", err)
	}
	for _, e := range exts {
		s.Extensions = append(s.Extensions, Extension{Name: e.ExtName})
	}

	// list enums
	enums, err := db.ListEnums(ctx)
	if err != nil {
		return nil, fmt.Errorf("list types: %w", err)
	}
	for _, e := range enums {
		s.Enums = append(s.Enums, Enum{
			Name:   e.EnumName,
			Values: strings.Split(string(e.EnumValues), ","),
		})
	}

	// list functions
	funcs, err := db.ListFunctions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list functions: %w", err)
	}
	for _, f := range funcs {
		s.Functions = append(s.Functions, Function{
			Name: f.FunctionName,
			Def:  f.FuncDef,
		})
	}

	seqs, err := db.ListSequences(ctx)
	if err != nil {
		return nil, fmt.Errorf("list sequences: %w", err)
	}
	for _, seq := range seqs {
		if seq.TableName != "" {
			continue
		}
		s.Sequences = append(s.Sequences, Sequence{
			Name:       seq.SequenceName,
			StartValue: seq.StartValue.Int64,
			Increment:  seq.Increment.Int64,
			MinValue:   seq.MinValue.Int64,
			MaxValue:   seq.MaxValue.Int64,
			Cache:      seq.Cache.Int64,
		})
	}

	cols, err := db.ListColumns(ctx)
	if err != nil {
		return nil, fmt.Errorf("list columns: %w", err)
	}

	var tables []string
	for _, c := range cols {
		// these are always sorted by schema, then table
		if len(tables) == 0 || tables[len(tables)-1] != c.TableName {
			tables = append(tables, c.TableName)
		}
	}

	cstr, err := db.ListConstraints(ctx)
	if err != nil {
		return nil, fmt.Errorf("list constraints: %w", err)
	}
	idxs, err := db.ListIndexes(ctx)
	if err != nil {
		return nil, fmt.Errorf("list indexes: %w", err)
	}
	trgs, err := db.ListTriggers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list triggers: %w", err)
	}

	for _, tbl := range tables {
		t := Table{Name: tbl}
		for _, c := range cols {
			if c.TableName != tbl {
				continue
			}
			t.Columns = append(t.Columns, Column{
				Name:         c.ColumnName,
				Type:         c.ColumnType,
				NotNull:      c.NotNull,
				DefaultValue: c.ColumnDefault,
			})
		}

		for _, c := range cstr {
			if c.TableName != tbl {
				continue
			}
			if strings.HasPrefix(c.ConstraintDefinition, "TRIGGER") {
				// skip triggers
				continue
			}

			t.Constraints = append(t.Constraints, Constraint{
				Name: c.ConstraintName,
				Def:  c.ConstraintDefinition,
			})
		}

		for _, idx := range idxs {
			if idx.TableName != tbl {
				continue
			}
			t.Indexes = append(t.Indexes, Index{
				Name: idx.IndexName,
				Def:  idx.IndexDefinition,
			})
		}

		for _, trg := range trgs {
			if trg.TableName != tbl {
				continue
			}
			t.Triggers = append(t.Triggers, Trigger{
				Name: trg.TriggerName,
				Def:  trg.TriggerDefinition,
			})
		}

		for _, seq := range seqs {
			if seq.TableName != tbl {
				continue
			}

			t.Sequences = append(t.Sequences, Sequence{
				Name:       seq.SequenceName,
				StartValue: seq.StartValue.Int64,
				Increment:  seq.Increment.Int64,
				MinValue:   seq.MinValue.Int64,
				MaxValue:   seq.MaxValue.Int64,
				Cache:      seq.Cache.Int64,
				OwnedBy:    seq.TableName + "." + seq.ColumnName,
			})
		}

		s.Tables = append(s.Tables, t)
	}

	return &s, nil
}
