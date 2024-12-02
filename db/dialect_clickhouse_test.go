package db

import (
	"errors"
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
		expectErr error
		valueType reflect.Type
	}{
		{
			name:      "Date",
			value:     "2021-01-01",
			expect:    "2021-01-01",
			expectErr: nil,
			valueType: reflect.TypeOf(time.Time{}),
		}, {
			name:      "Invalid Date",
			value:     "2021-99-01",
			expect:    nil,
			expectErr: errors.New(`could not convert 2021-99-01 to date: parsing time "2021-99-01": month out of range`),
			valueType: reflect.TypeOf(time.Time{}),
		},
		{
			name:      "ISO 8601 datetime",
			value:     "2021-01-01T00:00:00Z",
			expect:    int64(1609459200),
			expectErr: nil,
			valueType: reflect.TypeOf(time.Time{}),
		},
		{
			name:      "common datetime",
			value:     "2021-01-01 00:00:00",
			expect:    int64(1609459200),
			expectErr: nil,
			valueType: reflect.TypeOf(time.Time{}),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := convertToType(test.value, test.valueType)
			if test.expectErr != nil {
				assert.EqualError(t, err, test.expectErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expect, res)
			}
		})
	}
}
