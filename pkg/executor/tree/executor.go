package tree

type treeExecutorImpl struct{}

type TreeExecutor interface {
	Execute() error
}

func NewTreeExectuor() TreeExecutor {
	return treeExecutorImpl{}
}

func (d treeExecutorImpl) Execute() error {
	return nil
}
