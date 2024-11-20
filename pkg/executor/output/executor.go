package output

import (
	"io"

	"github.com/naonao2323/testgen/pkg/common"
	"github.com/naonao2323/testgen/pkg/template"
)

type OutputResult struct{}

type OutputExecutor interface {
	Execute(writer io.Writer, request common.Request, table string, columns map[string]common.GoDataType, pk []string) (OutputResult, error)
}

type outputExecutor struct {
	template *template.Template
}

func NewOutputExecutor(template *template.Template) OutputExecutor {
	return outputExecutor{
		template: template,
	}
}

func (t outputExecutor) Execute(writer io.Writer, request common.Request, table string, columns map[string]common.GoDataType, pk []string) (OutputResult, error) {
	newData := func() template.DaoPostgres {
		data := make(map[template.Column]template.DataType, len(columns))
		for clumn, dataType := range columns {
			converted := common.Convert(dataType)
			if converted == "-1" {
				// TODO: error handling
				continue
			}
			data[clumn] = common.Convert(dataType)
		}
		return template.DaoPostgres{
			TableName: table,
			Pk:        pk,
			Dao:       data,
		}
	}
	switch request {
	case common.DaoPostgresRequest:
		err := t.template.Execute(template.PostgresDao, writer, newData())
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
