package table

import (
	"sync"

	"github.com/naonao2323/testgen/pkg/common"
	"github.com/naonao2323/testgen/pkg/extractor"
)

type TableResult struct {
	Table  string
	Clumns map[string]common.GoDataType
	Pk     []string
}

type TableExecutor interface {
	Execute(table string) (TableResult, error)
}

type tableExecutor struct {
	tableGetter extractor.Extractor
}

func NewTableExecutor(tableGetter extractor.Extractor) TableExecutor {
	return tableExecutor{
		tableGetter: tableGetter,
	}
}

func (t tableExecutor) Execute(table string) (TableResult, error) {
	columns := make(map[string]common.GoDataType)
	pk := make([]string, 0, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		resp := t.tableGetter.GetColumns(table)
		for k, v := range resp {
			columns[k] = v
		}
	}()
	go func() {
		defer wg.Done()
		pk = t.tableGetter.GetPk(table)
	}()
	wg.Wait()
	return TableResult{Clumns: columns, Pk: pk, Table: table}, nil
}
