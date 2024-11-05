package db

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func Test_convertToType(t *testing.T) {

	tests := []struct {
		name      string
		value     string
		expect    any
		valueType reflect.Type
	}{
		{
			name:      "Date",
			value:     "2021-01-01",
			expect:    "2021-01-01",
			valueType: reflect.TypeOf(time.Time{}),
		},
		{
			name:      "ISO 8601 datetime",
			value:     "2021-01-01T00:00:00Z",
			expect:    int64(1609459200),
			valueType: reflect.TypeOf(time.Time{}),
		},
		{
			name:      "common datetime",
			value:     "2021-01-01 00:00:00",
			expect:    int64(1609459200),
			valueType: reflect.TypeOf(time.Time{}),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := convertToType(test.value, test.valueType)
			assert.NoError(t, err)
			assert.Equal(t, test.expect, res)
		})
	}
}
