package model

import (
	"net/http"
	"net/url"
	"testing"
)

type TestStruct struct {
	Name string
	Age  int
}

func TestRestFilterString(t *testing.T) {
	// given
	query := "name.eq=rick"
	def := map[string]string{
		"Name": "string",
		"Age":  "int",
	}

	//  when
	r := &http.Request{}
	r.URL = &url.URL{RawQuery: query}
	filters := RestFilters(def, r)

	// then
	if len(filters) != 1 {
		t.Error("should return one filter exactly")
	}

	i := filters[0]
	fs, ok := i.(*FilterString)
	if !ok {
		t.Error("should be FilterString")
	}

	if fs.Field != "Name" {
		t.Error("Field should be 'name'")
	}

	if fs.Value != "rick" {
		t.Error("Field should be 'rick'")
	}

}

func TestRestFilterInt(t *testing.T) {
	// given
	query := "age.gt=3&name.eq=fred"
	def := map[string]string{
		"Name": "string",
		"Age":  "int",
	}

	//  when
	r := &http.Request{}
	r.URL = &url.URL{RawQuery: query}
	filters := RestFilters(def, r)

	// then
	if len(filters) != 2 {
		t.Errorf("should be size 2, but it was %d instead", len(filters))
	}

	fs, ok := filters[1].(*FilterString)
	if !ok {
		t.Error("should be FilterString")
	}

	if fs.Field != "Name" {
		t.Error("Field should be 'name'")
	}

	if fs.Value != "fred" {
		t.Error("Value should be 'fred'")
	}

	fi, ok := filters[0].(*FilterInt)
	if !ok {
		t.Error("should be FilterInt")
	}

	if fi.Field != "Age" {
		t.Error("Field should be 'Age'")
	}

	if fi.Value != 3 {
		t.Error("Value should be '3'")
	}
}

func TestFilterReduce(t *testing.T) {
	// given
	query := "age.gt=3&name.eq=Rick"
	def := map[string]string{
		"Name": "string",
		"Age":  "int",
	}

	arr := []TestStruct{
		TestStruct{"Rick", 38},
		TestStruct{"Pau", 36},
		TestStruct{"Vale", 6},
		TestStruct{"Sofi", 4},
	}

	//  when
	r := &http.Request{}
	r.URL = &url.URL{RawQuery: query}
	filters := RestFilters(def, r)

	result := make([]TestStruct, 0)
	for _, t := range arr {
		if FilterMatch(t, filters) {
			result = append(result, t)
		}
	}

	// then
	if len(result) != 1 {
		t.Error("Should return only one")
	}
}
