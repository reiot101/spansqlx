package internal

import (
	"errors"
	"reflect"

	"cloud.google.com/go/spanner"
	"github.com/reiot101/spansqlx/reflectx"
)

// ScanAll scans all rows into a destination, which must be a slice of any
// type. If the destination slice type is a Struct, then Struct will be
// used on each row.  If the destination is some other kind of base type, then
// each row must only have one column which can scan into that type.
func ScanAll(rows []*spanner.Row, dest interface{}) error {
	var vp reflect.Value

	value := reflect.ValueOf(dest)
	if value.Kind() != reflect.Ptr {
		return errors.New("scansqlx: must pass a pointer, not a value, to Struct destination")
	}
	if value.IsNil() {
		return errors.New("scansqlx: nil pointer passed to Struct destination")
	}

	direct := reflect.Indirect(value)

	slice, err := reflectx.BaseType(value.Type(), reflect.Slice)
	if err != nil {
		return err
	}
	isPtr := slice.Elem().Kind() == reflect.Ptr
	base := reflectx.Deref(slice.Elem())

	for _, row := range rows {
		// create a new struct type (which returns PtrTo) and indirect it
		vp = reflect.New(base)

		switch base.Kind() {
		case reflect.Struct:
			// scan into the struct field pointers and append to our results
			err = row.ToStruct(vp.Interface())
		default:
			// scan into the columns field pointers and append to our results
			err = row.Columns(vp.Interface())
		}

		if err != nil {
			return err
		}

		// append
		if isPtr {
			direct.Set(reflect.Append(direct, vp))
		} else {
			direct.Set(reflect.Append(direct, reflect.Indirect(vp)))
		}
	}

	return nil
}

// ScanAny a single Row into the dest map[string]interface{} or struct.
func ScanAny(row *spanner.Row, dest interface{}) error {
	value := reflect.ValueOf(dest)
	if value.Kind() != reflect.Ptr {
		return errors.New("scansqlx: must pass a pointer, not a value, to Struct destination")
	}
	if value.IsNil() {
		return errors.New("scansqlx: nil pointer passed to Struct destination")
	}

	base := reflectx.Deref(value.Type())

	var err error

	switch base.Kind() {
	case reflect.Struct:
		err = row.ToStruct(dest)
	default:
		err = row.Columns(dest)
	}

	if err != nil {
		return err
	}

	return nil
}
