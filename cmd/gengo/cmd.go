package main

import (
	"fmt"
	"log"

	"github.com/naonao2323/testgen/pkg/cli/dao"
	"github.com/naonao2323/testgen/pkg/cli/gengo"
	"github.com/spf13/cobra"
)

func printMascot() {
	fmt.Println("\033[34m GGG   EEEEE  N   N  GGG   OOO")
	fmt.Println("G   G  E      NN  N G   G O   O")
	fmt.Println("G      EEEE   N N N G     O   O")
	fmt.Println("G  GG  E      N  NN G  GG O   O")
	fmt.Println(" GGG   EEEEE  N   N  GGG   OOO\033[0m")
}

func runParent(cmd *cobra.Command, args []string) {
	printMascot()
}

func main() {
	gengoCmd := &cobra.Command{
		Use:              "gengo",
		Short:            "Gengo CLI tool",
		PersistentPreRun: runParent,
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
