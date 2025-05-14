package optimizer

import (
	"context"
	"testing"

	"github.com/naonao2323/testgen/pkg/common"
	"github.com/naonao2323/testgen/pkg/extractor"
	"github.com/naonao2323/testgen/pkg/state"
	"github.com/naonao2323/testgen/pkg/util"
)

type fakeExtractor struct {
	getPk      []string
	columns    map[string]common.GoDataType
	tableNames []string
}

func FakeExtractor(getPk []string, columns map[string]common.GoDataType, tableNames []string) extractor.Extractor {
	return fakeExtractor{
		getPk:      getPk,
		columns:    columns,
		tableNames: tableNames,
	}
}

func (f fakeExtractor) GetPk(table string) []string {
	return f.getPk
}

func (f fakeExtractor) GetColumns(table string) map[string]common.GoDataType {
	return f.columns
}

func (f fakeExtractor) ListTableNames() []string {
	return f.tableNames
}

func (f fakeExtractor) ListReservedWord() []string {
	return nil
}

func TestOptimize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		include   *[]string
		req       common.Request
		extractor extractor.Extractor
		tableCnt  int
		isThere   map[string]struct{}
	}{
		{
			name:      "table is empty",
			include:   &[]string{"test", "test2"},
			req:       common.DaoPostgresRequest,
			extractor: FakeExtractor([]string{}, map[string]common.GoDataType{}, []string{}),
			tableCnt:  0,
			isThere:   map[string]struct{}{},
		},
		{
			name:      "include is empty",
			include:   &[]string{},
			req:       common.DaoPostgresRequest,
			extractor: FakeExtractor([]string{}, map[string]common.GoDataType{}, []string{"test", "test2"}),
			tableCnt:  0,
			isThere:   map[string]struct{}{},
		},
		{
			name:      "return events when include is not empty",
			include:   &[]string{"test", "test2"},
			req:       common.DaoPostgresRequest,
			extractor: FakeExtractor([]string{}, map[string]common.GoDataType{}, []string{"test", "test2", "test3"}),
			tableCnt:  2,
			isThere:   map[string]struct{}{"test": {}, "test2": {}},
		},
		{
			name:      "return events when include is empty",
			include:   nil,
			extractor: FakeExtractor([]string{}, map[string]common.GoDataType{}, []string{"test", "test2", "test3"}),
			tableCnt:  3,
			isThere:   map[string]struct{}{"test": {}, "test2": {}, "test3": {}},
		},
	}
	for _, _test := range tests {
		test := _test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ctx, done := util.WithCondition(ctx, test.tableCnt)
			optimizer := NewOptimizer(test.extractor)
			events := optimizer.Optimize(ctx, test.include, test.req)
			actual := make([]state.DaoEvent, 0, test.tableCnt)
			for i := 0; i < test.tableCnt; i++ {
				select {
				case <-ctx.Done():
				case e, k := <-events:
					if !k && test.tableCnt != 0 {
						t.Fatal("events is close")
					}
					select {
					case <-ctx.Done():
						t.Fatal("enpected context done")
					default:
						actual = append(actual, e)
						done()
					}
				}
			}
			wait := func() {
				if test.tableCnt <= 0 {
					return
				}
				<-ctx.Done()
				close(events)
			}
			wait()
			for i := range actual {
				result := *actual[i].Props.StartResult
				table := result.Table
				_, ok := test.isThere[table]
				if !ok {
					t.Fatal("unexpected event")
				}
			}
		})
	}
}
