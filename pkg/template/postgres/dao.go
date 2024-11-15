package postgres

const DaoPostgresTemplate = `package dao
import (
	"database/sql"

	_ "github.com/lib/pq"
)

type {{ .StructName }} struct {
	{{- range .Fields }}
	{{ .Name }} {{ .Type }}
	{{- end }}
}

type {{ .TableName }}Dao struct {}

func (d {{.TableName }}Dao) Update(pk int, target {{.StructName}}) (int, error) {
	m, err := db.Exec("UPDATE {{.TableName}} SET {{- range .Fields }}{{ .Name }} = target.{{.NAME}},{{- end }} WHERE {{.PK_FILED}} = pk")
	if err != nil {
		return 0, err
	}
	return m
}

func (d {{.TableName }}Dao) Create() (int, error) {
	m, err := db.Exec("INSERT INTO {{.TableName}} VALUES ({{- range .Fields }}{{ .Name }}{{- end }});")
	if err != nil {
		return 0, err
	}
	return m
}

func (d {{.TableName }}Dao) Delete(pk int) (int, error) {
	m, err := db.Exec("Delete FROM {{.TableName}} Where {{ .PK }} = {{ .PK_FILED }}")
	if err != nil {
		return 0, err
	}
	return m
}

func (d {{.TableName }}Dao) List(){
	m, err := db.QueryRow("")
	if err != nil {
		return nil, err
	}
	return m, nil
}
`
