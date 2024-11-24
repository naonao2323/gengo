package output

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sync"
	"time"

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

const (
	charSet    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charSetLen = len(charSet)
)

func newData(table string, columns map[string]common.GoDataType, pk []string) template.DaoPostgres {
	data := make(map[template.Column]template.DataType)
	toInsert := make(map[template.Column]template.Value)
	// toDelete := make([]string, len(columns))
	// toUpdate := make(map[template.Column]template.Value, len(columns))
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for clumn, dataType := range columns {
			converted := common.Convert(dataType)
			if converted == "-1" {
				// TODO: error handling
				continue
			}
			data[clumn] = common.Convert(dataType)
		}
	}()
	go func() {
		defer wg.Done()
		rand.Seed(time.Now().UnixNano())
		for column, dataType := range columns {
			switch dataType {
			case common.Int:
				toInsert[column] = fmt.Sprintf("%d", rand.Intn(100))
			case common.Bool:
				toInsert[column] = fmt.Sprintf("%v", rand.Intn(100)%2 == 0)
			case common.String:
				// check制約の考慮
				randText := make([]byte, 10)
				for i := 0; i < len(randText); i++ {
					randText[i] = charSet[rand.Intn(charSetLen)]
				}
				toInsert[column] = string(randText)
			case common.Float64:
				toInsert[column] = fmt.Sprintf("%v", rand.Float64())
			}
		}
	}()
	wg.Wait()
	return template.DaoPostgres{
		TableName: table,
		Pk:        pk,
		Dao:       data,
		ToInsert:  toInsert,
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
