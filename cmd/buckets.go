package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/the20100/simple-r2-cli/internal/output"
)

var bucketsCmd = &cobra.Command{
	Use:   "buckets",
	Short: "Manage R2 buckets",
}

var bucketsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all R2 buckets",
	Long: `List all R2 buckets in the account.

Examples:
  r2 buckets list
  r2 buckets list --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		buckets, err := client.ListBuckets()
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(buckets, output.IsPretty(cmd))
		}
		if len(buckets) == 0 {
			fmt.Println("No buckets found.")
			return nil
		}
		headers := []string{"NAME", "CREATED"}
		rows := make([][]string, len(buckets))
		for i, b := range buckets {
			rows[i] = []string{
				b.Name,
				output.FormatTime(b.CreationDate),
			}
		}
		output.PrintTable(headers, rows)
		return nil
	},
}

func init() {
	bucketsCmd.AddCommand(bucketsListCmd)
	rootCmd.AddCommand(bucketsCmd)

	RegisterSchema("buckets.list", SchemaEntry{
		Command:     "r2 buckets list",
		Description: "List all R2 buckets in the account",
		Examples:    []string{"r2 buckets list", "r2 buckets list --json"},
		Mutating:    false,
	})
}
