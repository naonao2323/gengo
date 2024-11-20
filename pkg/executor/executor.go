package executor

type StartResult struct {
	Table string
}

type TreeResult struct{}

type TreeExecutor interface {
	Execute() (TreeResult, error)
}

type treeExecutor struct{}

func (t treeExecutor) Execute() (TreeResult, error) {
	return TreeResult{}, nil
}

func NewTreeExecutor() TreeExecutor {
	return treeExecutor{}
}
