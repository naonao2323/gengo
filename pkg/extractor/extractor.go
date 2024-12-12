package extractor

import (
	"context"

	"github.com/naonao2323/testgen/pkg/common"
	"github.com/naonao2323/testgen/pkg/extractor/mysql"
	"github.com/naonao2323/testgen/pkg/extractor/postgres"
)

type Extractor interface {
	GetPk(table string) []string
	GetColumns(table string) map[string]common.GoDataType
	ListTableNames() []string
	ListReservedWord() []string
}

func Extract(ctx context.Context, provider Provider, schema string, source string) Extractor {
	switch provider {
	case Mysql:
		return nil
	case Postgres:
		extract := new(extract[postgres.PostgresDataType])
		db := postgres.NewDB(source)
		extract.tables = postgres.InitTables(ctx, db, schema)
		extract.reserved = postgres.InitReservedWords(ctx, db)
		return extract
	default:
		return nil
	}
}

func (e extract[A]) ListReservedWord() []string {
	return e.reserved.ListReservedWord()
}

func (e extract[A]) ListTableNames() []string {
	return e.tables.ListTableNames()
}

func (e extract[A]) GetPk(table string) []string {
	return e.tables.GetPk(table)
}

func (e extract[A]) GetColumns(table string) map[string]common.GoDataType {
	columnTypes, err := e.tables.GetColumnType(table)
	if err != nil {
		return nil
	}
	columns := e.tables.GetColumnNames(table)
	converted := make(map[string]common.GoDataType, len(columns))
	for i := range columns {
		dataType, ok := columnTypes[columns[i]]
		if !ok {
			return nil
		}
		converted[columns[i]] = convert(dataType)
	}
	return converted
}

type extract[A postgres.PostgresDataType | mysql.MysqlDataType] struct {
	tables TablesGetter[A]
	// tableTree TableTreeGetter
	reserved ReservedGetter[A]
}

type TablesGetter[A postgres.PostgresDataType | mysql.MysqlDataType] interface {
	GetPk(table string) []string
	GetColumnNames(table string) []string
	GetColumnType(table string) (map[string]A, error)
	ListTableNames() []string
}

type ReservedGetter[A postgres.PostgresDataType | mysql.MysqlDataType] interface {
	ListReservedWord() []string
}

type TableTreeGetter interface{}

type Provider int

const (
	Mysql Provider = iota
	Postgres
)

func convert[A postgres.PostgresDataType | mysql.MysqlDataType](dataType A) common.GoDataType {
	switch t := any(dataType).(type) {
	case postgres.PostgresDataType:
		return convertPostgresToGo(t)
	case mysql.MysqlDataType:
		return -1
	default:
		return -1
	}
}

func convertPostgresToGo(postgresType postgres.PostgresDataType) common.GoDataType {
	switch postgresType {
	case postgres.INTEGER, postgres.BIGINT, postgres.SMALLINT:
		return common.Int
	case postgres.NUMERIC, postgres.DECIMAL, postgres.REAL, postgres.DOUBLE, postgres.DOUBLEPRECISION:
		return common.Float64
	case postgres.TEXT, postgres.VARCHAR, postgres.CHAR:
		return common.String
	case postgres.BOOLEAN:
		return common.Bool
	case postgres.DATE, postgres.TIME, postgres.TIMESTAMP, postgres.INTERVAL:
		return common.String
	case postgres.INTEGERARRAY:
		return common.Int
	case postgres.TEXTARRAY:
		return common.String
	case postgres.JSON, postgres.JSONB:
		return common.String
	case postgres.UUID:
		return common.String
	}
	return -1
}
