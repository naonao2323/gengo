package optimizer

import (
	"context"

	"github.com/naonao2323/testgen/pkg/common"
	"github.com/naonao2323/testgen/pkg/executor"
	"github.com/naonao2323/testgen/pkg/extractor"
	"github.com/naonao2323/testgen/pkg/state"
)

func NewOptimizer(extractor extractor.Extractor) Optimizer {
	return optimizer{
		extractor,
	}
}

type optimizer struct {
	extractor extractor.Extractor
}

type Optimizer interface {
	Optimize(ctx context.Context, concurrent int, include []string, request common.Request) chan state.DaoEvent
}

// dao, testContainer, testFixtureとそれぞれのプロバイダーの組み合わせがあって、どの組みを実行したいのかを良い感じに最適化する
func (o optimizer) Optimize(ctx context.Context, concurrent int, include []string, request common.Request) chan state.DaoEvent {
	tables := o.extractor.ListTableNames()
	filtered := filterTables(tables, include)
	events := make(chan state.DaoEvent, len(filtered))
	if len(filtered) <= 0 {
		close(events)
		return events
	}
	go func() {
		o.publish(ctx, filtered, request, events)
	}()
	return events
}

func (o optimizer) publish(ctx context.Context, tables []string, request common.Request, events chan state.DaoEvent) {
	for i := range tables {
		go func() {
			event := state.DaoEvent{
				State:   state.DaoStatePrepare,
				Request: request,
				Props: state.Props{
					StartResult: &executor.StartResult{Table: tables[i]},
				},
			}
			select {
			case <-ctx.Done():
				return
			case events <- event:
				return
			}
		}()
	}
}

func filterTables(tables []string, include []string) []string {
	filtered := make([]string, 0, len(tables))
	if len(include) == 0 {
		return tables
	}
	for i := range tables {
		for j := range include {
			if include[j] == tables[i] {
				filtered = append(filtered, tables[i])
			}
		}
	}
	return filtered
}
