package cmd

import (
	"github.com/JensvandeWiel/go-bat/internal"
	"github.com/JensvandeWiel/go-bat/pkg"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new go-bat project",
	Args:  cobra.ExactArgs(1),
	RunE:  RunNew,
}

var extras []string
var force bool

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringArrayVar(&extras, "extra", []string{}, "Add extra features to the project choice of: inertia-react")
	newCmd.Flags().BoolVar(&force, "force", false, "Force the creation of the project even if the directory is not empty")
}

func RunNew(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	logger := pkg.NewLogger(pkg.LoggerOutputTypeHuman, &slog.HandlerOptions{Level: slog.LevelDebug}, false)

	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	p, err := internal.NewProject(projectName, projectName, workDir, force, logger, internal.ParseExtras(extras...)...)
	if err != nil {
		return err
	}

	err = p.Create()
	if err != nil {
		return err
	}

	return nil
}
