package extractor

import (
	"errors"
	"testing"

	"github.com/naonao2323/testgen/pkg/common"
	"github.com/naonao2323/testgen/pkg/extractor/mysql"
	"github.com/naonao2323/testgen/pkg/extractor/postgres"
	"github.com/stretchr/testify/assert"
)

type fakeTableGetter[A postgres.PostgresDataType | mysql.MysqlDataType] struct {
	pk          []string
	columnNames []string
	columnType  map[string]A
	err         error
}

func (ft fakeTableGetter[A]) GetPk(table string) []string {
	return ft.pk
}

func (ft fakeTableGetter[A]) GetColumnNames(table string) []string {
	return ft.columnNames
}

func (ft fakeTableGetter[A]) GetColumnType(table string) (map[string]A, error) {
	return ft.columnType, ft.err
}

func (ft fakeTableGetter[A]) ListTableNames() []string {
	return []string{"test"}
}

func Test_Exractor_GetColumn(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		table   string
		extract func() extract[postgres.PostgresDataType]
		expect  map[string]common.GoDataType
	}{
		{
			name:  "fail to fetch get column type",
			table: "users",
			extract: func() extract[postgres.PostgresDataType] {
				return extract[postgres.PostgresDataType]{
					tables: fakeTableGetter[postgres.PostgresDataType]{
						pk:          nil,
						columnNames: nil,
						columnType:  nil,
						err:         errors.New("fail to fetch column type"),
					},
				}
			},
			expect: nil,
		},
		{
			name:  "succeded in converting go data type",
			table: "users",
			extract: func() extract[postgres.PostgresDataType] {
				return extract[postgres.PostgresDataType]{
					tables: fakeTableGetter[postgres.PostgresDataType]{
						pk:          []string{"test", "test2"},
						columnNames: []string{"test", "test2", "test3", "test4"},
						columnType: map[string]postgres.PostgresDataType{
							"test":  postgres.INTEGER,
							"test2": postgres.DOUBLE,
							"test3": postgres.JSON,
							"test4": postgres.BOOLEAN,
						},
						err: nil,
					},
				}
			},
			expect: map[string]common.GoDataType{
				"test":  common.Int,
				"test2": common.Float64,
				"test3": common.String,
				"test4": common.Bool,
			},
		},
	}

	for _, _test := range tests {
		test := _test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actual := test.extract().GetColumns(test.table)
			assert.Equal(t, test.expect, actual)
		})
	}
}
