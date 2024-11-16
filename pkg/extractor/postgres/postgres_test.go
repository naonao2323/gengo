package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var URL = "postgres://root:rootpass@localhost:2000/app?sslmode=disable"

func init() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://root:rootpass@localhost:2000/app?sslmode=disable"
	}
	URL = dsn
}

// / user table  memo table			good table		comment table
// / id int      id int				id int			id int
// /             user_id int(FK)	user_id int(FK)	memo_id int(FK)
func migrate(db *sql.DB) error {
	tx, err := db.Begin()
	withRollback := func(qeury string) error {
		_, err = tx.Exec(qeury)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				return err
			}
			return err
		}
		return nil
	}
	if err != nil {
		return err
	}
	if err := withRollback(`CREATE TABLE users (id INT PRIMARY KEY)`); err != nil {
		return err
	}
	if err := withRollback(`CREATE TABLE memos (id INT PRIMARY KEY,user_id INT,FOREIGN KEY (user_id) REFERENCES users(id))`); err != nil {
		return err
	}
	if err := withRollback(`CREATE TABLE goods (id INT PRIMARY KEY,user_id INT,FOREIGN KEY (user_id) REFERENCES users(id))`); err != nil {
		return err
	}
	if err := withRollback(`CREATE TABLE comments (id INT PRIMARY KEY,memo_id INT,FOREIGN KEY (memo_id) REFERENCES memos(id))`); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func drop(db *sql.DB) error {
	tx, err := db.Begin()
	withRollback := func(qeury string) error {
		_, err = tx.Exec(qeury)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				return err
			}
			return err
		}
		return nil
	}
	if err != nil {
		return err
	}
	if err := withRollback(`DROP TABLE comments`); err != nil {
		return err
	}
	if err := withRollback(`DROP TABLE goods`); err != nil {
		return err
	}
	if err := withRollback(`DROP TABLE memos`); err != nil {
		return err
	}
	if err := withRollback(`DROP TABLE users`); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func TestMain(m *testing.M) {
	db := NewDB(URL)
	if err := drop(db); err != nil {
		fmt.Println(err)
	}
	if err := migrate(db); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	m.Run()
	if err := drop(db); err != nil {
		os.Exit(1)
	}
}

func Test_FetchTables(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		schema   string
		expected Tables
	}{
		{
			"no tables belong to schema",
			"none",
			nil,
		},
		{
			"3 tables belong to schema",
			"public",
			Tables{
				table{
					"users",
					"BASE TABLE",
				},
				table{
					"comments",
					"BASE TABLE",
				},
				table{
					"memos",
					"BASE TABLE",
				},
				table{
					"goods",
					"BASE TABLE",
				},
			},
		},
	}

	for _, _test := range tests {
		test := _test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			db := NewDB(URL)
			tables, err := FetchTables(ctx, db, test.schema)
			require.NoError(t, err)
			sort.Slice(test.expected, func(i, j int) bool { return test.expected[i].tableName > test.expected[j].tableName })
			sort.Slice(tables, func(i, j int) bool { return tables[i].tableName > tables[j].tableName })
			assert.Equal(t, test.expected, tables)
		})
	}
}

func Test_GetRows(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		table  string
		expect Rows
	}{
		{
			name:  "get user table",
			table: "users",
			expect: Rows{
				{
					name:       "id",
					table:      "users",
					order:      1,
					isNull:     false,
					dataType:   "integer",
					Referenced: nil,
				},
			},
		},
		{
			name:  "get memo table",
			table: "memos",
			expect: Rows{
				{
					name:       "id",
					table:      "memos",
					order:      1,
					isNull:     false,
					dataType:   "integer",
					Referenced: nil,
				},
				{
					name:     "user_id",
					table:    "memos",
					order:    2,
					isNull:   true,
					dataType: "integer",
					Referenced: []row{
						{
							name:       "id",
							table:      "users",
							order:      1,
							isNull:     false,
							dataType:   "integer",
							Referenced: nil,
						},
					},
				},
			},
		},
		{
			name:  "get comment table",
			table: "comments",
			expect: Rows{
				{
					name:       "id",
					table:      "comments",
					order:      1,
					isNull:     false,
					dataType:   "integer",
					Referenced: nil,
				},
				{
					name:     "memo_id",
					table:    "comments",
					order:    2,
					isNull:   true,
					dataType: "integer",
					Referenced: Rows{
						{
							name:       "id",
							table:      "memos",
							order:      1,
							isNull:     false,
							dataType:   "integer",
							Referenced: nil,
						},
					},
				},
			},
		},
	}

	for _, _test := range tests {
		test := _test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			db := NewDB(URL)
			rows, err := GetRows(ctx, db, test.table, 10)
			require.NoError(t, err)
			fmt.Printf("ssssss%+v\n", rows)
			assert.Equal(t, test.expect, rows)
		})
	}
}
