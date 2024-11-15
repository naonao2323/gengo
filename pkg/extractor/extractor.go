package extractor

import (
	"context"
	"fmt"

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
	result, err := extractor.FetchSchema(ctx, conn, "public")
	if err != nil {
		panic(err)
	}
	var table table
	result.Scan(&table.tableName, &table.tableSchema, &table.tableType)
	fmt.Println(table)
	return extract
}

type extract struct {
	table table
}

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
