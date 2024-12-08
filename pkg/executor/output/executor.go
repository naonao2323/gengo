package output

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/naonao2323/testgen/pkg/common"
	"github.com/naonao2323/testgen/pkg/template"
)

type OutputResult struct{}

type OutputExecutor interface {
	Execute(request common.Request, table string, columns map[string]common.GoDataType, pk []string) (*OutputResult, error)
}

type outputExecutor struct {
	template *template.Template
}

func NewOutputExecutor(template *template.Template) OutputExecutor {
	return outputExecutor{
		template: template,
	}
}

func (t outputExecutor) Execute(request common.Request, table string, columns map[string]common.GoDataType, pk []string) (*OutputResult, error) {
	writer, err := newWriter(table, file)
	if err != nil {
		return nil, err
	}
	switch request {
	case common.DaoPostgresRequest:
		err := t.template.Execute(template.PostgresDao, writer, newData(table, columns, pk))
		if err != nil {
			return &OutputResult{}, err
		}
		return &OutputResult{}, nil
	case common.FrameworkPostgresRequest:
	case common.TestContainerPostgresRequest:
	case common.TestFixturePostgresRequest:
	}
	return &OutputResult{}, nil
}

type writer int

const (
	file writer = iota
	sdout
)

func newWriter(table string, writer writer) (io.Writer, error) {
	switch writer {
	case file:
		file, err := os.Create(fmt.Sprintf("../../test/%v.go", table))
		if err != nil {
			return nil, err
		}
		return file, nil
	case sdout:
		return os.Stdout, nil
	default:
		return nil, errors.New("unknown writer type")
	}
}

func columnsKey(columns map[string]common.GoDataType) []string {
	keys := make([]string, 0, len(columns))
	for key := range columns {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func newData(table string, columns map[string]common.GoDataType, pk []string) template.Data {
	data := make(map[template.Column]template.DataType)

	for clumn, dataType := range columns {
		converted := common.Convert(dataType)
		if converted == "-1" {
			// TODO: error handling
			continue
		}
		data[clumn] = common.Convert(dataType)
	}
	return template.Data{
		TableName: table,
		Pk:        pk,
		DataTypes: data,
		Columns:   columnsKey(columns),
	}
}
