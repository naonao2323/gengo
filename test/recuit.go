package dao

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type recuit struct {
	company_id string
	id string
}

type recuitDao struct {}

func (d recuitDao) Create(db *sql.DB, target recuit) (int64, error) {
	m, err := db.Exec(`INSERT INTO recuit (id,company_id) VALUES ($1,$2) `,target.id,target.company_id)
	if err != nil {
		return 0, err
	}
	c, err := m.RowsAffected()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d recuitDao) Update(db *sql.DB, id string, target recuit) (int64, error) {
	m, err := db.Exec(`UPDATE recuit SET recuit.company_id = $1 WHERE id = $2`, target.company_id, id)
	if err != nil {
		return 0, err
	}
	c, err := m.RowsAffected()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d recuitDao) Delete(db *sql.DB, id string) (int64, error) {
	m, err := db.Exec(`DELETE FROM recuit Where id = $1`, id)
	if err != nil {
		return 0, err
	}
	c, err := m.RowsAffected()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d recuitDao) Get(db *sql.DB, id string) (*recuit, error) {
	m, err := db.QueryRow("SELECT id,company_id FROM recuit WHERE id = $1")
	if err != nil {
		return nil, err
	}
	if err := m.Err(); err != nil {
		return nil, err
	}
	var resp recuit
	if err := m.Scan(&resp.company_id,&resp.id); err != nil {
		return nil, err
	}
	return &resp, nil
}
