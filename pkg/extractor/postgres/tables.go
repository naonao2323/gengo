package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type table struct {
	name    string
	columns []column
}

type column struct {
	name     string
	isNull   string
	isPk     bool
	order    int
	dataType PostgresDataType
}

type (
	tableName = string
	Tables    map[tableName]table
)

func (ts Tables) GetPk(table string) []string {
	resp := make([]string, 0, len(ts[table].columns))
	for _, c := range ts[table].columns {
		if c.isPk {
			resp = append(resp, c.name)
		}
	}
	return resp
}

func (ts Tables) GetColumns(table string) []column {
	return ts[table].columns
}

func (ts Tables) GetColumnNames(table string) []string {
	columns := make([]string, 0, len(ts[table].columns))
	for i := range ts[table].columns {
		columns = append(columns, ts[table].columns[i].name)
	}
	return columns
}

func (ts Tables) GetColumnType(table string) (map[string]PostgresDataType, error) {
	t, ok := ts[table]
	if !ok {
		return nil, errors.New("no table")
	}
	dataTypes := make(map[string]PostgresDataType)
	for i := range t.columns {
		dataTypes[t.columns[i].name] = t.columns[i].dataType
	}
	return dataTypes, nil
}

func (ts Tables) ListTableNames() []string {
	names := make([]string, 0, len(ts))
	for k := range ts {
		names = append(names, k)
	}
	return names
}

func InitTables(ctx context.Context, db *sql.DB, schema string) Tables {
	tables := make(Tables)
	tableNames, err := listTableNames(ctx, db, schema)
	if err != nil {
		panic(err)
	}
	for i := range tableNames {
		table, err := fetchTable(ctx, db, tableNames[i])
		if err != nil {
			fmt.Printf("%v", err)
			panic(err.Error)
		}
		if table == nil {
			panic("no table")
		}
		tables[table.name] = *table

	}

	return tables
}

func fetchTable(ctx context.Context, db *sql.DB, name string) (*table, error) {
	result, err := db.QueryContext(
		ctx,
		`
		SELECT
			c.column_name,
			c.is_nullable,
			c.ordinal_position,
			c.data_type,
			CASE
				WHEN kcu.column_name IS NOT NULL THEN 'TRUE'
				ELSE 'FALSE'
			END AS is_pk
		FROM
    		information_schema.columns c
		LEFT JOIN
			information_schema.key_column_usage kcu
			ON c.table_name = kcu.table_name
			AND c.column_name = kcu.column_name
			AND kcu.constraint_name IN (
				SELECT constraint_name
				FROM information_schema.table_constraints
				WHERE table_name = c.table_name
				AND constraint_type = 'PRIMARY KEY'
			)
		WHERE
			c.table_name = $1
		`,
		name,
	)
	if err != nil {
		return nil, err
	}
	defer result.Close()
	columns := make([]column, 0, 10)
	for result.Next() {
		column := new(column)
		dataType := new(string)
		if err := result.Scan(&column.name, &column.isNull, &column.order, &dataType, &column.isPk); err != nil {
			return nil, err
		}
		converted, err := convert(*dataType)
		if err != nil {
			return nil, err
		}
		column.dataType = converted
		columns = append(columns, *column)
	}
	return &table{name, columns}, nil
}

func listTableNames(ctx context.Context, db *sql.DB, schema string) ([]string, error) {
	result, err := db.QueryContext(
		ctx,
		`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = $1
		`,
		schema,
	)
	if err != nil {
		return nil, err
	}
	defer result.Close()
	tables := make([]string, 0, 100)
	for result.Next() {
		table := new(string)
		if err := result.Scan(table); err != nil {
			return nil, err
		}
		if *table == "" {
			return nil, nil
		}
		tables = append(tables, *table)
	}
	return tables, nil
}
