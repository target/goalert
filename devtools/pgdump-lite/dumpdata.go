package pgdump

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/target/goalert/devtools/pgdump-lite/pgd"
)

type TableData struct {
	TableName string
	Columns   []string
	Rows      [][]string
}

// DumpDataWithSchema will return all data from all tables (except those in skipTables) in a structured format.
func DumpDataWithSchema(ctx context.Context, conn pgd.DBTX, out io.Writer, skipTables []string, schema *Schema) error {
	var err error
	if schema == nil {
		schema, err = DumpSchema(ctx, conn)
		if err != nil {
			return fmt.Errorf("dump schema: %w", err)
		}
	}

	for _, t := range schema.Tables {
		if slices.Contains(skipTables, t.Name) {
			continue
		}

		err = dumpTableDataWith(ctx, conn, out, t.Name)
		if err != nil {
			return fmt.Errorf("dump table data: %w", err)
		}
	}

	return nil
}

// DumpDataWithSchema will return all data from all tables (except those in skipTables) in a structured format.
func DumpDataWithSchemaParallel(ctx context.Context, conn *pgxpool.Pool, out io.Writer, skipTables []string, schema *Schema) error {
	var err error
	if schema == nil {
		schema, err = DumpSchema(ctx, conn)
		if err != nil {
			return fmt.Errorf("dump schema: %w", err)
		}
	}

	type streamW struct {
		name string
		pw   *io.PipeWriter
	}

	streams := make(chan io.Reader, len(schema.Tables))
	inputs := make(chan streamW, len(schema.Tables))
	for _, t := range schema.Tables {
		if slices.Contains(skipTables, t.Name) {
			continue
		}
		pr, pw := io.Pipe()
		streams <- pr
		inputs <- streamW{name: t.Name, pw: pw}

		go func() {
			err := conn.AcquireFunc(ctx, func(conn *pgxpool.Conn) error {
				s := <-inputs
				w := bufio.NewWriterSize(s.pw, 65535)
				err := dumpTableData(ctx, conn.Conn(), w, s.name)
				if err != nil {
					return s.pw.CloseWithError(fmt.Errorf("dump table data: %w", err))
				}
				defer s.pw.Close()
				return w.Flush()
			})
			if err != nil {
				panic(err)
			}
		}()
	}
	close(streams)

	for r := range streams {
		_, err = io.Copy(out, r)
		if err != nil {
			return fmt.Errorf("write table data: %w", err)
		}
	}

	return nil
}

func dumpTableDataWith(ctx context.Context, db pgd.DBTX, out io.Writer, tableName string) error {
	switch db := db.(type) {
	case *pgx.Conn:
		return dumpTableData(ctx, db, out, tableName)
	case *pgxpool.Pool:
		return db.AcquireFunc(ctx, func(conn *pgxpool.Conn) error {
			return dumpTableData(ctx, conn.Conn(), out, tableName)
		})
	default:
		return fmt.Errorf("unsupported DBTX type: %T", db)
	}
}

func dumpTableData(ctx context.Context, conn *pgx.Conn, out io.Writer, tableName string) error {
	_, err := fmt.Fprintf(out, "COPY %s FROM stdin WITH (FORMAT csv, HEADER MATCH, ENCODING utf8);\n", pgx.Identifier{tableName}.Sanitize())
	if err != nil {
		return fmt.Errorf("write COPY statement: %w", err)
	}
	_, err = conn.PgConn().CopyTo(ctx, out, fmt.Sprintf("COPY %s TO STDOUT WITH (FORMAT csv, HEADER true, ENCODING utf8)", pgx.Identifier{tableName}.Sanitize()))
	if err != nil {
		return fmt.Errorf("copy data: %w", err)
	}
	_, err = fmt.Fprintf(out, "\\.\n\n")
	if err != nil {
		return fmt.Errorf("write end of COPY: %w", err)
	}

	return nil
}

func DumpData(ctx context.Context, conn pgd.DBTX, out io.Writer, skipTables []string) error {
	schema, err := DumpSchema(ctx, conn)
	if err != nil {
		return err
	}

	return DumpDataWithSchema(ctx, conn, out, skipTables, schema)
}
