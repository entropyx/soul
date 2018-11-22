package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "soul",
	Short: "Code, test and deploy your Golang microservice with a single tool",
	Long: `Code, test and deploy your Golang microservice with a single tool.
                Complete documentation is available at https://github.com/entropyx/soul`,
}

func Execute() {
	rootCmd.AddCommand(configureCmd)
	configureCmd.AddCommand(configureCodeshipCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
