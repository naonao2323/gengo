package dao

import (
	"context"
	"errors"
	"sync"

	"github.com/naonao2323/testgen/pkg/common"
	"github.com/naonao2323/testgen/pkg/config"
	"github.com/naonao2323/testgen/pkg/config/yaml"
	"github.com/naonao2323/testgen/pkg/executor"
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
	extractor  extractor.Extractor
	optimizer  optimizer.Optimizer
	config     config.Config
	confPath   string
	outputPath string
}

func NewCommand() *cobra.Command {
	d := &dao{}
	cmd := cobra.Command{
		Use:     "dao",
		Short:   "generate dao by cli",
		RunE:    d.run,
		PreRunE: d.setup,
	}
	cmd.Flags().StringVar(&d.confPath, "path", d.confPath, "config file path")
	cmd.Flags().StringVar(&d.outputPath, "outputPath", d.outputPath, "output dir path")
	return &cmd
}

func (d *dao) setup(cmd *cobra.Command, args []string) error {
	if d.confPath == "" {
		return errors.New("undefined conf path")
	}
	if d.outputPath == "" {
		return errors.New("undefined output path")
	}
	config, err := yaml.NewConfig(d.confPath)
	if err != nil {
		return err
	}
	d.config = config
	ctx := context.Background()
	extractor := extractor.Extract(ctx, extractor.Postgres, d.config.GetSchema(), d.config.GetDbUrl())
	optimizer := optimizer.NewOptimizer(extractor)
	d.extractor = extractor
	d.optimizer = optimizer
	return nil
}

func (d *dao) run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	events := d.optimizer.Optimize(ctx, d.config.GetInclude(), common.DaoPostgresRequest)
	ctx, cancel := util.WithCondition(ctx, len(events))
	errors := make(chan error, len(events))
	template, err := template.NewTemplate(nil)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	for i := 0; i < d.config.GetParallel(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			state := state.NewDaoState(
				cancel,
				executor.NewTreeExecutor(),
				table.NewTableExecutor(d.extractor),
				output.NewOutputExecutor(template, d.outputPath),
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
