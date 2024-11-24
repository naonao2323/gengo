package dao

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type company struct {
	id string
}

type companyDao struct {}

func (d companyDao) Create(db *sql.DB, target company) (int64, error) {
	m, err := db.Exec(`INSERT INTO company (id) VALUES ($1) `,target.id)
	if err != nil {
		return 0, err
	}
	c, err := m.RowsAffected()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d companyDao) Update(db *sql.DB, id string, target company) (int64, error) {
	m, err := db.Exec(`UPDATE company SET  WHERE id = $1`, id)
	if err != nil {
		return 0, err
	}
	c, err := m.RowsAffected()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d companyDao) Delete(db *sql.DB, id string) (int64, error) {
	m, err := db.Exec(`DELETE FROM company Where id = $1`, id)
	if err != nil {
		return 0, err
	}
	c, err := m.RowsAffected()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d companyDao) Get(db *sql.DB, id string) (*company, error) {
	m, err := db.QueryRow("SELECT id FROM company WHERE id = $1")
	if err != nil {
		return nil, err
	}
	if err := m.Err(); err != nil {
		return nil, err
	}
	var resp company
	if err := m.Scan(&resp.id); err != nil {
		return nil, err
	}
	return &resp, nil
}
