package dao

import (
	"fmt"

	"github.com/spf13/cobra"
)

type dao struct{}

func NewCommand() *cobra.Command {
	g := dao{}
	cmd := cobra.Command{
		Use:   "dao",
		Short: "generate dao by cli",
		RunE:  g.run,
	}
	return &cmd
}

func (a *dao) run(cmd *cobra.Command, args []string) error {
	fmt.Println("dao")
	return nil
}
