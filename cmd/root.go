package cmd

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/api"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/config"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/tui"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/version"
)

var (
	serverURL string
	cfg       *config.Config
	client    *api.Client
)

// rootCmd is the root cobra command.
var rootCmd = &cobra.Command{
	Use:   "cvx",
	Short: "CertVaultCLIX — Interactive TUI for CertVault",
	Long: `CertVaultCLIX (cvx) is an interactive Terminal UI client
for the CertVault self-signed SSL certificate management platform.

Run without arguments to launch the interactive TUI, or use
subcommands for scripting and automation.`,
	Version: version.String(),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTUI()
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "", "CertVault server URL (overrides config)")
}

func initConfig() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load config: %v\n", err)
		cfg = &config.Config{ServerURL: config.DefaultServerURL}
	}

	if serverURL != "" {
		cfg.ServerURL = serverURL
	}

	client = api.NewClient(cfg.ServerURL)
	if cfg.Session != "" {
		client.SetSession(cfg.Session)
	}
}

func runTUI() error {
	app := tui.NewApp(client, cfg)
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Save session on exit
	if finalApp, ok := finalModel.(*tui.App); ok {
		session := finalApp.Client().GetSession()
		if session != "" && cfg != nil {
			cfg.Session = session
			_ = config.Save(cfg)
		}
	}
	return nil
}

// pingCmd checks server connectivity.
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Check server connectivity",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.Ping(context.Background()); err != nil {
			return fmt.Errorf("server unreachable: %w", err)
		}
		fmt.Printf("✓ Connected to %s\n", cfg.ServerURL)
		return nil
	},
}

// versionCmd prints the version.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("cvx " + version.String())
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)
	rootCmd.AddCommand(versionCmd)
}
