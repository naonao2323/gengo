package extractor

import (
	"context"

	"github.com/naonao2323/testgen/pkg/extractor/mysql"
	"github.com/naonao2323/testgen/pkg/extractor/postgres"
)

type Extractor interface {
	GetPk(table string) []string
	GetColumns(table string) map[string]GoDataType
}

func (e extract[A]) GetPk(table string) []string {
	return e.tables.GetPk(table)
}

// TODO: test
func (e extract[A]) GetColumns(table string) map[string]GoDataType {
	columnTypes, err := e.tables.GetColumnType(table)
	if err != nil {
		return nil
	}
	columns := e.tables.GetColumnNames(table)
	converted := make(map[string]GoDataType, len(columns))
	for i := range columns {
		dataType, ok := columnTypes[columns[i]]
		if !ok {
			return nil
		}
		converted[columns[i]] = convert(dataType)
	}
	return converted
}

func Extract(ctx context.Context, provider Provider, schema string, source string) Extractor {
	switch provider {
	case Mysql:
		return nil
	case Postgres:
		extract := new(extract[postgres.PostgresDataType])
		db := postgres.NewDB(source)
		extract.tables = postgres.InitTables(ctx, db, schema)
		return extract
	default:
		return nil
	}
}

type extract[A postgres.PostgresDataType | mysql.MysqlDataType] struct {
	tables TablesGetter[A]
	// tableTree TableTreeGetter
}

type TablesGetter[A postgres.PostgresDataType | mysql.MysqlDataType] interface {
	GetPk(table string) []string
	GetColumnNames(table string) []string
	GetColumnType(table string) (map[string]A, error)
}

type TableTreeGetter interface{}

type Provider int

const (
	Mysql Provider = iota
	Postgres
)

type GoDataType int

const (
	Int GoDataType = iota
	Float64
	String
	Bool
)

func convert[A postgres.PostgresDataType | mysql.MysqlDataType](dataType A) GoDataType {
	switch t := any(dataType).(type) {
	case postgres.PostgresDataType:
		return convertPostgresToGo(t)
	case mysql.MysqlDataType:
		return -1
	default:
		return -1
	}
}

func convertPostgresToGo(postgresType postgres.PostgresDataType) GoDataType {
	switch postgresType {
	case postgres.INTEGER, postgres.BIGINT, postgres.SMALLINT:
		return Int
	case postgres.NUMERIC, postgres.DECIMAL, postgres.REAL, postgres.DOUBLE, postgres.DOUBLEPRECISION:
		return Float64
	case postgres.TEXT, postgres.VARCHAR, postgres.CHAR:
		return String
	case postgres.BOOLEAN:
		return Bool
	case postgres.DATE, postgres.TIME, postgres.TIMESTAMP, postgres.INTERVAL:
		return String
	case postgres.INTEGERARRAY:
		return Int
	case postgres.TEXTARRAY:
		return String
	case postgres.JSON, postgres.JSONB:
		return String
	case postgres.UUID:
		return String
	}
	return -1
}
