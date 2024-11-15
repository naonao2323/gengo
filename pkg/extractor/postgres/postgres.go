package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func NewDB() *sql.DB {
	db, err := sql.Open("postgres", "postgres://root:rootpass@localhost:5432/app?sslmode=disable")
	if err != nil {
		fmt.Println(err)
	}
	return db
}

type ProviderGetter interface {
	FetchTables()
	GetTableName()
	GetTableTypeName()
}
type table struct {
	tableName string
	tableType string
}

type Tables []table

func (t Tables) GetTableName(index int) string {
	return t[index].tableName
}

func (t Tables) GetTableTypeName(index int) string {
	return t[index].tableType
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

// ここcolumnです。
// referencedは複数にならない。
type row struct {
	name       string
	table      string
	order      int
	isNull     bool
	dataType   string
	Referenced *row
}

func (r *row) GetName(refered bool) string {
	if refered {
		return r.Referenced.name
	}
	return r.name
}

func (r *row) GetTable(refered bool) string {
	if refered {
		return r.Referenced.table
	}
	return r.table
}

func (r *row) GetOrder(refered bool) int {
	if refered {
		return r.Referenced.order
	}
	return r.order
}

func (r *row) GetIsNull(refered bool) bool {
	if refered {
		return r.Referenced.isNull
	}
	return r.isNull
}

func (r *row) GetDataType(refered bool) string {
	if refered {
		return r.Referenced.dataType
	}
	return r.dataType
}

type Rows []row

func (r Rows) GetName(index int) string {
	return r[index].name
}

func (r Rows) GetOrder(index int) int {
	return r[index].order
}

func (r Rows) GetIsNull(index int) bool {
	return r[index].isNull
}

func (r Rows) GetDataType(index int) string {
	return r[index].dataType
}

// refenced nameを返す。
func (r Rows) GetReferencedRowName(index int) string {
	return r[index].Referenced.name
}

func (r Rows) GetReferencedRow(index int) *row {
	return r[index].Referenced
}

func (r Rows) GetTableName(index int) string {
	return r[index].table
}

func GetRows(ctx context.Context, db *sql.DB, table string) (Rows, error) {
	// timeoutを設定できるようにするdefault 10秒
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	defer catch()
	rows, err := getRows(ctx, db, table)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// 浅い取得
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
		var resp row
		if err := result.Scan(&resp.name, &resp.order, &resp.dataType, &resp.isNull); err != nil {
			return nil, err
		}
		resp.table = table
		resp.Referenced = new(row)
		rows = append(rows, resp)
	}
	return rows, nil
}

type Referenced = row

func (r Referenced) Get() *Referenced {
	// ここが配列になる。
	return r.Referenced
}

func GetReferencedRow(ctx context.Context, db *sql.DB, row row) (Referenced, error) {
	// 再帰的に処理をして、外部キーを取得する。
	var state getReferenceRow
	state = row.getReferencedRow
	for {
		state = state(ctx, db)
		if state != nil {
			break
		}
	}
	return row, nil
}

func (r *row) getReferencedRow(ctx context.Context, db *sql.DB) getReferenceRow {
	// 関連している外部キーを取得する。
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
		r.table,
		r.name,
	)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer result.Close()
	var constraints []string
	for result.Next() {
		var constraint string
		if err := result.Scan(&constraint); err != nil {
			fmt.Println(err)
			return nil
		}
		constraints = append(constraints, constraint)
	}

	filtered, err := filterConstrainsts(ctx, db, constraints, r.table)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if len(filtered) != 1 {
		return nil
	}
	// ここが複数になる。
	refered, err := findReferenced(ctx, db, filtered[0])
	if err != nil {
		fmt.Println(err)
		return nil
	}
	r.Referenced = refered
	return refered.getReferencedRow
}

func findReferenced(ctx context.Context, db *sql.DB, constraint string) (*row, error) {
	var row row
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
	result.Close()
	var refencedTable, refencedColumn string
	if err := result.Scan(&refencedTable, &refencedColumn); err != nil {
		return nil, err
	}

	result, err = db.QueryContext(
		ctx,
		`
		SELECT
			column_name,
			ordinal_position,
			data_type,
			is_nullable
		FROM information_schema.columns
		WHERE table_name = $1 AND column_name = $2;
		`,
		refencedTable,
		refencedColumn,
	)
	if err != nil {
		return nil, err
	}
	result.Close()
	if err := result.Scan(&row.name, &row.order, &row.dataType, &row.isNull); err != nil {
		return nil, err
	}
	row.table = refencedTable
	return &row, nil
}

// primary keyとunique keyをフィルタリングしている。
func filterConstrainsts(ctx context.Context, db *sql.DB, constraints []string, table string) ([]string, error) {
	filtered := make([]string, len(constraints))
	copy(filtered, constraints)
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
		if table != referedTable {
			filtered = append(filtered, constraint)
		}
		result.Close()
	}
	return filtered, nil
}

type getReferenceRow func(ctx context.Context, db *sql.DB) getReferenceRow

func catch() {
	if err := recover(); err != nil {
		fmt.Println("catch  panic", err)
	}
	fmt.Println("ok")
}
