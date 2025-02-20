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
var packageName string
var noGit bool
var noInstallDeps bool

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringArrayVar(&extras, "extra", []string{}, "Add extra features to the project choice of: inertia-react, inertia-svelte, database-pgsql, frontend-auth")
	newCmd.Flags().StringVar(&packageName, "package-name", "", "The name of the package, defaults to the project name")
	newCmd.Flags().BoolVar(&force, "force", false, "Force the creation of the project even if the directory is not empty")
	newCmd.Flags().BoolVar(&noGit, "no-git", false, "Do not create a git repository")
	newCmd.Flags().BoolVar(&noInstallDeps, "no-installdeps", false, "Does not install dependencies after creating the project")

}

func RunNew(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	logger := pkg.NewLogger(pkg.LoggerOutputTypeHuman, &slog.HandlerOptions{Level: slog.LevelDebug}, false)

	workDir, err := os.Getwd()
	if err != nil {
		return err
	}
	if packageName == "" {
		packageName = projectName
	}
	p, err := internal.NewProject(projectName, packageName, workDir, force, noInstallDeps, noGit, logger, internal.ParseExtras(internal.ParseExtraTypes(extras))...)
	if err != nil {
		return err
	}

	err = p.Create()
	if err != nil {
		return err
	}

	return nil
}
