package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/the20100/simple-r2-cli/internal/config"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Cloudflare R2 authentication",
}

var authSetupCmd = &cobra.Command{
	Use:   "setup <account-id> <access-key-id> <secret-access-key>",
	Short: "Save R2 credentials to the config file",
	Long: `Save Cloudflare R2 credentials to the local config file.

Get your R2 API token from the Cloudflare dashboard:
  R2 Object Storage > Overview > Manage R2 API Tokens

You need three values:
  1. Account ID (visible in the dashboard URL)
  2. Access Key ID (from the API token)
  3. Secret Access Key (shown once when creating the token)

The credentials are stored at:
  macOS:   ~/Library/Application Support/r2/config.json
  Linux:   ~/.config/r2/config.json
  Windows: %AppData%\r2\config.json

You can also set these env vars instead:
  R2_ACCOUNT_ID, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		accountID := args[0]
		accessKeyID := args[1]
		secretAccessKey := args[2]

		if len(accountID) < 8 {
			return fmt.Errorf("account ID looks too short")
		}
		if len(accessKeyID) < 8 {
			return fmt.Errorf("access key ID looks too short")
		}
		if len(secretAccessKey) < 8 {
			return fmt.Errorf("secret access key looks too short")
		}

		if err := config.Save(&config.Config{
			AccountID:       accountID,
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAccessKey,
		}); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Printf("Credentials saved to %s\n", config.Path())
		fmt.Printf("Account ID:        %s\n", maskOrEmpty(accountID))
		fmt.Printf("Access Key ID:     %s\n", maskOrEmpty(accessKeyID))
		fmt.Printf("Secret Access Key: %s\n", maskOrEmpty(secretAccessKey))
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		fmt.Printf("Config: %s\n\n", config.Path())

		envAccount := os.Getenv("R2_ACCOUNT_ID")
		envKey := os.Getenv("R2_ACCESS_KEY_ID")
		envSecret := os.Getenv("R2_SECRET_ACCESS_KEY")

		if envAccount != "" && envKey != "" && envSecret != "" {
			fmt.Println("Source: environment variables (takes priority over config)")
			fmt.Printf("Account ID:        %s\n", maskOrEmpty(envAccount))
			fmt.Printf("Access Key ID:     %s\n", maskOrEmpty(envKey))
			fmt.Printf("Secret Access Key: %s\n", maskOrEmpty(envSecret))
		} else if c.AccountID != "" && c.AccessKeyID != "" && c.SecretAccessKey != "" {
			fmt.Println("Source: config file")
			fmt.Printf("Account ID:        %s\n", maskOrEmpty(c.AccountID))
			fmt.Printf("Access Key ID:     %s\n", maskOrEmpty(c.AccessKeyID))
			fmt.Printf("Secret Access Key: %s\n", maskOrEmpty(c.SecretAccessKey))
		} else {
			fmt.Println("Status: not authenticated")
			fmt.Printf("\nRun: r2 auth setup <account-id> <access-key-id> <secret-access-key>\n")
			fmt.Printf("Or:  export R2_ACCOUNT_ID=... R2_ACCESS_KEY_ID=... R2_SECRET_ACCESS_KEY=...\n")
		}
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove saved credentials from the config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Clear(); err != nil {
			return fmt.Errorf("removing config: %w", err)
		}
		fmt.Println("Credentials removed from config.")
		return nil
	},
}

func init() {
	authCmd.AddCommand(authSetupCmd, authStatusCmd, authLogoutCmd)
	rootCmd.AddCommand(authCmd)
}
