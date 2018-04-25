package model

import (
	"net/http"
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
	operation string
	Field     string
	Value     string
}

// Eval Implementation of Filter.Eval
func (f *FilterString) Eval(it interface{}) bool {
	v := reflect.ValueOf(it)
	fld := reflect.Indirect(v).FieldByName(f.Field)
	left := fld.String()

	switch f.operation {
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
	operation string
	Field     string
	Value     int64
}

// Eval Implementation of Filter.Eval
func (f *FilterInt) Eval(it interface{}) bool {
	v := reflect.ValueOf(it)
	fld := reflect.Indirect(v).FieldByName(f.Field)
	left := fld.Int()

	switch f.operation {
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

// FilterMatch takes an array of struct and array of filter and retuns
// the result
func FilterMatch(it interface{}, filters []Filter) bool {
	for _, filter := range filters {
		if !filter.Eval(it) {
			return false
		}
	}
	return true
}

//
type defMap struct {
	Field string
	Type  string
}

// RestFilters gnerates filters from request
func RestFilters(def map[string]string, r *http.Request) []Filter {
	values := r.URL.Query()

	// lowercase keys
	dm := make(map[string]defMap)
	for k, t := range def {
		canonName := strings.ToLower(k)
		dm[canonName] = defMap{k, t}
	}

	//
	result := make([]Filter, 0)
	for key, vals := range values {
		parts := strings.Split(strings.TrimSpace(key), ".")
		n := len(parts)
		var fieldName string
		var operation string
		if n == 1 {
			fieldName = parts[0]
			operation = "eq"
		} else if n == 2 {
			fieldName = strings.ToLower(strings.TrimSpace(parts[0]))
			operation = strings.TrimSpace(parts[1])
		} else {
			continue
		}

		if fieldName == "" || operation == "" {
			continue
		}

		if operation != "eq" && operation != "ne" && operation != "gt" && operation != "ge" && operation != "lt" && operation != "le" && operation != "like" {
			continue
		}

		d, ok := dm[fieldName]
		if !ok {
			continue
		}

		switch d.Type {
		case "string":
			for _, val := range vals {
				result = append(result, &FilterString{operation, d.Field, val})
			}
		case "int":
			for _, val := range vals {
				i, _ := strconv.ParseInt(val, 10, 64)
				result = append(result, &FilterInt{operation, d.Field, i})
			}
		}

	}
	return result
}
