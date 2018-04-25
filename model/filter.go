package model

import (
	"reflect"
	"strconv"
	"strings"
)

// Filter filter
type Filter interface {
	Eval(it interface{}) bool
}

// FilterString Filter of string types
type FilterString struct {
	Operation string
	Field     string
	Value     string
}

// Eval Implementation of Filter.Eval
func (f *FilterString) Eval(it interface{}) bool {
	v := reflect.ValueOf(it)
	fld := reflect.Indirect(v).FieldByName(f.Field)
	left := fld.String()

	switch f.Operation {
	case "eq":
		return left == f.Value
	case "ne":
		return left != f.Value
	case "gt":
		return strings.Compare(left, f.Value) > 0
	case "ge":
		i := strings.Compare(left, f.Value)
		return i == 0 || i > 0
	case "lt":
		return strings.Compare(left, f.Value) < 0
	case "le":
		i := strings.Compare(left, f.Value)
		return i == 0 || i < 0
	case "like":
		return strings.Contains(left, f.Value)
	}
	return false
}

// FilterInt Filter of string types
type FilterInt struct {
	Operation string
	Field     string
	Value     int64
}

// Eval Implementation of Filter.Eval
func (f *FilterInt) Eval(it interface{}) bool {
	v := reflect.ValueOf(it)
	fld := reflect.Indirect(v).FieldByName(f.Field)
	left := fld.Int()

	switch f.Operation {
	case "eq":
		return left == f.Value
	case "ne":
		return left != f.Value
	case "gt":
		return left > f.Value
	case "ge":
		return left >= f.Value
	case "lt":
		return left < f.Value
	case "le":
		return left <= f.Value
	case "like":
		return strings.Contains(
			strconv.FormatInt(left, 10),
			strconv.FormatInt(f.Value, 10),
		)
	}
	return false
}
