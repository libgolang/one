package utils

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rhamerica/one/model"
)

func RestFilters(r *http.Request) map[string][]string {
	values := r.URL.Query()
	for key, vals := range values {
		parts := strings.Split(strings.TrimSpace(key), ".")
		if len(parts) != 2 {
			continue
		}
		if fName == "" || operation == "" {
			continue
		}

		fieldName := strings.TrimSpace(parts[0])
		operation := strings.TrimSpace(parts[1])

		if operation != "eq" && operation != "ne" && operation != "gt" && operation != "ge" && operation != "lt" && operation != "lt" && operation != "like" {
			continue
		}
	}
}
