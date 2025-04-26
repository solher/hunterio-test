package pgutil

import (
	"reflect"

	"github.com/jackc/pgx/v5"
)

func ToNamedArgs(s any) pgx.NamedArgs {
	// If it's already a pgx.NamedArgs, return it directly
	if namedArgs, ok := s.(pgx.NamedArgs); ok {
		return namedArgs
	}

	// Initialize empty named args
	namedArgs := pgx.NamedArgs{}

	v := reflect.ValueOf(s)

	// Check if s is a pointer, and if so, dereference it
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		t := v.Type()
		for i := range v.NumField() {
			field := t.Field(i)
			tag := field.Tag.Get("db")
			if tag != "" && tag != "-" {
				namedArgs[tag] = v.Field(i).Interface()
			}
		}
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			key := iter.Key()
			if strKey, ok := key.Interface().(string); ok {
				namedArgs[strKey] = iter.Value().Interface()
			}
		}
	}

	return namedArgs
}
