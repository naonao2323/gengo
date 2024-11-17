package main

import (
	"fmt"
	"log"

	"github.com/naonao2323/testgen/pkg/cli/dao"
	"github.com/naonao2323/testgen/pkg/cli/gengo"
	"github.com/spf13/cobra"
)

func main() {
	gengoCmd := &cobra.Command{
		Use:   "gengo",
		Short: "Gengo CLI tool",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Please use a subcommand like 'gengo gen' or 'gengo dao'.")
		},
	}
	gengoCmd.AddCommand(gengo.NewCommand())
	gengoCmd.AddCommand(dao.NewCommand())
	if err := Execute(gengoCmd); err != nil {
		log.Fatal(err)
	}
}

func Execute(cmd *cobra.Command) error {
	err := cmd.Execute()
	if err != nil {
		return err
	}
	return nil
}
