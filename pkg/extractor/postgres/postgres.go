package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

func NewDB(dataSource string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dataSource)
	if err != nil {
		return nil, err
	}
	return db, err
}

type PostgresDataType int

const (
	INTEGER PostgresDataType = iota
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

func convert(dataType string) (PostgresDataType, error) {
	fmt.Printf("%v\n", dataType)
	dataTypeMap := map[string]PostgresDataType{
		"integer":                     INTEGER,
		"bigint":                      BIGINT,
		"smallint":                    SMALLINT,
		"numeric":                     NUMERIC,
		"decimal":                     DECIMAL,
		"real":                        REAL,
		"double":                      DOUBLE,
		"double precision":            DOUBLEPRECISION,
		"text":                        TEXT,
		"varchar":                     VARCHAR,
		"character varying":           VARCHAR,
		"char":                        CHAR,
		"date":                        DATE,
		"time":                        TIME,
		"timestamp":                   TIMESTAMP,
		"interval":                    INTERVAL,
		"boolean":                     BOOLEAN,
		"integer[]":                   INTEGERARRAY,
		"text[]":                      TEXTARRAY,
		"json":                        JSON,
		"jsonb":                       JSONB,
		"uuid":                        UUID,
		"timestamp with time zone":    TIMESTAMP,
		"timestamp without time zone": TIMESTAMP,
	}
	normalize := func() string {
		return strings.TrimSpace(strings.ToLower(dataType))
	}
	normalized := normalize()
	if postgresType, exists := dataTypeMap[normalized]; exists {
		return postgresType, nil
	}
	return -1, fmt.Errorf("unknown Postgres data type: %s", dataType)
}
