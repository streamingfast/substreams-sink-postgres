package db

import (
	"fmt"
	"sort"
	"strings"

	"go.uber.org/zap"
)

// Insert a row in the DB, it is assumed the table exists, you can do a
// check before with HasTable()
func (l *Loader) Insert(tableName string, primaryKey map[string]string, data map[string]string, reversibleBlockNum *uint64) error {
	uniqueID := createRowUniqueID(primaryKey)

	if l.tracer.Enabled() {
		l.logger.Debug("processing insert operation", zap.String("table_name", tableName), zap.String("primary_key", uniqueID), zap.Int("field_count", len(data)))
	}

	table, found := l.tables[tableName]
	if !found {
		return fmt.Errorf("unknown table %q", tableName)
	}

	entry, found := l.entries.Get(tableName)
	if !found {
		if l.tracer.Enabled() {
			l.logger.Debug("adding tracking of table never seen before", zap.String("table_name", tableName))
		}

		entry = NewOrderedMap[string, *Operation]()
		l.entries.Set(tableName, entry)
	}

	if _, found := entry.Get(uniqueID); found && !l.getDialect().AllowPkDuplicates() {
		return fmt.Errorf("attempting to insert in table %q a primary key %q, that is already scheduled for insertion, insert should only be called once for a given primary key", tableName, primaryKey)
	}

	if l.tracer.Enabled() {
		l.logger.Debug("primary key entry never existed for table, adding insert operation", zap.String("primary_key", uniqueID), zap.String("table_name", tableName))
	}

	// We need to make sure to add the primary key(s) in the data so that those column get created correctly, but only if there is data
	for _, primary := range l.tables[tableName].primaryColumns {
		if dataFromPrimaryKey, ok := primaryKey[primary.name]; ok {
			data[primary.name] = dataFromPrimaryKey
		}
	}

	entry.Set(uniqueID, l.newInsertOperation(table, primaryKey, data, reversibleBlockNum))
	l.entriesCount++
	return nil
}

func createRowUniqueID(m map[string]string) string {
	if len(m) == 1 {
		for _, v := range m {
			return v
		}
	}

	i := 0
	keys := make([]string, len(m))
	for k := range m {
		keys[i] = k
		i++
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	values := make([]string, len(keys))
	for i, key := range keys {
		values[i] = m[key]
	}

	return strings.Join(values, "/")
}

func (l *Loader) GetPrimaryKey(tableName string, pk string) (map[string]string, error) {
	primaryKeyColumns := l.tables[tableName].primaryColumns

	switch len(primaryKeyColumns) {
	case 0:
		return nil, fmt.Errorf("substreams sent a single primary key, but our sql table has none. This is unsupported.")
	case 1:
		return map[string]string{primaryKeyColumns[0].name: pk}, nil
	}

	cols := make([]string, len(primaryKeyColumns))
	for i := range primaryKeyColumns {
		cols[i] = primaryKeyColumns[i].name
	}
	return nil, fmt.Errorf("substreams sent a single primary key, but our sql table has a composite primary key (columns: %s). This is unsupported.", strings.Join(cols, ","))
}

// Update a row in the DB, it is assumed the table exists, you can do a
// check before with HasTable()
func (l *Loader) Update(tableName string, primaryKey map[string]string, data map[string]string, reversibleBlockNum *uint64) error {
	if l.getDialect().OnlyInserts() {
		return fmt.Errorf("update operation is not supported by the current database")
	}

	uniqueID := createRowUniqueID(primaryKey)
	if l.tracer.Enabled() {
		l.logger.Debug("processing update operation", zap.String("table_name", tableName), zap.String("primary_key", uniqueID), zap.Int("field_count", len(data)))
	}

	table, found := l.tables[tableName]
	if !found {
		return fmt.Errorf("unknown table %q", tableName)
	}

	if len(table.primaryColumns) == 0 {
		return fmt.Errorf("trying to perform an UPDATE operation but table %q don't have a primary key(s) set, this is not accepted", tableName)
	}

	entry, found := l.entries.Get(tableName)
	if !found {
		if l.tracer.Enabled() {
			l.logger.Debug("adding tracking of table never seen before", zap.String("table_name", tableName))
		}

		entry = NewOrderedMap[string, *Operation]()
		l.entries.Set(tableName, entry)
	}

	if op, found := entry.Get(uniqueID); found {
		if op.opType == OperationTypeDelete {
			return fmt.Errorf("attempting to update an object with primary key %q, that schedule to be deleted", primaryKey)
		}

		if l.tracer.Enabled() {
			l.logger.Debug("primary key entry already exist for table, merging fields together", zap.String("primary_key", uniqueID), zap.String("table_name", tableName))
		}

		op.mergeData(data)
		entry.Set(uniqueID, op)
		return nil
	} else {
		l.entriesCount++
	}

	if l.tracer.Enabled() {
		l.logger.Debug("primary key entry never existed for table, adding update operation", zap.String("primary_key", uniqueID), zap.String("table_name", tableName))
	}

	entry.Set(uniqueID, l.newUpdateOperation(table, primaryKey, data, reversibleBlockNum))
	return nil
}

// Delete a row in the DB, it is assumed the table exists, you can do a
// check before with HasTable()
func (l *Loader) Delete(tableName string, primaryKey map[string]string, reversibleBlockNum *uint64) error {
	if l.getDialect().OnlyInserts() {
		return fmt.Errorf("delete operation is not supported by the current database")
	}

	uniqueID := createRowUniqueID(primaryKey)
	if l.tracer.Enabled() {
		l.logger.Debug("processing delete operation", zap.String("table_name", tableName), zap.String("primary_key", uniqueID))
	}

	table, found := l.tables[tableName]
	if !found {
		return fmt.Errorf("unknown table %q", tableName)
	}

	if len(table.primaryColumns) != 1 {
		return fmt.Errorf("trying to perform a DELETE operation but table %q don't have a primary key(s) set, this is not accepted", tableName)
	}

	entry, found := l.entries.Get(tableName)
	if !found {
		if l.tracer.Enabled() {
			l.logger.Debug("adding tracking of table never seen before", zap.String("table_name", tableName))
		}

		entry = NewOrderedMap[string, *Operation]()
		l.entries.Set(tableName, entry)
	}

	if _, found := entry.Get(uniqueID); !found {
		if l.tracer.Enabled() {
			l.logger.Debug("primary key entry never existed for table", zap.String("primary_key", uniqueID), zap.String("table_name", tableName))
		}

		l.entriesCount++
	}

	if l.tracer.Enabled() {
		l.logger.Debug("adding deleting operation", zap.String("primary_key", uniqueID), zap.String("table_name", tableName))
	}

	entry.Set(uniqueID, l.newDeleteOperation(table, primaryKey, reversibleBlockNum))
	return nil
}
