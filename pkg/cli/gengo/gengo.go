package gengo

import (
	"fmt"

	"github.com/spf13/cobra"
)

type gengo struct {
	destination string
}

func NewCommand() *cobra.Command {
	g := gengo{}
	cmd := cobra.Command{
		Use:   "gen",
		Short: "generate testing framework by cli",
		RunE:  g.run,
	}
	cmd.Flags().StringVar(&g.destination, "destination", g.destination, "output path of the artifact")
	return &cmd
}

func (a *gengo) run(cmd *cobra.Command, args []string) error {
	fmt.Println("gengo")
	return nil
}
