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
		println(err)
	}
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

// 再帰処理をするところをここにリフトアップ
// これだと最初のtableの外部キーしか考慮できてない
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

// これは絡むひとつしか対応してない
func (r *row) createTableTree(ctx context.Context, db *sql.DB) error {
	constraints, err := getConstraints(ctx, db, r.table, r.name)
	if err != nil {
		return err
	}
	filtered, err := filterConstraints(ctx, db, constraints, r.table)
	if err != nil {
		return err
	}
	if len(filtered) <= 0 {
		return nil
	}
	referenced, err := getReferencedRows(ctx, db, filtered)
	if err != nil {
		return err
	}
	for i := range referenced {
		referenced[i].createTableTree(ctx, db)
	}
	r.Referenced = referenced
	return nil
}

func filterConstraints(ctx context.Context, db *sql.DB, constraints []string, targetTable string) ([]string, error) {
	filtered := make([]string, 0, len(constraints))
	for _, constraint := range constraints {
		result := db.QueryRowContext(
			ctx,
			`
				SELECT table_name, column_name
				FROM information_schema.constraint_column_usage
				WHERE constraint_name = $1
				`,
			constraint,
		)
		var referedTable, referedColumn string
		if err := result.Scan(&referedTable, &referedColumn); err != nil {
			return nil, err
		}
		if targetTable != referedTable {
			filtered = append(filtered, constraint)
		}
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
			builder.Grow(len(constraints[i]) + 2)
			builder.WriteString("'")
			builder.WriteString(constraints[i])
			builder.WriteString("'")
			if i != len(constraints)-1 {
				builder.WriteRune(',')
			}
		}
		return builder.String()
	}
	query := fmt.Sprintf(`
	SELECT table_name, column_name
	FROM information_schema.constraint_column_usage
	WHERE constraint_name IN (%s)`, buildIn(constraints))
	result, err := db.QueryContext(
		ctx,
		query,
	)
	if err != nil {
		return nil, err
	}
	defer result.Close()
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
	var null string
	if err := result.Scan(&row.table, &row.name, &row.order, &row.dataType, &null); err != nil {
		return nil, err
	}
	row.isNull = IsNull(null)
	return &row, nil
}

func IsNull(is string) bool {
	if is == "YES" {
		return true
	}
	return false
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
		var null string
		if err := result.Scan(&row.name, &row.order, &row.dataType, &null); err != nil {
			return nil, err
		}
		row.isNull = IsNull(null)
		row.table = table
		rows = append(rows, row)
	}
	return rows, nil
}
