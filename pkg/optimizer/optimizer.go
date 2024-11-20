package optimizer

import (
	"context"

	"github.com/naonao2323/testgen/pkg/executor"
	"github.com/naonao2323/testgen/pkg/executor/common"
	"github.com/naonao2323/testgen/pkg/extractor"
	"github.com/naonao2323/testgen/pkg/state"
)

// 副作用を中心にstate machineを作成して、うまく回せるようにする。
// evnetを生成する責務を持つ。
// tableからeventを作成する。

func NewOptimizer(extractor extractor.Extractor) Optimizer {
	return optimizer{
		extractor,
	}
}

type optimizer struct {
	extractor extractor.Extractor
}

type Optimizer interface {
	Optimize(ctx context.Context, concurrent int, request common.Request) chan state.DaoEvent
}

// dao, testContainer, testFixtureとそれぞれのプロバイダーの組み合わせがあって、どの組みを実行したいのかを良い感じに最適化する
func (o optimizer) Optimize(ctx context.Context, concurrent int, request common.Request) chan state.DaoEvent {
	tables := o.extractor.ListTableNames()
	events := make(chan state.DaoEvent, concurrent)
	go func() {
		o.publish(ctx, tables, request, events)
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
