package optimizer

// /ここで良い感じにどこのファイルに、どんな依存関係を整理して、ストラレジーを生成する。
type Stategy = int

const (
	Dao Stategy = iota
	Fixture
	Test
	Framework
)
