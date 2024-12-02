package db

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

type TypeGetter func(tableName string, columnName string) (reflect.Type, error)

type Queryable interface {
	query(d dialect) (string, error)
}

type OperationType string

const (
	OperationTypeInsert OperationType = "INSERT"
	OperationTypeUpdate OperationType = "UPDATE"
	OperationTypeDelete OperationType = "DELETE"
)

type Operation struct {
	table              *TableInfo
	opType             OperationType
	primaryKey         map[string]string
	data               map[string]string
	reversibleBlockNum *uint64 // nil if that block is known to be irreversible
}

func (o *Operation) String() string {
	return fmt.Sprintf("%s/%s (%s)", o.table.identifier, createRowUniqueID(o.primaryKey), strings.ToLower(string(o.opType)))
}

func (l *Loader) newInsertOperation(table *TableInfo, primaryKey map[string]string, data map[string]string, reversibleBlockNum *uint64) *Operation {
	return &Operation{
		table:              table,
		opType:             OperationTypeInsert,
		primaryKey:         primaryKey,
		data:               data,
		reversibleBlockNum: reversibleBlockNum,
	}
}

func (l *Loader) newUpdateOperation(table *TableInfo, primaryKey map[string]string, data map[string]string, reversibleBlockNum *uint64) *Operation {
	return &Operation{
		table:              table,
		opType:             OperationTypeUpdate,
		primaryKey:         primaryKey,
		data:               data,
		reversibleBlockNum: reversibleBlockNum,
	}
}

func (l *Loader) newDeleteOperation(table *TableInfo, primaryKey map[string]string, reversibleBlockNum *uint64) *Operation {
	return &Operation{
		table:              table,
		opType:             OperationTypeDelete,
		primaryKey:         primaryKey,
		reversibleBlockNum: reversibleBlockNum,
	}
}

func (o *Operation) mergeData(newData map[string]string) error {
	if o.opType == OperationTypeDelete {
		return fmt.Errorf("unable to merge data for a delete operation")
	}

	for k, v := range newData {
		o.data[k] = v
	}
	return nil
}

var integerRegex = regexp.MustCompile(`^\d+$`)
var dateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
var reflectTypeTime = reflect.TypeOf(time.Time{})

func EscapeIdentifier(valueToEscape string) string {
	if strings.Contains(valueToEscape, `"`) {
		valueToEscape = strings.ReplaceAll(valueToEscape, `"`, `""`)
	}

	return `"` + valueToEscape + `"`
}

func escapeStringValue(valueToEscape string) string {
	if strings.Contains(valueToEscape, `'`) {
		valueToEscape = strings.ReplaceAll(valueToEscape, `'`, `''`)
	}

	return `'` + valueToEscape + `'`
}

// to store in an history table
func primaryKeyToJSON(primaryKey map[string]string) string {
	m, err := json.Marshal(primaryKey)
	if err != nil {
		panic(err) // should never happen with map[string]string
	}
	return string(m)
}

// to store in an history table
func jsonToPrimaryKey(in string) (map[string]string, error) {
	out := make(map[string]string)
	err := json.Unmarshal([]byte(in), &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
