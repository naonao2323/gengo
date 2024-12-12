package template

import (
	"testing"
)

func TestFuncMapKeyListLiner(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		in       []string
		expected string
	}{
		{
			name:     "list Liner",
			in:       []string{"test", "test2"},
			expected: "test,test2",
		},
		{
			name:     "list Liner when in is empty",
			in:       []string{},
			expected: "",
		},
	}
	funcMapKey := newFuncMap()
	for _, _test := range tests {
		test := _test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			listLiner := funcMapKey[ListLiner].(func(in []string) string)
			actual := listLiner(test.in)
			if actual != test.expected {
				t.Fatalf("does match resp actual: %v, expected: %v", test.expected, actual)
			}
		})
	}
}

func TestFuncMapKeyMapLiner(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		in       map[string]string
		expected string
	}{
		{
			name: "map liner",
			in: map[string]string{
				"test1": "string",
				"test2": "int",
			},
			expected: "test1,test2",
		},
		{
			name: "map liner",
			in: map[string]string{
				"test1": "string",
				"test2": "int",
			},
			expected: "test1,test2",
		},
	}
	funcMap := newFuncMap()
	for _, _test := range tests {
		test := _test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			mapLiner := funcMap[MapLiner].(func(in map[string]string) string)
			actual := mapLiner(test.in)
			if actual != test.expected {
				t.Fatalf("does match resp actual: %v, expected: %v", actual, test.expected)
			}
		})
	}
}

func TestFuncMapKeyWhere(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		pk       []string
		expected string
	}{
		{
			name:     "where",
			pk:       []string{"test1", "test2"},
			expected: "test1 = $1 AND test2 = $2",
		},
		{
			name:     "call where when in is empty",
			pk:       []string{},
			expected: "",
		},
	}
	funcMap := newFuncMap()
	for _, _test := range tests {
		test := _test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			where := funcMap[Where].(func(pk []string) string)
			actual := where(test.pk)
			if actual != test.expected {
				t.Fatalf("does match resp actual: %v, expected: %v", actual, test.expected)
			}
		})
	}
}

func TestFuncMapArgumentPk(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		pk       []string
		types    DataTypeByColumn
		expected string
	}{
		{
			name: "call ArgumentPk",
			pk:   []string{"test1", "test2"},
			types: DataTypeByColumn{
				"test1": "string",
				"test2": "int",
			},
			expected: "test1 string, test2 int",
		},
		{
			name:     "call ArgumentPk when pk is empty",
			pk:       []string{},
			types:    DataTypeByColumn{},
			expected: "",
		},
	}
	funcMap := newFuncMap()
	for _, _test := range tests {
		test := _test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			argumentPk := funcMap[ArgumentPk].(func(pk []string, types DataTypeByColumn) string)
			actual := argumentPk(test.pk, test.types)
			if actual != test.expected {
				t.Fatalf("does match resp actual: %v, expected: %v", actual, test.expected)
			}
		})
	}
}

func TestFuncMapScan(t *testing.T) {
	tests := []struct {
		name     string
		columns  []string
		target   string
		expected string
	}{
		{
			name:     "call Scan",
			columns:  []string{"test1", "test2"},
			target:   "resp",
			expected: "&resp.test1,&resp.test2",
		},
		{
			name:     "call Scan when empty columns",
			columns:  []string{},
			target:   "resp",
			expected: "",
		},
	}

	funcMap := newFuncMap()
	for _, _test := range tests {
		test := _test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			scan := funcMap[Scan].(func(columns []string, target string) string)
			actual := scan(test.columns, test.target)
			if actual != test.expected {
				t.Fatalf("does match resp actual: %v, expected: %v", actual, test.expected)
			}
		})
	}
}

func TestFuncMapInsert(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		table    string
		columns  []Column
		reserved map[string]struct{}
		expected string
	}{
		{
			name:     "when columns is empty",
			table:    "test",
			columns:  []Column{},
			reserved: map[string]struct{}{},
			expected: "INSERT INTO test () VALUES () ",
		},
		{
			name:     "when colums is not empty",
			table:    "test",
			columns:  []Column{"test1", "test2", "test3"},
			reserved: map[string]struct{}{},
			expected: "INSERT INTO test (test1,test2,test3) VALUES ($1,$2,$3) ",
		},
		{
			name:    "when columns is not empty and include reserved",
			table:   "test",
			columns: []Column{"test1", "test2", "test3"},
			reserved: map[string]struct{}{
				"test1": {},
			},
			expected: "INSERT INTO test ('test1',test2,test3) VALUES ($1,$2,$3) ",
		},
	}
	funcMap := newFuncMap()
	for _, _test := range tests {
		test := _test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			insert := funcMap[Insert].(func(table string, columns []Column, reserverd map[string]struct{}) string)
			actual := insert(test.table, test.columns, test.reserved)
			if actual != test.expected {
				t.Fatalf("does match resp actual: %v,expected: %v", actual, test.expected)
			}
		})
	}
}

func TestFuncMapUpdate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		table    string
		columns  []Column
		pk       []string
		reserved map[string]struct{}
		expected string
	}{
		{
			name:     "when columns is empty",
			table:    "test",
			columns:  []Column{},
			pk:       []string{"test1"},
			reserved: map[string]struct{}{},
			expected: "",
		},
		{
			name:     "when columns is not empty",
			table:    "test",
			columns:  []Column{"test1", "test2", "test3", "pk1", "pk2"},
			pk:       []string{"pk1", "pk2"},
			reserved: map[string]struct{}{},
			expected: "UPDATE test SET test1 = $1, test2 = $2, test3 = $3 WHERE pk1 = $4 AND pk2 = $5",
		},
		{
			name:    "when columns is not empty and include reserved",
			table:   "test",
			columns: []Column{"test1", "test2", "pk1", "pk2"},
			pk:      []string{"pk1", "pk2"},
			reserved: map[string]struct{}{
				"test1": {},
			},
			expected: "UPDATE test SET 'test1' = $1, test2 = $2 WHERE pk1 = $3 AND pk2 = $4",
		},
	}
	funcMap := newFuncMap()
	for _, _test := range tests {
		test := _test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			update := funcMap[Update].(func(table string, columns []Column, pk []string, reserved map[string]struct{}) string)
			actual := update(test.table, test.columns, test.pk, test.reserved)
			if actual != test.expected {
				t.Fatalf("does match resp actual: %v, expected: %v", actual, test.expected)
			}
		})
	}
}

func TestFuncMapDelete(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		table    string
		pk       []string
		expected string
	}{
		{
			name:     "call Delete when pk is empty",
			table:    "test",
			pk:       []string{},
			expected: "",
		},
		{
			name:     "call Delete",
			table:    "test",
			pk:       []string{"test1", "test2"},
			expected: "DELETE FROM test WHERE test1 = $1, test2 = $2",
		},
	}
	funcMap := newFuncMap()
	for _, _test := range tests {
		test := _test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			delete := funcMap[Delete].(func(table string, pk []string) string)
			actual := delete(test.table, test.pk)
			if actual != test.expected {
				t.Fatalf("does match resp actual: %v, expected: %v", actual, test.expected)
			}
		})
	}
}

func TestFuncMapSelect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		table    string
		columns  []Column
		pk       []Column
		reserved map[string]struct{}
		expected string
	}{
		{
			name:     "when columns are enmpty",
			table:    "test",
			columns:  []Column{},
			pk:       []Column{},
			reserved: map[string]struct{}{},
			expected: "",
		},
		{
			name:     "when columns is not empty",
			table:    "test",
			columns:  []Column{"test1", "test2", "test3", "pk1", "pk2"},
			pk:       []Column{"pk1", "pk2"},
			reserved: map[string]struct{}{},
			expected: "SELECT test1, test2, test3 FROM test WHERE pk1 = $1, pk2 = $2",
		},
		{
			name:    "when columns is not empty",
			table:   "test",
			columns: []Column{"test1", "test2", "test3", "pk1", "pk2"},
			pk:      []Column{"pk1", "pk2"},
			reserved: map[string]struct{}{
				"test1": {},
			},
			expected: "SELECT 'test1', test2, test3 FROM test WHERE pk1 = $1, pk2 = $2",
		},
	}
	funcMap := newFuncMap()
	for _, _test := range tests {
		test := _test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			do := funcMap[Select].(func(table string, columns []Column, pk []Column, reserved map[string]struct{}) string)
			actual := do(test.table, test.columns, test.pk, test.reserved)
			if actual != test.expected {
				t.Fatalf("does match resp actual: %v, expected: %v", actual, test.expected)
			}
		})
	}
}

func TestFuncMapIsPrimaryKeyOnly(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		pk       []Column
		columns  DataTypeByColumn
		expected bool
	}{
		{
			name: "pk is empty",
			pk:   []Column{},
			columns: DataTypeByColumn{
				"test1": Insert,
				"test2": Insert,
			},
			expected: false,
		},
		{
			name:     "columns is empty",
			pk:       []Column{"test1", "test2"},
			columns:  DataTypeByColumn{},
			expected: false,
		},
		{
			name: "there are non-primary key columns",
			pk:   []Column{"test1", "test2"},
			columns: DataTypeByColumn{
				"test1": Insert,
				"test2": Insert,
				"test3": Insert,
				"test4": Insert,
			},
			expected: true,
		},
		{
			name:     "there are no non-primary key columns",
			pk:       []Column{"test1", "test2"},
			columns:  DataTypeByColumn{},
			expected: false,
		},
	}
	funcMap := newFuncMap()
	for _, _test := range tests {
		test := _test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			only := funcMap[IsPrimaryKeyOnly].(func(pk []Column, columns DataTypeByColumn) bool)
			actual := only(test.pk, test.columns)
			if actual != test.expected {
				t.Fatalf("does match resp actual: %v, expected: %v", actual, test.expected)
			}
		})
	}
}
