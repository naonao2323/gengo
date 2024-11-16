package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

func NewDB(dataSource string) *sql.DB {
	db, err := sql.Open("postgres", dataSource)
	if err != nil {
		fmt.Println(err)
	}
	return db
}

type FKey struct {
	name   string
	isNull bool
}
type FKeyTree struct {
	table      string
	referenced map[FKey]FKeyTree
}

func InitForeignKeyTree(ctx context.Context, db *sql.DB, entrypointTable string) (FKeyTree, error) {
	var tree FKeyTree
	tree.table = entrypointTable
	foreignKeyConstraints, err := getForeignConstraints(ctx, db, entrypointTable)
	if err != nil {
		return tree, err
	}
	refer, err := getReferenced(ctx, db, foreignKeyConstraints)
	if err != nil {
		return tree, err
	}
	tree.referenced = refer
	for k, v := range tree.referenced {
		tree.referenced[k], err = InitForeignKeyTree(ctx, db, v.table)
		if err != nil {
			return tree, err
		}
	}
	return tree, nil
}

func getReferenced(ctx context.Context, db *sql.DB, constraints []string) (map[FKey]FKeyTree, error) {
	if len(constraints) <= 0 {
		return nil, nil
	}
	tree := make(map[FKey]FKeyTree, len(constraints))
	placeholder := func(v ...string) string {
		var b strings.Builder
		for i := range v {
			b.Grow(len(v[i]) + 3)
			b.WriteString("'")
			b.WriteString(v[i])
			b.WriteString("'")
			if i != len(v)-1 {
				b.WriteString(",")
			}
		}
		return b.String()
	}
	isNull := func(v string) bool {
		return v == "YES"
	}
	query := fmt.Sprintf(
		`
		SELECT src_col.attname AS source_column, tgt_table.relname AS target_table
		FROM pg_constraint con
		JOIN pg_class src_table ON con.conrelid = src_table.oid
		JOIN pg_class tgt_table ON con.confrelid = tgt_table.oid
		JOIN pg_attribute src_col ON src_col.attnum = ANY(con.conkey) AND src_col.attrelid = src_table.oid
		JOIN pg_attribute tgt_col ON tgt_col.attnum = ANY(con.confkey) AND tgt_col.attrelid = tgt_table.oid
		WHERE con.contype = 'f' AND con.conname IN (%s);
		`,
		placeholder(constraints...),
	)
	result, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer result.Close()
	type pair struct {
		sourceColumn string
		targetTable  string
	}
	pairs := make([]pair, 0, len(constraints))
	for result.Next() {
		pair := new(pair)
		if err := result.Scan(&pair.sourceColumn, &pair.targetTable); err != nil {
			return nil, err
		}
		pairs = append(pairs, *pair)
	}
	for i := range pairs {
		result := db.QueryRow(
			`
			SELECT column_name, is_nullable
			FROM information_schema.columns
			WHERE column_name = $1;
			`,
			pairs[i].sourceColumn,
		)
		if err := result.Err(); err != nil {
			return nil, err
		}
		var fkey FKey
		var null string
		if err := result.Scan(&fkey.name, &null); err != nil {
			return nil, err
		}
		fkey.isNull = isNull(null)
		tree[fkey] = FKeyTree{
			table: pairs[i].targetTable,
		}
	}
	return tree, nil
}

func getForeignConstraints(ctx context.Context, db *sql.DB, table string) ([]string, error) {
	result, err := db.QueryContext(
		ctx,
		`
		SELECT constraint_name
		FROM information_schema.table_constraints
		WHERE table_name = $1 AND constraint_type = 'FOREIGN KEY'
		`,
		table,
	)
	if err != nil {
		return nil, err
	}
	defer result.Close()
	var constraints []string
	for result.Next() {
		var constraint string
		if err := result.Scan(&constraint); err != nil {
			return nil, err
		}
		constraints = append(constraints, constraint)
	}
	return constraints, nil
}

type table struct {
	name    string
	columns []column
}

type column struct {
	name     string
	isNull   string
	order    int
	dataType string
}

type (
	tableName = string
	Tables    map[tableName]table
)

func InitTables(ctx context.Context, db *sql.DB, schema string) Tables {
	tables := make(Tables)
	tableNames, err := listTableNames(ctx, db, schema)
	if err != nil {
		panic(err)
	}
	for i := range tableNames {
		table, err := fetchTable(ctx, db, tableNames[i])
		if err != nil {
			panic(err)
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
			column_name,
			is_nullable,
			ordinal_position,
			data_type
		FROM information_schema.columns
		WHERE table_name = $1
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
		if err := result.Scan(&column.name, &column.isNull, &column.order, &column.dataType); err != nil {
			return nil, err
		}
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
