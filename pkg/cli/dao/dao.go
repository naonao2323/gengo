package dao

import (
	"context"
	"sync"

	"github.com/naonao2323/testgen/pkg/executor"
	"github.com/naonao2323/testgen/pkg/executor/common"
	"github.com/naonao2323/testgen/pkg/executor/output"
	"github.com/naonao2323/testgen/pkg/executor/table"
	"github.com/naonao2323/testgen/pkg/extractor"
	"github.com/naonao2323/testgen/pkg/optimizer"
	"github.com/naonao2323/testgen/pkg/state"
	"github.com/naonao2323/testgen/pkg/template"
	"github.com/naonao2323/testgen/pkg/util"
	"github.com/spf13/cobra"
)

type dao struct {
	extractor extractor.Extractor
	optimizer optimizer.Optimizer
}

func NewCommand() *cobra.Command {
	ctx := context.Background()
	schema := "public"
	source := "postgres://root:rootpass@localhost:5432/app?sslmode=disable"
	extractor := extractor.Extract(ctx, extractor.Postgres, schema, source)
	optimizer := optimizer.NewOptimizer(extractor)
	g := dao{
		extractor: extractor,
		optimizer: optimizer,
	}
	cmd := cobra.Command{
		Use:   "dao",
		Short: "generate dao by cli",
		RunE:  g.run,
	}
	return &cmd
}

func (d *dao) run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	events := d.optimizer.Optimize(ctx, 10, common.DaoPostgresRequest)
	ctx, cancel := util.WithCondition(ctx, len(events))
	errors := make(chan error, len(events))
	template, err := template.NewTemplate(nil)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			state := state.NewDaoState(
				cancel,
				executor.NewTreeExecutor(),
				table.NewTableExecutor(d.extractor),
				output.NewOutputExecutor(template),
			)
			if err := state.Run(ctx, events); err != nil {
				errors <- err
				return
			}
		}()
	}
	go func() {
		wg.Wait()
		close(errors)
	}()
	for err := range errors {
		// ちゃんとwrapする
		if err != nil {
			return err
		}
	}
	return nil
}
