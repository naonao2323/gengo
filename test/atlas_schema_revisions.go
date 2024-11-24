package dao

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type atlas_schema_revisions struct {
	applied int
	description string
	error string
	error_stmt string
	executed_at string
	execution_time int
	hash string
	operator_version string
	partial_hashes string
	total int
	type int
	version string
}

type atlas_schema_revisionsDao struct {}

func (d atlas_schema_revisionsDao) Create(db *sql.DB, target atlas_schema_revisions) (int64, error) {
	m, err := db.Exec(`INSERT INTO atlas_schema_revisions (error,version,applied,hash,executed_at,operator_version,type,partial_hashes,execution_time,error_stmt,total,description) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) `,target.execution_time,target.error_stmt,target.total,target.description,target.error,target.version,target.applied,target.hash,target.executed_at,target.operator_version,target.type,target.partial_hashes)
	if err != nil {
		return 0, err
	}
	c, err := m.RowsAffected()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d atlas_schema_revisionsDao) Update(db *sql.DB, version string, target atlas_schema_revisions) (int64, error) {
	m, err := db.Exec(`UPDATE atlas_schema_revisions SET atlas_schema_revisions.applied = $1atlas_schema_revisions.hash = $2,atlas_schema_revisions.executed_at = $3,atlas_schema_revisions.operator_version = $4,atlas_schema_revisions.type = $5,atlas_schema_revisions.partial_hashes = $6,atlas_schema_revisions.error = $7,atlas_schema_revisions.error_stmt = $8,atlas_schema_revisions.total = $9,atlas_schema_revisions.description = $10,atlas_schema_revisions.execution_time = $11 WHERE version = $12`, target.total, target.description, target.execution_time, target.error_stmt, target.type, target.partial_hashes, target.error, target.applied, target.hash, target.executed_at, target.operator_version, version)
	if err != nil {
		return 0, err
	}
	c, err := m.RowsAffected()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d atlas_schema_revisionsDao) Delete(db *sql.DB, version string) (int64, error) {
	m, err := db.Exec(`DELETE FROM atlas_schema_revisions Where version = $1`, version)
	if err != nil {
		return 0, err
	}
	c, err := m.RowsAffected()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d atlas_schema_revisionsDao) Get(db *sql.DB, version string) (*atlas_schema_revisions, error) {
	m, err := db.QueryRow("SELECT total,description,execution_time,error_stmt,type,partial_hashes,error,version,applied,hash,executed_at,operator_version FROM atlas_schema_revisions WHERE version = $1")
	if err != nil {
		return nil, err
	}
	if err := m.Err(); err != nil {
		return nil, err
	}
	var resp atlas_schema_revisions
	if err := m.Scan(&resp.operator_version,&resp.type,&resp.partial_hashes,&resp.error,&resp.version,&resp.applied,&resp.hash,&resp.executed_at,&resp.error_stmt,&resp.total,&resp.description,&resp.execution_time); err != nil {
		return nil, err
	}
	return &resp, nil
}
