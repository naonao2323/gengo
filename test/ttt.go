package dao

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type ttt struct {
}

type tttDao struct {}

func (d tttDao) Create(db *sql.DB, target ttt) (int64, error) {
	m, err := db.Exec(`INSERT INTO ttt () VALUES () `,)
	if err != nil {
		return 0, err
	}
	c, err := m.RowsAffected()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d tttDao) Update(db *sql.DB,  target ttt) (int64, error) {
	m, err := db.Exec(`UPDATE ttt SET  WHERE `, )
	if err != nil {
		return 0, err
	}
	c, err := m.RowsAffected()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d tttDao) Delete(db *sql.DB, ) (int64, error) {
	m, err := db.Exec(`DELETE FROM ttt Where `, )
	if err != nil {
		return 0, err
	}
	c, err := m.RowsAffected()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d tttDao) Get(db *sql.DB, ) (*ttt, error) {
	m, err := db.QueryRow("SELECT  FROM ttt WHERE ")
	if err != nil {
		return nil, err
	}
	if err := m.Err(); err != nil {
		return nil, err
	}
	var resp ttt
	if err := m.Scan(); err != nil {
		return nil, err
	}
	return &resp, nil
}
