package template

import (
	"testing"
)

func TestNewFuncMap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
	}{
		{
			name: "debug",
		},
	}

	for _, _test := range tests {
		test := _test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// template, err := NewTemplate()
			// require.NoError(t, err)
			// template.Execute(PostgresDao, os.Stdout)
		})
	}
}
