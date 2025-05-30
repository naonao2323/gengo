package output

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/naonao2323/testgen/pkg/common"
	"github.com/naonao2323/testgen/pkg/extractor"
	"github.com/naonao2323/testgen/pkg/template"
)

type OutputResult struct{}

type OutputExecutor interface {
	Execute(request common.Request, table string, columns map[string]common.GoDataType, pk []string) (*OutputResult, error)
}

type outputExecutor struct {
	extractor  extractor.Extractor
	template   *template.Template
	outputPath string
	writer     Writer
}

func NewOutputExecutor(template *template.Template, outputPath string, extractor extractor.Extractor, writer Writer) OutputExecutor {
	return outputExecutor{
		template:   template,
		outputPath: outputPath,
		extractor:  extractor,
		writer:     writer,
	}
}

func (t outputExecutor) Execute(request common.Request, table string, columns map[string]common.GoDataType, pk []string) (*OutputResult, error) {
	writer, err := newWriter(t.outputPath, table, t.writer)
	if err != nil {
		return nil, err
	}
	switch request {
	case common.DaoPostgresRequest:
		err := t.template.Execute(template.PostgresDao, writer, t.newData(table, columns, pk))
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

type Writer int

const (
	File Writer = iota
	Sdout
	Unknown Writer = -1
)

func newWriter(output string, table string, writer Writer) (io.Writer, error) {
	switch writer {
	case File:
		return fileWriter(output, table)
	case Sdout:
		return os.Stdout, nil
	default:
		return nil, errors.New("unknown writer type")
	}
}

func fileWriter(output string, table string) (io.Writer, error) {
	_, err := os.Stat(output)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(output, 0o755)
			if err != nil {
				return nil, err
			}
		}
	}
	file, err := os.Create(fmt.Sprintf("%v/%v.go", output, table))
	if err != nil {
		return nil, err
	}
	return file, nil
}

func columnsKey(columns map[string]common.GoDataType) []string {
	keys := make([]string, 0, len(columns))
	for key := range columns {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (t outputExecutor) newData(table string, columns map[string]common.GoDataType, pk []string) template.Data {
	data := make(map[template.Column]template.DataType)
	for clumn, dataType := range columns {
		converted := common.Convert(dataType)
		if converted == "-1" {
			// TODO: error handling
			continue
		}
		data[clumn] = common.Convert(dataType)
	}
	toSet := func(target []string) map[string]struct{} {
		set := make(map[string]struct{}, len(target))
		for i := range target {
			set[target[i]] = struct{}{}
		}

		return set
	}
	reserved := t.extractor.ListReservedWord()
	return template.Data{
		TableName: table,
		Pk:        pk,
		DataTypes: data,
		Columns:   columnsKey(columns),
		Reserved:  toSet(reserved),
	}
}
