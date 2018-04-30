package utils

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/libgolang/one/model"
)

// RestFilterReduce filter out items out of the slice by using a definition
func RestFilterReduce(def map[string]string, r *http.Request, slicePtr interface{}) {
	p := reflect.ValueOf(slicePtr)
	if p.Kind() != reflect.Ptr {
		panic("Must be a slice pointer")
	}

	//
	filters := RestFilters(def, r)

	// reflect
	origSlice := p.Elem()

	sliceType := origSlice.Type().Elem() // element contained by slice/map
	if origSlice.Kind() == reflect.Map {
		keys := origSlice.MapKeys()
		for _, key := range keys {
			el := origSlice.MapIndex(key)
			if !FilterMatch(el.Interface(), filters) {
				origSlice.SetMapIndex(key, reflect.Value{})
			}
		}
		//origSlice.Set(newSlice)
	} else if origSlice.Kind() == reflect.Slice {
		newSlice := reflect.MakeSlice(reflect.SliceOf(sliceType), 0, 10)
		for i := 0; i < origSlice.Len(); i++ {
			el := origSlice.Index(i)
			if FilterMatch(el.Interface(), filters) {
				newSlice = reflect.Append(newSlice, el)
			}
		}
		origSlice.Set(newSlice)
	} else {
		panic("Invalid Type passed")
	}

}

//
type defMap struct {
	Field string
	Type  string
}

// RestFilters gnerates filters from request
func RestFilters(def map[string]string, r *http.Request) []model.Filter {
	values := r.URL.Query()

	// lowercase keys
	dm := make(map[string]defMap)
	for k, t := range def {
		canonName := strings.ToLower(k)
		dm[canonName] = defMap{k, t}
	}

	//
	result := make([]model.Filter, 0)
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
				result = append(result, &model.FilterString{
					Operation: operation,
					Field:     d.Field,
					Value:     val,
				})
			}
		case "int":
			for _, val := range vals {
				i, _ := strconv.ParseInt(val, 10, 64)
				result = append(result, &model.FilterInt{
					Operation: operation,
					Field:     d.Field,
					Value:     i,
				})
			}
		case "bool":
			for _, val := range vals {
				b, _ := strconv.ParseBool(val)
				result = append(result, &model.FilterBool{
					Operation: operation,
					Field:     d.Field,
					Value:     b,
				})
			}
		}

	}
	return result
}

// FilterMatch takes an array of struct and array of filter and retuns
// the result
func FilterMatch(it interface{}, filters []model.Filter) bool {
	for _, filter := range filters {
		if !filter.Eval(it) {
			return false
		}
	}
	return true
}
