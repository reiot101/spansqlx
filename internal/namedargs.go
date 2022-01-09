package internal

import (
	"fmt"
	"regexp"
)

var namedValueParamNameRegex = regexp.MustCompile("@(\\w+)")

// NamedValueParamNames parsing sql query name values
func NamedValueParamNames(sql string, n int) ([]string, error) {
	var names []string

	matches := namedValueParamNameRegex.FindAllStringSubmatch(sql, n)
	if m := len(matches); n != -1 && m < n {
		return nil, fmt.Errorf("scansqlx: query has %d placeholders but %d arguments are provided", m, n)
	}

	for _, m := range matches {
		names = append(names, m[1])
	}

	return names, nil
}
