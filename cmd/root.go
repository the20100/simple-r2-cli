package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/the20100/simple-r2-cli/internal/config"
	"github.com/the20100/simple-r2-cli/internal/r2"
)

var (
	jsonFlag   bool
	prettyFlag bool
	fieldsFlag string
	dryRunFlag bool
	client     *r2.Client
	cfg        *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "r2",
	Short: "Cloudflare R2 CLI — manage R2 object storage via the S3-compatible API",
	Long: `r2 is a CLI tool for Cloudflare R2 object storage.

It outputs JSON when piped (for agent use) and human-readable tables in a terminal.

Credential resolution order:
  1. R2_ACCESS_KEY_ID + R2_SECRET_ACCESS_KEY + R2_ACCOUNT_ID env vars
  2. Config file  (~/.config/r2/config.json  via: r2 auth setup)

Examples:
  r2 auth setup
  r2 buckets list
  r2 objects list --bucket my-bucket
  r2 objects get --bucket my-bucket path/to/file.txt
  r2 objects put --bucket my-bucket path/to/file.txt --file ./local-file.txt
  r2 objects delete --bucket my-bucket path/to/file.txt`,
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Force JSON output")
	rootCmd.PersistentFlags().BoolVar(&prettyFlag, "pretty", false, "Force pretty-printed JSON output (implies --json)")
	rootCmd.PersistentFlags().StringVar(&fieldsFlag, "fields", "", "Comma-separated list of fields to include in response (reduces output size)")
	rootCmd.PersistentFlags().BoolVar(&dryRunFlag, "dry-run", false, "Validate the request locally without hitting the API")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if isAuthCommand(cmd) || cmd.Name() == "info" || cmd.Name() == "update" || cmd.Name() == "schema" {
			return nil
		}
		accountID, accessKeyID, secretAccessKey, err := resolveCredentials()
		if err != nil {
			return err
		}
		client, err = r2.NewClient(accountID, accessKeyID, secretAccessKey)
		if err != nil {
			return fmt.Errorf("creating R2 client: %w", err)
		}
		return nil
	}

	rootCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show tool info: config path, auth status, and environment",
	Run: func(cmd *cobra.Command, args []string) {
		printInfo()
	},
}

func printInfo() {
	fmt.Printf("r2 — Cloudflare R2 CLI\n\n")
	exe, _ := os.Executable()
	fmt.Printf("  binary:  %s\n", exe)
	fmt.Printf("  os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()
	fmt.Println("  config paths by OS:")
	fmt.Printf("    macOS:    ~/Library/Application Support/r2/config.json\n")
	fmt.Printf("    Linux:    ~/.config/r2/config.json\n")
	fmt.Printf("    Windows:  %%AppData%%\\r2\\config.json\n")
	fmt.Printf("  config:   %s\n", config.Path())
	fmt.Println()
	fmt.Printf("    R2_ACCOUNT_ID      = %s\n", maskOrEmpty(os.Getenv("R2_ACCOUNT_ID")))
	fmt.Printf("    R2_ACCESS_KEY_ID   = %s\n", maskOrEmpty(os.Getenv("R2_ACCESS_KEY_ID")))
	fmt.Printf("    R2_SECRET_ACCESS_KEY = %s\n", maskOrEmpty(os.Getenv("R2_SECRET_ACCESS_KEY")))
}

func maskOrEmpty(v string) string {
	if v == "" {
		return "(not set)"
	}
	if len(v) <= 8 {
		return "***"
	}
	return v[:4] + "..." + v[len(v)-4:]
}

func resolveEnv(names ...string) string {
	for _, name := range names {
		if v := os.Getenv(name); v != "" {
			return v
		}
	}
	return ""
}

func resolveCredentials() (string, string, string, error) {
	accountID := resolveEnv(
		"R2_ACCOUNT_ID",
		"CF_ACCOUNT_ID",
		"CLOUDFLARE_ACCOUNT_ID",
	)
	accessKeyID := resolveEnv(
		"R2_ACCESS_KEY_ID",
		"R2_KEY",
		"R2_API_KEY",
		"AWS_ACCESS_KEY_ID",
	)
	secretAccessKey := resolveEnv(
		"R2_SECRET_ACCESS_KEY",
		"R2_SECRET",
		"R2_API_SECRET",
		"AWS_SECRET_ACCESS_KEY",
	)

	if accountID != "" && accessKeyID != "" && secretAccessKey != "" {
		return accountID, accessKeyID, secretAccessKey, nil
	}

	var err error
	cfg, err = config.Load()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to load config: %w", err)
	}
	if cfg.AccountID != "" && cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		if accountID == "" {
			accountID = cfg.AccountID
		}
		if accessKeyID == "" {
			accessKeyID = cfg.AccessKeyID
		}
		if secretAccessKey == "" {
			secretAccessKey = cfg.SecretAccessKey
		}
		if accountID != "" && accessKeyID != "" && secretAccessKey != "" {
			return accountID, accessKeyID, secretAccessKey, nil
		}
	}
	return "", "", "", fmt.Errorf("not authenticated — run: r2 auth setup\nor set R2_ACCOUNT_ID, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY env vars")
}

func isAuthCommand(cmd *cobra.Command) bool {
	if cmd.Name() == "auth" {
		return true
	}
	p := cmd.Parent()
	for p != nil {
		if p.Name() == "auth" {
			return true
		}
		p = p.Parent()
	}
	return false
}
