package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

func NewDB(dataSource string) *sql.DB {
	db, err := sql.Open("postgres", dataSource)
	if err != nil {
		fmt.Println(err)
	}
	return db
}

type (
	table struct {
		tableName string
		tableType string
	}
	Tables []table
	row    struct {
		name       string
		table      string
		order      int
		isNull     bool
		dataType   string
		Referenced []row
	}
	Rows []row
)

func catch() {
	if err := recover(); err != nil {
		fmt.Println("catch  panic", err)
	}
	fmt.Println("ok")
}

func FetchTables(ctx context.Context, db *sql.DB, schema string) (Tables, error) {
	defer catch()
	var tables Tables
	result, err := db.QueryContext(
		ctx,
		`SELECT table_name, table_type
		 FROM information_schema.tables
		 WHERE table_schema = $1`,
		schema,
	)
	if err != nil {
		fmt.Println("query with context error")
		return nil, err
	}
	defer result.Close()
	for result.Next() {
		var table table
		if err := result.Scan(&table.tableName, &table.tableType); err != nil {
			println("table error")
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func GetRows(ctx context.Context, db *sql.DB, table string, timeout time.Duration) (Rows, error) {
	// timeoutを設定できるようにするdefault 10秒
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	defer catch()
	rows, err := getRows(ctx, db, table)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].createTableTree(ctx, db)
	}
	return rows, nil
}

func (r *row) createTableTree(ctx context.Context, db *sql.DB) error {
	constraints, err := getConstraints(ctx, db, r.table, r.name)
	if err != nil {
		return err
	}
	filtered, err := filterConstraints(ctx, db, constraints, r.table)
	if err != nil {
		return err
	}

	rows, err := getReferencedRows(ctx, db, filtered)
	if err != nil {
		return err
	}
	for i := range rows {
		rows[i].createTableTree(ctx, db)
	}
	r.Referenced = rows
	return nil
}

func filterConstraints(ctx context.Context, db *sql.DB, constraints []string, targetTable string) ([]string, error) {
	filtered := make([]string, 0, len(constraints))
	for _, constraint := range constraints {
		result, err := db.QueryContext(
			ctx,
			`
				SELECT table_name, column_name
				FROM information_schema.constraint_column_usage
				WHERE constraint_name = $1
				`,
			constraint,
		)
		if err != nil {
			return nil, err
		}
		var referedTable, referedColumn string
		if err := result.Scan(&referedTable, &referedColumn); err != nil {
			return nil, err
		}
		if targetTable != referedTable {
			filtered = append(filtered, constraint)
		}
		result.Close()
	}
	return filtered, nil
}

func getConstraints(ctx context.Context, db *sql.DB, table, column string) ([]string, error) {
	result, err := db.QueryContext(
		ctx,
		`SELECT
			constraint_name
		 FROM
		 	information_schema.key_column_usage
		 WHERE
		 	table_name = $1
			AND
			column_name = $2
		`,
		table,
		column,
	)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer result.Close()
	var constraints []string
	for result.Next() {
		var constraint string
		if err := result.Scan(&constraint); err != nil {
			fmt.Println(err)
			return nil, err
		}
		constraints = append(constraints, constraint)
	}
	return constraints, nil
}

func getReferencedRows(ctx context.Context, db *sql.DB, constraints []string) ([]row, error) {
	rows := make([]row, 0, len(constraints))
	buildIn := func(constraints []string) string {
		var builder strings.Builder
		for i := range constraints {
			builder.Grow(len(constraints[i]))
			builder.WriteString(constraints[i])
			if i == len(constraints)-1 {
				builder.WriteRune(',')
			}
		}
		return builder.String()
	}
	result, err := db.QueryContext(
		ctx,
		`
			SELECT table_name, column_name
			FROM information_schema.constraint_column_usage
			WHERE constraint_name in ($1)
			`,
		buildIn(constraints),
	)
	if err != nil {
		return nil, err
	}
	result.Close()
	for result.Next() {
		var table, column string
		if err := result.Scan(&table, &column); err != nil {
			return nil, err
		}
		row, err := getRow(ctx, db, table, column)
		if err != nil {
			return nil, err
		}
		if row != nil {
			rows = append(rows, *row)
		}
	}
	return rows, nil
}

func getRow(ctx context.Context, db *sql.DB, table, column string) (*row, error) {
	result := db.QueryRowContext(
		ctx,
		`
		SELECT
			table_name,
			column_name,
			ordinal_position,
			data_type,
			is_nullable
		FROM information_schema.columns
		WHERE table_name = $1 AND column_name = $2;
		`,
		table,
		column,
	)
	if err := result.Err(); err != nil {
		return nil, err
	}
	var row row
	if err := result.Scan(&row.table, &row.name, &row.order, &row.dataType); err != nil {
		return nil, err
	}
	return &row, nil
}

func getRows(ctx context.Context, db *sql.DB, table string) (Rows, error) {
	result, err := db.QueryContext(
		ctx,
		`SELECT
			column_name,
			ordinal_position,
			data_type,
			is_nullable
		FROM information_schema.columns
		WHERE table_name = $1;
		`,
		table,
	)
	if err != nil {
		return nil, err
	}
	var rows Rows
	defer result.Close()
	for result.Next() {
		var row row
		if err := result.Scan(&row.name, &row.order, &row.dataType, &row.isNull); err != nil {
			return nil, err
		}
		row.table = table
		row.Referenced = make(Rows, 10)
		rows = append(rows, row)
	}
	return rows, nil
}
