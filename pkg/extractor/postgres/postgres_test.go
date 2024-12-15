package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var URL = "postgres://root:rootpass@localhost:2000/app?sslmode=disable"

func init() {
	dsn := os.Getenv("POSTGRES_URL")
	if dsn == "" {
		dsn = "postgres://root:rootpass@localhost:2000/app?sslmode=disable"
	}
	URL = dsn
}

// / user table  memo table			good table		comment table	blog table
// / id int      id int				id int			id int			id int
// /             user_id int(FK)	user_id int(FK)	memo_id int(FK)
// /			 blog_id int(FK)
func migrate(db *sql.DB) error {
	tx, err := db.Begin()
	withRollback := func(qeury string) error {
		_, err = tx.Exec(qeury)
		if err != nil {
			if rerr := tx.Rollback(); rerr != nil {
				return rerr
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
	if err := withRollback(`CREATE TABLE blogs (id INT PRIMARY KEY)`); err != nil {
		return err
	}
	if err := withRollback(`CREATE TABLE memos (id INT PRIMARY KEY,user_id INT,FOREIGN KEY (user_id) REFERENCES users(id), blog_id INT, FOREIGN KEY (blog_id) REFERENCES blogs(id))`); err != nil {
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
	if err := withRollback(`DROP TABLE blogs`); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func TestMain(m *testing.M) {
	db, err := NewDB(URL)
	if err != nil {
		fmt.Println("connet db error", err)
		os.Exit(1)
	}
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

func TestGetForeignKeyTree(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		table    string
		expected FKeyTree
	}{
		{
			name:  "fetch users tree",
			table: "users",
			expected: FKeyTree{
				table: "users",
			},
		},
		{
			name:  "fetch blogs tree",
			table: "blogs",
			expected: FKeyTree{
				table: "blogs",
			},
		},
		{
			name:  "fetch goods tree",
			table: "goods",
			expected: FKeyTree{
				table: "goods",
				referenced: map[FKey]FKeyTree{
					{
						name:   "user_id",
						isNull: true,
					}: {
						table: "users",
					},
				},
			},
		},
		{
			name:  "fetch memos tree",
			table: "memos",
			expected: FKeyTree{
				table: "memos",
				referenced: map[FKey]FKeyTree{
					{
						name:   "user_id",
						isNull: true,
					}: {
						table: "users",
					},
					{
						name:   "blog_id",
						isNull: true,
					}: {
						table: "blogs",
					},
				},
			},
		},
		{
			name:  "fetch comments tree",
			table: "comments",
			expected: FKeyTree{
				table: "comments",
				referenced: map[FKey]FKeyTree{
					{
						name:   "memo_id",
						isNull: true,
					}: {
						table: "memos",
						referenced: map[FKey]FKeyTree{
							{
								name:   "user_id",
								isNull: true,
							}: {
								table: "users",
							},
							{
								name:   "blog_id",
								isNull: true,
							}: {
								table: "blogs",
							},
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
			db, err := NewDB(URL)
			if err != nil {
				t.Fatal(err)
			}
			result, err := InitForeignKeyTree(ctx, db, test.table)
			require.NoError(t, err)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestInitTables(t *testing.T) {
	tests := []struct {
		name     string
		schema   string
		expected Tables
	}{
		{
			name:     "There are no tables associated with the schema",
			schema:   "test",
			expected: Tables{},
		},
		{
			name:   "There are tables associated with the schema",
			schema: "public",
			expected: Tables{
				"users": table{
					name: "users",
					columns: []column{
						{
							name:     "id",
							isNull:   "NO",
							order:    1,
							dataType: INTEGER,
							isPk:     true,
						},
					},
				},
				"blogs": table{
					name: "blogs",
					columns: []column{
						{
							name:     "id",
							isNull:   "NO",
							order:    1,
							dataType: INTEGER,
							isPk:     true,
						},
					},
				},
				"memos": table{
					name: "memos",
					columns: []column{
						{
							name:     "id",
							isNull:   "NO",
							order:    1,
							dataType: INTEGER,
							isPk:     true,
						},
						{
							name:     "user_id",
							isNull:   "YES",
							order:    2,
							dataType: INTEGER,
							isPk:     false,
						},
						{
							name:     "blog_id",
							isNull:   "YES",
							order:    3,
							dataType: INTEGER,
							isPk:     false,
						},
					},
				},
				"comments": table{
					name: "comments",
					columns: []column{
						{
							name:     "id",
							isNull:   "NO",
							order:    1,
							dataType: INTEGER,
							isPk:     true,
						},
						{
							name:     "memo_id",
							isNull:   "YES",
							order:    2,
							dataType: INTEGER,
							isPk:     false,
						},
					},
				},
				"goods": table{
					name: "goods",
					columns: []column{
						{
							name:     "id",
							isNull:   "NO",
							order:    1,
							dataType: INTEGER,
							isPk:     true,
						},
						{
							name:     "user_id",
							isNull:   "YES",
							order:    2,
							dataType: INTEGER,
							isPk:     false,
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
			db, err := NewDB(URL)
			if err != nil {
				t.Fatal(err)
			}
			tables, err := InitTables(ctx, db, test.schema)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, test.expected, tables)
		})
	}
}

func TestGetPk(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		table  string
		tables Tables
		expect []string
	}{
		{
			name:  "there is no pk",
			table: "users",
			tables: Tables{
				"users": table{
					name: "users",
					columns: []column{
						{
							name: "test",
							isPk: false,
						},
					},
				},
			},
			expect: []string{},
		},
		{
			name:  "there is pk",
			table: "users",
			tables: Tables{
				"users": table{
					name: "users",
					columns: []column{
						{
							name: "test",
							isPk: true,
						},
					},
				},
			},
			expect: []string{"test"},
		},
		{
			name:  "there are pk",
			table: "users",
			tables: Tables{
				"users": table{
					name: "users",
					columns: []column{
						{
							name: "test",
							isPk: true,
						},
						{
							name: "test2",
							isPk: true,
						},
					},
				},
			},
			expect: []string{"test", "test2"},
		},
	}

	for _, _test := range tests {
		test := _test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actual := test.tables.GetPk(test.table)
			assert.Equal(t, test.expect, actual)
			assert.Equal(t, len(test.expect), len(actual))
		})
	}
}
