package sqlutil

import (
	"bufio"
	"bytes"
	"regexp"
	"strings"
)

func sqlSplitBlock(delim []byte, blockIdx int, data []byte, atEOF bool) (advance int, token []byte, err error) {
	nextBlockIdx := bytes.Index(data[blockIdx+2:], delim)
	if nextBlockIdx == -1 {
		if atEOF {
			// return rest as the final query
			return len(data), data, nil
		}

		return 0, nil, nil
	}

	next := blockIdx + 2 + nextBlockIdx + 2

	advance, token, err = sqlSplitQuery(data[next:], atEOF)
	if err != nil {
		return 0, nil, err
	}

	if advance == 0 {
		return 0, nil, nil
	}

	return next + advance, data[:next+len(token)], nil
}

// https://www.postgresql.org/docs/16/sql-syntax-lexical.html#SQL-SYNTAX-DOLLAR-QUOTING
var splitRx = regexp.MustCompile(`\$[^$]*\$`)

// sqlSplitQuery is a bufio.SplitFunc that splits the input SQL query data into smaller queries or blocks based on semicolons and custom block patterns.
//
// More information on bufio.SplitFunc can be found here:
// https://pkg.go.dev/bufio#SplitFunc
func sqlSplitQuery(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// If there's no more data and we're at the end of the file, return no data.
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// Find the index of the first semicolon, which typically delimits SQL statements.
	semiIdx := bytes.IndexRune(data, ';')

	// Attempt to find the start and end indices of a predefined block pattern in the data.
	blockIdx := splitRx.FindIndex(data)

	// Handling for cases with a predefined block pattern and where a semicolon does NOT precede the block pattern.
	if blockIdx != nil && (semiIdx == -1 || semiIdx > blockIdx[0]) {
		blockDelim := data[blockIdx[0]:blockIdx[1]]
		// Process the delimited block separately, considering its specific start position within 'data'.
		return sqlSplitBlock(blockDelim, blockIdx[0], data, atEOF)
	}

	if semiIdx == -1 {
		if atEOF {
			// Returning the rest of the data as the final query when no more delimiters are found and we're at EOF.
			return len(data), data, nil
		}

		return 0, nil, nil // Waiting for more data or a delimiter.
	}

	// Return data up to the first semicolon as a discrete SQL statement.
	return semiIdx + 1, data[:semiIdx], nil
}

// SplitQuery will split a SQL query into individual queries.
//
// It will split on semicolons, but will ignore semicolons inside of $$ blocks.
func SplitQuery(query string) []string {
	s := bufio.NewScanner(strings.NewReader(query))
	s.Split(sqlSplitQuery)

	var queries []string
	for s.Scan() {
		if strings.TrimSpace(s.Text()) == "" {
			continue
		}
		queries = append(queries, s.Text())
	}

	return queries
}
