package extractor

// type ExtractGetter interface{}

// type Provider int

// const (
// 	Mysql Provider = iota
// 	Postgres
// )

// // 再帰的な処理のロジックはこっちに持たせる。
// func Extract(provider Provider, schema string) ExtractGetter {
// 	var extract extract
// 	switch provider {
// 	case Mysql:
// 	case Postgres:
// 		conn := postgres.NewDB()
// 		ctx := context.Background()
// 		extract.tables = createTableTree(ctx, schema, conn, Postgres)
// 	}
// 	return extract
// }

// type extract struct {
// 	tables []table
// }

// type table struct {
// 	columns []column
// }

// type column struct {
// 	name       string
// 	table      string
// 	order      int
// 	isNull     bool
// 	dataType   string
// 	referenced *column
// }

// // 探索できるロジックとかも作る。
// // ツリーを作成する
// func createTableTree[A Provider](ctx context.Context, schema string, conn *sql.DB, provider A) []table {
// 	tables, err := postgres.FetchTables(ctx, conn, schema)
// 	if err != nil {
// 		panic(err)
// 	}
// 	tree := make([]table, 0, len(tables))
// 	for i := range tables {
// 		rows, err := postgres.GetRows(ctx, conn, tables.GetTableName(i))
// 		if err != nil {
// 			panic(err)
// 		}
// 		var table table
// 		table.columns = make([]column, 0, len(rows))
// 		for j, row := range rows {
// 			column := column{
// 				name:       row.GetName(false),
// 				table:      tables.GetTableName(i),
// 				order:      row.GetOrder(false),
// 				isNull:     row.GetIsNull(false),
// 				dataType:   row.GetDataType(false),
// 				referenced: new(column),
// 			}
// 			// モリモリのきがくる
// 			refered, err := postgres.GetReferencedRow(ctx, conn, rows[j])
// 			if err != nil {
// 				panic(err)
// 			}
// 			// refered.Next()で呼び出して木の再現をしたい。
// 			// referedを打ちこむ
// 			// ここで木の探索を行うために、mysqlとpostgresで抽象的な操作をする必要がある。
// 			table.columns = append(table.columns, column)
// 		}
// 		tree = append(tree, table)
// 	}

// 	return tree
// }

// type PostgesRowGetter interface {
// 	Get() *postgres.Referenced
// }

// type MysqlRowGetter interface {
// 	Get() *postgres.Referenced
// }

// type Getter interface {
// 	PostgesRowGetter | MysqlRowGetter
// }

// func RowIterator[A PostgesRowGetter | MysqlRowGetter](refered A) iter.Seq[A] {
// 	return func(yield func(A) bool) {
// 		rows := refered.Get()
// 		yield(rows)
// 	}
// }

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
