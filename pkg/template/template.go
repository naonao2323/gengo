package template

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/naonao2323/testgen/pkg/template/postgres"
)

type (
	Column           = string
	Value            = string
	DataType         = string
	DataTypeByColumn = map[Column]DataType
	Data             struct {
		TableName string
		Pk        []Column
		DataTypes DataTypeByColumn
		Columns   []Column
		Reserved  map[string]struct{}
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
	ListLiner        FuncMapKey = FuncMapKey("listLiner")
	Where                       = FuncMapKey("where")
	BackQuote                   = FuncMapKey("backQuote")
	PkType                      = FuncMapKey("pkType")
	Argument                    = FuncMapKey("argument")
	Scan                        = FuncMapKey("scan")
	Insert                      = FuncMapKey("insert")
	Update                      = FuncMapKey("update")
	Delete                      = FuncMapKey("delete")
	Select                      = FuncMapKey("select")
	WithTarget                  = FuncMapKey("withTarget")
	WithPk                      = FuncMapKey("withPk")
	PkLiner                     = FuncMapKey("pkLiner")
	ArgumentPk                  = FuncMapKey("argumentPk")
	IsPrimaryKeyOnly            = FuncMapKey("isPrimaryKeyOnly")
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
		Where: func(pk []string) string {
			if len(pk) == 0 {
				return ""
			}
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
		PkType: func(pk Column, columnsByType map[Column]DataType) string {
			v, ok := columnsByType[pk]
			if !ok {
				panic("unexpected columns")
			}
			return v
		},
		Argument: func(dao DataTypeByColumn) string {
			argument := make([]string, 0)
			for k, v := range dao {
				argument = append(argument, fmt.Sprintf("%v %v", k, v))
			}
			return liner(argument)
		},
		ArgumentPk: func(pk []string, types DataTypeByColumn) string {
			var builder strings.Builder
			for i := range pk {
				builder.WriteString(fmt.Sprintf("%v %v", pk[i], types[pk[i]]))
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
		Insert: func(table string, columns []Column, reserved map[string]struct{}) string {
			var builder strings.Builder
			builder.WriteString(fmt.Sprintf("INSERT INTO %s ", table))
			func() {
				builder.WriteString("(")
				defer builder.WriteString(") ")
				for i := range columns {
					if _, ok := reserved[columns[i]]; ok {
						builder.WriteString(fmt.Sprintf("'%s'", columns[i]))
					} else {
						builder.WriteString(columns[i])
					}
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
		Update: func(table string, columns []Column, pk []string, reserved map[string]struct{}) string {
			if len(columns)-len(pk) <= 0 {
				return ""
			}
			incrementer := func() func() int {
				add := 0
				return func() int {
					add++
					return add
				}
			}()
			elimitedPk := func() []Column {
				set := make(map[string]struct{}, len(pk))
				for i := range pk {
					set[pk[i]] = struct{}{}
				}
				if len(columns)-len(pk) <= 0 {
					// TODO: log
					return nil
				}
				elimited := make([]Column, 0, 2)
				for i := range columns {
					if _, ok := set[columns[i]]; ok {
						continue
					}
					elimited = append(elimited, columns[i])
				}
				return elimited
			}()
			var builder strings.Builder
			builder.WriteString("UPDATE ")
			builder.WriteString(fmt.Sprintf("%s ", table))
			builder.WriteString("SET")
			for i := range elimitedPk {
				if _, ok := reserved[elimitedPk[i]]; ok {
					builder.WriteString(fmt.Sprintf(" '%s' = $%d", elimitedPk[i], incrementer()))
				} else {
					builder.WriteString(fmt.Sprintf(" %s = $%d", elimitedPk[i], incrementer()))
				}
				if i < len(elimitedPk)-1 {
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
			if len(pk) == 0 {
				return ""
			}
			var builder strings.Builder
			builder.WriteString(fmt.Sprintf("DELETE FROM %s WHERE ", table))
			for i := range pk {
				builder.WriteString(fmt.Sprintf("%s = $%v", pk[i], i+1))
				if i < len(pk)-1 {
					builder.WriteString(", ")
				}
			}
			return builder.String()
		},
		Select: func(table string, columns []Column, pk []Column, reserved map[string]struct{}) string {
			if len(columns)-len(pk) <= 0 {
				return ""
			}
			eliminatedPk := func() []Column {
				set := make(map[string]struct{}, len(pk))
				for i := range pk {
					set[pk[i]] = struct{}{}
				}
				eliminated := make([]string, 0, len(columns))
				for i := range columns {
					if _, ok := set[columns[i]]; !ok {
						eliminated = append(eliminated, columns[i])
					}
				}
				return eliminated
			}()
			var builder strings.Builder
			builder.WriteString("SELECT ")
			for i := range eliminatedPk {
				if _, ok := reserved[eliminatedPk[i]]; ok {
					builder.WriteString(fmt.Sprintf("'%s'", eliminatedPk[i]))
				} else {
					builder.WriteString(eliminatedPk[i])
				}
				if i < len(eliminatedPk)-1 {
					builder.WriteString(", ")
				}
			}
			builder.WriteString(fmt.Sprintf(" FROM %s ", table))
			builder.WriteString("WHERE ")
			for i := range pk {
				if _, ok := reserved[pk[i]]; ok {
					builder.WriteString(fmt.Sprintf("'%s' = $%d", pk[i], i+1))
				} else {
					builder.WriteString(fmt.Sprintf("%s = $%d", pk[i], i+1))
				}
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
		WithPk: func(target string, columns []Column, pk []string) string {
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
		PkLiner: func(pk []Column) string {
			var builder strings.Builder
			for i := range pk {
				builder.WriteString(fmt.Sprintf("%v", pk[i]))
				if i < len(pk)-1 {
					builder.WriteString(", ")
				}
			}
			return builder.String()
		},
		IsPrimaryKeyOnly: func(pk []Column, columns DataTypeByColumn) bool {
			cnt := 0
			in := func(column Column) bool {
				for i := range pk {
					if column == pk[i] {
						return true
					}
				}
				return false
			}
			for c := range columns {
				if !in(c) {
					cnt++
				}
			}
			return cnt == len(pk)
		},
	}
}

func (t *Template) Execute(templateType DefaultTemplateType, writer io.Writer, data Data) error {
	if t == nil {
		return nil
	}
	err := t.template.ExecuteTemplate(writer, templateType, data)
	if err != nil {
		return err
	}
	return nil
}
