package dao

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type memo struct {
	id int
	user_id int
}

type memoDao struct {}

func (d memoDao) Create(db *sql.DB, target memo) (int64, error) {
	m, err := db.Exec(`INSERT INTO memo (user_id,id) VALUES ($1,$2) `,target.user_id,target.id)
	if err != nil {
		return 0, err
	}
	c, err := m.RowsAffected()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d memoDao) Update(db *sql.DB, id int, target memo) (int64, error) {
	m, err := db.Exec(`UPDATE memo SET memo.user_id = $1 WHERE id = $2`, target.user_id, id)
	if err != nil {
		return 0, err
	}
	c, err := m.RowsAffected()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d memoDao) Delete(db *sql.DB, id int) (int64, error) {
	m, err := db.Exec(`DELETE FROM memo Where id = $1`, id)
	if err != nil {
		return 0, err
	}
	c, err := m.RowsAffected()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d memoDao) Get(db *sql.DB, id int) (*memo, error) {
	m, err := db.QueryRow("SELECT id,user_id FROM memo WHERE id = $1")
	if err != nil {
		return nil, err
	}
	if err := m.Err(); err != nil {
		return nil, err
	}
	var resp memo
	if err := m.Scan(&resp.user_id,&resp.id); err != nil {
		return nil, err
	}
	return &resp, nil
}
