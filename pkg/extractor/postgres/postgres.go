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

// 外部キーの依存関係を持たせる。
// Fkeyは、自分のキー
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
	placdholder := func(v ...string) string {
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
		`SELECT
			src_col.attname AS source_column,
			tgt_table.relname AS target_table
		FROM
			pg_constraint con
		JOIN
			pg_class src_table ON con.conrelid = src_table.oid
		JOIN
			pg_class tgt_table ON con.confrelid = tgt_table.oid
		JOIN
			pg_attribute src_col ON src_col.attnum = ANY(con.conkey) AND src_col.attrelid = src_table.oid
		JOIN
			pg_attribute tgt_col ON tgt_col.attnum = ANY(con.confkey) AND tgt_col.attrelid = tgt_table.oid
		WHERE
			con.contype = 'f' AND con.conname IN (%s);
		`,
		placdholder(constraints...),
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
			`SELECT
				column_name,
				is_nullable
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
		`SELECT constraint_name
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
