package output

import (
	"io"

	"github.com/naonao2323/testgen/pkg/executor/common"
	"github.com/naonao2323/testgen/pkg/template"
)

type OutputResult struct{}

type OutputExecutor interface {
	Execute(writer io.Writer, request common.Request, table string, columns map[string]common.GoDataType, pk []string) (OutputResult, error)
}

type outputExecutor struct{}

func (t outputExecutor) Execute(writer io.Writer, request common.Request, table string, columns map[string]common.GoDataType, pk []string) (OutputResult, error) {
	newData := func() template.DaoPostgres {
		data := make(map[template.Column]template.DataType, len(columns))
		for clumn, dataType := range columns {
			data[clumn] = template.Convert(template.GoDataType(dataType))
		}
		return template.DaoPostgres{
			TableName: table,
			Pk:        pk,
			Dao:       data,
		}
	}
	switch request {
	case common.DaoPostgresRequest:
		tmp, err := newTemplate()
		if err != nil {
			return OutputResult{}, err
		}
		err = tmp.Execute(template.PostgresDao, writer, newData())
		if err != nil {
			return OutputResult{}, err
		}
		return OutputResult{}, nil
	case common.FrameworkPostgresRequest:
	case common.TestContainerPostgresRequest:
	case common.TestFixturePostgresRequest:
	}
	return OutputResult{}, nil
}

func NewOutputExecutor() OutputExecutor {
	return outputExecutor{}
}

func newTemplate() (*template.Template, error) {
	template, err := template.NewTemplate(nil)
	if err != nil {
		return nil, err
	}
	return template, nil
}
