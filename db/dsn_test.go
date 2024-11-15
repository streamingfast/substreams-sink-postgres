package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDSN(t *testing.T) {
	tests := []struct {
		name             string
		dns              string
		expectError      bool
		expectConnString string
		expectSchema     string
		expectPassword   string
	}{
		{
			name:             "golden path",
			dns:              "psql://postgres:postgres@localhost/substreams-dev?enable_incremental_sort=off&sslmode=disable",
			expectConnString: "host=localhost port=5432 user=postgres dbname=substreams-dev enable_incremental_sort=off sslmode=disable password=postgres",
			expectSchema:     "public",
			expectPassword:   "postgres",
		},
		{
			name:             "with schema",
			dns:              "psql://postgres:postgres@localhost/substreams-dev?enable_incremental_sort=off&sslmode=disable&schema=foo",
			expectConnString: "host=localhost port=5432 user=postgres dbname=substreams-dev enable_incremental_sort=off  sslmode=disable password=postgres",
			expectSchema:     "foo",
			expectPassword:   "postgres",
		},
		{
			name:             "with password",
			dns:              "clickhouse://default:password@localhost:9000/default",
			expectConnString: "clickhouse://default:password@localhost:9000/default",
			expectSchema:     "default",
			expectPassword:   "password",
		},
		{
			name:             "with blank password",
			dns:              "clickhouse://default:@localhost:9000/default",
			expectConnString: "clickhouse://default:@localhost:9000/default",
			expectSchema:     "default",
			expectPassword:   "",
		},
	}
	for _, test := range tests {
		t.Run(test.dns, func(t *testing.T) {
			d, err := ParseDSN(test.dns)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectConnString, d.ConnString())
				assert.Equal(t, test.expectSchema, d.schema)
				assert.Equal(t, test.expectPassword, d.password)
			}
		})
	}

}
