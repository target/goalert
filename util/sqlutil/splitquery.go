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

var splitRx = regexp.MustCompile(`\$[^$]*\$`)

func sqlSplitQuery(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	semiIdx := bytes.IndexRune(data, ';')

	idx := splitRx.FindIndex(data)
	if idx != nil && (semiIdx == -1 || semiIdx > idx[0]) {
		// have block start and it comes before semi (or no semi)
		return sqlSplitBlock(data[idx[0]:idx[1]], idx[0], data, atEOF)
	}

	// no block, or it comes later
	if semiIdx == -1 {
		if atEOF {
			// return rest as the final query
			return len(data), data, nil
		}

		return 0, nil, nil
	}

	// return up to semi
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
