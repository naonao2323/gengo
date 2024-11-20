package postgres

const DaoPostgresTemplate = `package dao

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type {{ .TableName }} struct {
	{{- range $key, $value := .Dao }}
	{{ $key }} {{ $value }}
	{{- end }}
}

type {{ .TableName }}Dao struct {}

func (d {{.TableName }}Dao) Create(db *sql.DB, target {{ .TableName }}) (int, error) {
	m, err := db.Exec({{ backQuote }}INSERT INTO {{ .TableName }} ({{ listLiner .Pk }}) VALUES ({{ mapLiner .ToInsert }}){{ backQuote }})
	if err != nil {
		return 0, err
	}
	return m
}

func (d {{.TableName }}Dao) Update(db *sql.DB, {{ range $pk := .Pk}}{{ $pk }} {{ pkType $pk $.Dao }},{{- end}} target {{ .TableName }}) (int, error) {
	m, err := db.Exec("UPDATE {{.TableName}} SET {{- range $key, $value := .ToInsert }} {{ $key }} = {{ $value }}{{- end}} WHERE {{ where $.Pk $.ToInsert }}")
	if err != nil {
		return 0, err
	}
	return m
}

func (d {{.TableName }}Dao) Delete(db *sql.DB, {{ argument $.Dao }}) (int, error) {
	m, err := db.Exec("DELETE FROM {{.TableName}} WHERE {{ where $.Pk $.ToInsert }}")
	if err != nil {
		return 0, err
	}
	return m
}

func (d {{.TableName }}Dao) Get(db *sql.DB, {{ argument $.Dao }}) ([]{{.TableName}}, error) {
	m, err := db.QueryRow("SELECT {{ mapLiner $.Dao }} FROM {{.TableName}} WHERE {{ where $.Pk $.ToInsert }}")
	if err != nil {
		return nil, err
	}
	if err := m.Err(); err != nil {
		return nil, err
	}
	var resp {{.TableName}}
	if err := m.Scan({{ scan $.Dao $.TableName }}); err != nil {
		return nil, err
	}
	return m, nil
}
`
