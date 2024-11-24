package dao

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type users struct {
	id int
}

type usersDao struct {}

func (d usersDao) Create(db *sql.DB, target users) (int64, error) {
	m, err := db.Exec(`INSERT INTO users (id) VALUES ($1) `,target.id)
	if err != nil {
		return 0, err
	}
	c, err := m.RowsAffected()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d usersDao) Update(db *sql.DB, id int, target users) (int64, error) {
	m, err := db.Exec(`UPDATE users SET  WHERE id = $1`, id)
	if err != nil {
		return 0, err
	}
	c, err := m.RowsAffected()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d usersDao) Delete(db *sql.DB, id int) (int64, error) {
	m, err := db.Exec(`DELETE FROM users Where id = $1`, id)
	if err != nil {
		return 0, err
	}
	c, err := m.RowsAffected()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d usersDao) Get(db *sql.DB, id int) (*users, error) {
	m, err := db.QueryRow("SELECT id FROM users WHERE id = $1")
	if err != nil {
		return nil, err
	}
	if err := m.Err(); err != nil {
		return nil, err
	}
	var resp users
	if err := m.Scan(&resp.id); err != nil {
		return nil, err
	}
	return &resp, nil
}
