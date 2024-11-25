package postgres

// ここprivateにして、呼び出す時に必要なデータやfuncMapがあるかチェックする。
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

func (d {{.TableName }}Dao) Create(db *sql.DB, target {{ .TableName }}) (int64, error) {
	m, err := db.Exec({{ backQuote }}{{ insert $.TableName $.Columns }}{{ backQuote }}, {{- withTarget "target" $.Columns }})
	if err != nil {
		return 0, err
	}
	c, err := m.RowsAffected()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d {{.TableName }}Dao) Update(db *sql.DB, {{ range $pk := .Pk}}{{ $pk }} {{ pkType $pk $.Dao }},{{- end}} target {{ .TableName }}) (int64, error) {
	m, err := db.Exec({{ backQuote }}{{ update $.TableName $.Columns $.Pk }}{{ backQuote }}, {{ withPk "target" $.Dao $.Pk}})
	if err != nil {
		return 0, err
	}
	c, err := m.RowsAffected()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d {{.TableName }}Dao) Delete(db *sql.DB, {{ argumentPk $.Pk $.Dao }}) (int64, error) {
	m, err := db.Exec({{ backQuote }}{{ delete $.TableName $.Pk }}{{ backQuote }}, {{ withTmp $.Pk }})
	if err != nil {
		return 0, err
	}
	c, err := m.RowsAffected()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d {{.TableName }}Dao) Get(db *sql.DB, {{ argumentPk $.Pk $.Dao }}) (*{{.TableName}}, error) {
	m := db.QueryRow("SELECT {{ listLiner $.Columns }} FROM {{.TableName}} WHERE {{ where $.Pk $.Dao }}", {{ listLiner $.Pk }})
	if err := m.Err(); err != nil {
		return nil, err
	}
	var resp {{.TableName}}
	if err := m.Scan({{ scan $.Columns "resp" }}); err != nil {
		return nil, err
	}
	return &resp, nil
}
`
