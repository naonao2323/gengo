package executor

type TemplateResult struct{}

type TableResult struct{}

type TreeResult struct{}

type OutputResult struct{}

type ExecuteStrategy interface {
	TemplateResult | TableResult | TreeResult | OutputResult
}

// Executorではなく、app、serviceが適切
type Executor[A ExecuteStrategy] interface {
	Execute() (A, error)
}
