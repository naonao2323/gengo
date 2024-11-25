package template

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/naonao2323/testgen/pkg/template/postgres"
)

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
		Columns   []Column
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
	ListLiner  FuncMapKey = FuncMapKey("listLiner")
	MapLiner              = FuncMapKey("mapLiner")
	Where                 = FuncMapKey("where")
	BackQuote             = FuncMapKey("backQuote")
	PkType                = FuncMapKey("pkType")
	Argument              = FuncMapKey("argument")
	Scan                  = FuncMapKey("scan")
	Insert                = FuncMapKey("insert")
	Update                = FuncMapKey("update")
	Delete                = FuncMapKey("delete")
	WithTarget            = FuncMapKey("withTarget")
	WithPk                = FuncMapKey("withPk")
	WithTmp               = FuncMapKey("withTmp")
	ArgumentPk            = FuncMapKey("argumentPk")
)

func NewTemplate(optionFuncMap template.FuncMap) (*Template, error) {
	funcMap := newFuncMap()
	for k, v := range optionFuncMap {
		_, ok := funcMap[k]
		if ok {
			// ログ
			continue
		}
		funcMap[k] = v
	}
	tmp, err := template.New(PostgresDao).Funcs(funcMap).Parse(postgres.DaoPostgresTemplate)
	if err != nil {
		return nil, err
	}
	_, err = tmp.New(PostgresTestFixture).Funcs(funcMap).Parse(postgres.DaoPostgresTemplate)
	if err != nil {
		return nil, err
	}
	_, err = tmp.New(PostgresTestContainer).Funcs(funcMap).Parse(postgres.DaoPostgresTemplate)
	if err != nil {
		return nil, err
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
		Where: func(pk []string, dao map[string]string) string {
			where := make([]string, 0)
			for i := range pk {
				where = append(where, fmt.Sprintf("%v = $%d", pk[i], i+1))
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
		PkType: func(pk Column, dao map[Column]DataType) string {
			v, ok := dao[pk]
			if !ok {
				panic("unexpected columns")
			}
			return v
		},
		Argument: func(dao Dao) string {
			argument := make([]string, 0)
			for k, v := range dao {
				argument = append(argument, fmt.Sprintf("%v %v", k, v))
			}
			return liner(argument)
		},
		ArgumentPk: func(pk []string, dao Dao) string {
			var builder strings.Builder
			for i := range pk {
				builder.WriteString(fmt.Sprintf("%v %v", pk[i], dao[pk[i]]))
				if i < len(pk)-1 {
					builder.WriteString(", ")
				}
			}
			return builder.String()
		},
		Scan: func(columns []string, target string) string {
			scan := make([]string, 0, len(columns))
			for i := range columns {
				scan = append(scan, fmt.Sprintf("&%v.%v", target, columns[i]))
			}
			return liner(scan)
		},
		Insert: func(table string, columns []Column) string {
			var builder strings.Builder
			builder.WriteString(fmt.Sprintf("INSERT INTO %s ", table))
			func() {
				builder.WriteString("(")
				defer builder.WriteString(") ")
				for i := range columns {
					builder.WriteString(columns[i])
					if i < len(columns)-1 {
						builder.WriteRune(',')
					}
				}
			}()
			builder.WriteString("VALUES ")
			func() {
				builder.WriteString("(")
				defer builder.WriteString(") ")
				for i := range columns {
					builder.WriteString(fmt.Sprintf("$%d", i+1))
					if i < len(columns)-1 {
						builder.WriteRune(',')
					}
				}
			}()
			return builder.String()
		},
		Update: func(table string, columns []Column, pk []string) string {
			incrementer := func() func() int {
				add := 0
				return func() int {
					add++
					return add
				}
			}()
			var builder strings.Builder
			builder.WriteString("UPDATE ")
			builder.WriteString(fmt.Sprintf("%s ", table))
			builder.WriteString("SET ")
		LOOP:
			for i := range columns {
				for j := range pk {
					if pk[j] == columns[i] {
						continue LOOP
					}
				}
				builder.WriteString(fmt.Sprintf("%s.%s = $%d", table, columns[i], incrementer()))
				if i < len(columns)-1 && i > 0 {
					builder.WriteRune(',')
				}
			}
			builder.WriteString(" WHERE ")
			for i := range pk {
				builder.WriteString(fmt.Sprintf("%s = $%d", pk[i], incrementer()))
				if i < len(pk)-1 {
					builder.WriteString(" AND ")
				}
			}
			return builder.String()
		},
		Delete: func(table string, pk []string) string {
			var builder strings.Builder
			builder.WriteString(fmt.Sprintf("DELETE FROM %s Where ", table))
			for i := range pk {
				builder.WriteString(fmt.Sprintf("%s = $%v", pk[i], i+1))
				if i < len(pk)-1 {
					builder.WriteString(", ")
				}
			}
			return builder.String()
		},
		WithTarget: func(target string, columns []Column) string {
			fields := make([]string, 0, len(columns))
			for i := range columns {
				fields = append(fields, fmt.Sprintf("%s.%s", target, columns[i]))
			}
			var builder strings.Builder
			for i := range fields {
				builder.WriteString(fields[i])
				if i < len(fields)-1 {
					builder.WriteRune(',')
				}
			}
			return builder.String()
		},
		WithPk: func(target string, dao Dao, pk []string) string {
			columns := columns(dao)
			var builder strings.Builder
		LOOP:
			for i := range columns {
				for j := range pk {
					if pk[j] == columns[i] {
						continue LOOP
					}
				}
				builder.WriteString(fmt.Sprintf("%s.%s", target, columns[i]))
				if i < len(columns)-1 {
					builder.WriteString(", ")
				}
				if i == len(columns)-1 {
					builder.WriteString((", "))
				}
			}
			for i := range pk {
				builder.WriteString(pk[i])
				if i < len(pk)-1 {
					builder.WriteString((", "))
				}
			}
			return builder.String()
		},
		WithTmp: func(pk []string) string {
			var builder strings.Builder
			for i := range pk {
				builder.WriteString(fmt.Sprintf("%v", pk[i]))
				if i < len(pk)-1 {
					builder.WriteString(", ")
				}
			}
			return builder.String()
		},
	}
}

func columns(dao Dao) []string {
	columns := make([]string, 0, len(dao))
	for key := range dao {
		columns = append(columns, key)
	}
	return columns
}

func (t *Template) Execute(templateType DefaultTemplateType, writer io.Writer, data DaoPostgres) error {
	if t == nil {
		return nil
	}
	err := t.template.ExecuteTemplate(writer, templateType, data)
	if err != nil {
		return err
	}
	return nil
}
