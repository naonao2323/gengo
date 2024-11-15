package extractor

import (
	"context"

	extractor "github.com/naonao2323/testgen/pkg/extractor/postgres"
)

type ExtractGetter interface{}

type Provider int

const (
	Mysql Provider = iota
	Postgres
)

func Extract() ExtractGetter {
	var extract extract
	conn := extractor.NewDB()
	ctx := context.Background()
	tables, err := extractor.FetchTables(ctx, conn, "public")
	if err != nil {
		panic(err)
	}
	for i := range tables {
		table := tables.GetTableName(i)
		rows, err := extractor.GetRows(ctx, conn, table)
		if err != nil {
			panic(err)
		}
	}
	return extract
}

type extract struct {
	table table
}

type table struct {
	columns []column
}

type column struct{}

type DataType int

const (
	INTEGER DataType = iota
	BIGINT
	SMALLINT
	NUMERIC
	DECIMAL
	REAL
	DOUBLE
	DOUBLEPRECISION
	TEXT
	VARCHAR
	CHAR
	DATE
	TIME
	TIMESTAMP
	INTERVAL
	BOOLEAN
	INTEGERARRAY
	TEXTARRAY
	JSON
	JSONB
	UUID
)
