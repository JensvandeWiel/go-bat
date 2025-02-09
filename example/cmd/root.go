package cmd

import (
	"context"
	"github.com/JensvandeWiel/go-bat/example/cmd/frontend"
	bat "github.com/JensvandeWiel/go-bat/pkg"
	valkeyTest "github.com/testcontainers/testcontainers-go/modules/valkey"
	"github.com/valkey-io/valkey-go"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "example",
	Short: "Example app with bat",
	RunE:  Run,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.example.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func Run(cmd *cobra.Command, args []string) error {
	valkeyContainer, err := valkeyTest.Run(context.Background(), "docker.io/valkey/valkey:7.2.5")
	if err != nil {
		return err
	}

	defer valkeyContainer.Terminate(context.Background())

	connStr, err := valkeyContainer.ConnectionString(context.Background())
	if err != nil {
		return err
	}

	client, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{strings.TrimLeft(connStr, "redis://")}})
	if err != nil {
		return err
	}
	defer client.Close()

	logger := bat.NewLogger(bat.LoggerOutputTypeHuman, &slog.HandlerOptions{Level: slog.LevelDebug}, false)
	if err != nil {
		return err
	}
	sExt, err := bat.NewSessionExtension()
	if err != nil {
		return err
	}

	fExt, err := bat.NewFlashExtension()
	if err != nil {
		return err
	}

	iExt, err := bat.NewInertiaExtension(frontend.DistDirFS, frontend.Manifest, true, bat.WithFrontendPath("./cmd/frontend"))
	if err != nil {
		return err
	}
	b, err := bat.NewBat(logger, bat.NewValkeyExtension(client),
		sExt,
		iExt,
		fExt)
	if err != nil {
		return err
	}
	err = b.RegisterControllers(&MainController{})
	if err != nil {
		return err
	}
	return b.Start(":8080")
}
