package internal

import (
	"log"
	"reflect"

	"cloud.google.com/go/spanner"
)

// PrepareStmtAll within sql and args(slice) generate Statement
func PrepareStmtAll(sql string, args ...interface{}) (spanner.Statement, error) {
	names, err := NamedValueParamNames(sql, len(args))
	if err != nil {
		return spanner.Statement{}, err
	}

	stmt := spanner.NewStatement(sql)

	for i := range names {
		stmt.Params[names[i]] = args[i]
	}

	log.Println("[spansqlx]", stmt)

	return stmt, nil
}

// PrepareStmtAny within sql and arg(single) generate Statement
func PrepareStmtAny(sql string, arg interface{}) (spanner.Statement, error) {
	parseReflect := func(elem reflect.Value) map[string]interface{} {
		m := make(map[string]interface{})

		switch elem.Kind() {
		case reflect.Map:
			keys := elem.MapKeys()
			for _, key := range keys {
				k := key.Convert(elem.Type().Key())
				m[k.String()] = elem.MapIndex(k)
			}
		case reflect.Struct:
			t := elem.Type()
			for i := 0; i < t.NumField(); i++ {
				m[t.Field(i).Name] = elem.Field(i).Interface()
			}
		}

		return m
	}

	v := reflect.ValueOf(arg)

	stmt := spanner.NewStatement(sql)

	if v.Kind() == reflect.Ptr {
		stmt.Params = parseReflect(v.Elem())
	} else {
		stmt.Params = parseReflect(v)
	}

	return stmt, nil
}
