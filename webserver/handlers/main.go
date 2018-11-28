package handlers

import (
	"errors"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/thedevsaddam/govalidator"
)

var httpReqs = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "How many HTTP requests processed, partitioned by status code and HTTP method.",
	},
	[]string{"code", "method"},
)

func init() {
	govalidator.AddCustomRule("fields", func(field string, rule string, message string, value interface{}) error {
		if value == nil {
			return nil
		}
		fields := strings.Split(value.(string), ",")
		if len(fields) == 0 {
			return errors.New("Field keyword should be field list divided by comma. Available fields: " + strings.TrimPrefix(rule, "fields:"))
		}
		fieldListStr := strings.TrimPrefix(rule, "fields:")
		fieldListSlice := strings.Split(fieldListStr, ",")
		for _, f := range fields {
			if !contains(fieldListSlice, f) {
				return errors.New("Field keyword should be field list divided by comma. Available fields: " + strings.TrimPrefix(rule, "fields:"))
			}
		}
		return nil
	})
	prometheus.MustRegister(httpReqs)
}
