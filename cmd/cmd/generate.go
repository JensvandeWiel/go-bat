/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/JensvandeWiel/go-bat/internal"
	"github.com/JensvandeWiel/go-bat/pkg"
	"log/slog"

	"github.com/spf13/cobra"
)

var dir string

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate [item] [name]",
	Short: "Generate a new item: model, controller",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := pkg.NewLogger(pkg.LoggerOutputTypeHuman, &slog.HandlerOptions{Level: slog.LevelDebug}, false)
		gen, err := internal.ParseGenerator(args[0], logger, dir)
		if err != nil {
			return err
		}

		logger.Info("Generating", "item", args[0], "name", args[1])
		err = gen.Generate(args[1])
		if err != nil {
			return err
		}

		logger.Info("Generated", "item", args[0], "name", args[1])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	generateCmd.Flags().StringVar(&dir, "dir", "", "The directory of the project, defaults to \".\"")
}
