package template

import (
	"fmt"
	"html/template"
	"io"
	"strings"

	"github.com/naonao2323/testgen/pkg/template/postgres"
)

// executorとtemplateの二重管理はまずいので、どこに置くか考える。
type GoDataType int

const (
	Int GoDataType = iota
	Float64
	String
	Bool
)

func Convert(dataType GoDataType) string {
	switch dataType {
	case Int:
		return "int"
	case Float64:
		return "float64"
	case String:
		return "string"
	case Bool:
		return "bool"
	default:
		return ""
	}
}

type (
	Column   = string
	Value    = string
	DataType = string
	// Daoじゃなくて抽象的な命名にする。
	Dao         map[Column]DataType
	DaoPostgres struct {
		TableName string
		Pk        []Column
		Dao       Dao
		ToInsert  map[Column]Value
	}
)

type Template struct {
	template *template.Template
	funcMap  template.FuncMap
}

type DefaultTemplateType = string

const (
	PostgresDao           = DefaultTemplateType("PostgresDao")
	PostgresTestFixture   = DefaultTemplateType("PostgresTestFixture")
	PostgresTestContainer = DefaultTemplateType("PostgresTestContainer")
)

type FuncMapKey = string

const (
	ListLiner FuncMapKey = FuncMapKey("ListLiner")
	MapLiner             = FuncMapKey("mapLiner")
	Where                = FuncMapKey("where")
	BackQuote            = FuncMapKey("backQuote")
)

func NewTemplate(optionFuncMap template.FuncMap) (*Template, error) {
	tmp, err := template.New(PostgresDao).Parse(postgres.DaoPostgresTemplate)
	if err != nil {
		return nil, err
	}
	_, err = tmp.New(PostgresTestFixture).Parse(postgres.DaoPostgresTemplate)
	if err != nil {
		return nil, err
	}
	_, err = tmp.New(PostgresTestContainer).Parse(postgres.DaoPostgresTemplate)
	if err != nil {
		return nil, err
	}
	funcMap := newFuncMap()
	for k, v := range optionFuncMap {
		_, ok := funcMap[k]
		if !ok {
			// ログ
			continue
		}
		funcMap[k] = v
	}
	templates := Template{
		template: tmp,
		funcMap:  funcMap,
	}
	return &templates, nil
}

func newFuncMap() template.FuncMap {
	liner := func(in []string) string {
		var builder strings.Builder
		builder.Grow(
			func() int {
				tmp := 0
				for i := range in {
					tmp += len(in[i])
				}
				return tmp
			}() + len(in),
		)
		for i := range in {
			builder.WriteString(in[i])
			if i != len(in)-1 {
				builder.WriteRune(',')
			}
		}
		return builder.String()
	}
	return template.FuncMap{
		ListLiner: func(in []string) string {
			return liner(in)
		},
		MapLiner: func(in map[string]string) string {
			keys := make([]string, 0, len(in))
			for key := range in {
				keys = append(keys, key)
			}

			return liner(keys)
		},
		Where: func(pk []string, toInsert map[string]string) string {
			where := make([]string, 0)
			for i := range pk {
				where = append(where, fmt.Sprintf("%v = %v", pk[i], toInsert[pk[i]]))
			}
			var resp strings.Builder
			for i := range where {
				resp.WriteString(where[i])
				if i != len(where)-1 {
					resp.WriteString(" AND ")
				}
			}
			return resp.String()
		},
		BackQuote: func() string { return "`" },
	}
}

func (t *Template) Execute(templateType DefaultTemplateType, writer io.Writer, data DaoPostgres) error {
	err := t.template.ExecuteTemplate(writer, templateType, data)
	if err != nil {
		return err
	}
	return nil
}
