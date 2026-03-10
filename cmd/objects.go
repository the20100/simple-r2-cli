package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/the20100/simple-r2-cli/internal/output"
	"github.com/the20100/simple-r2-cli/internal/validate"
)

var objectsCmd = &cobra.Command{
	Use:   "objects",
	Short: "Manage R2 objects (files)",
}

// ---- objects list ----

var (
	objectsListBucket            string
	objectsListPrefix            string
	objectsListLimit             int32
	objectsListContinuationToken string
)

var objectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List objects in a bucket",
	Long: `List objects in an R2 bucket.

Examples:
  r2 objects list --bucket my-bucket
  r2 objects list --bucket my-bucket --prefix images/
  r2 objects list --bucket my-bucket --limit 50 --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if objectsListBucket == "" {
			return fmt.Errorf("--bucket is required")
		}
		if err := validate.BucketName(objectsListBucket); err != nil {
			return err
		}

		objects, nextToken, err := client.ListObjects(objectsListBucket, objectsListPrefix, objectsListContinuationToken, objectsListLimit)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			result := map[string]any{
				"objects": objects,
			}
			if nextToken != "" {
				result["next_continuation_token"] = nextToken
			}
			return output.PrintJSON(result, output.IsPretty(cmd))
		}
		if len(objects) == 0 {
			fmt.Println("No objects found.")
			return nil
		}
		headers := []string{"KEY", "SIZE", "LAST MODIFIED"}
		rows := make([][]string, len(objects))
		for i, obj := range objects {
			rows[i] = []string{
				output.Truncate(obj.Key, 60),
				output.FormatSize(obj.Size),
				output.FormatTime(obj.LastModified),
			}
		}
		output.PrintTable(headers, rows)
		if nextToken != "" {
			fmt.Printf("\nMore results available. Use --continuation-token %q to get the next page.\n", nextToken)
		}
		return nil
	},
}

// ---- objects get ----

var (
	objectsGetBucket string
	objectsGetOutput string
)

var objectsGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Download an object from a bucket",
	Long: `Download an object from an R2 bucket.

If --output is not specified, the object is written to stdout.

Examples:
  r2 objects get --bucket my-bucket path/to/file.txt
  r2 objects get --bucket my-bucket path/to/file.txt --output ./local-file.txt
  r2 objects get --bucket my-bucket image.png --output ./image.png`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		if objectsGetBucket == "" {
			return fmt.Errorf("--bucket is required")
		}
		if err := validate.BucketName(objectsGetBucket); err != nil {
			return err
		}
		if err := validate.ObjectKey(key); err != nil {
			return err
		}

		if dryRunFlag {
			fmt.Printf("DRY RUN — would download object %q from bucket %q\n", key, objectsGetBucket)
			if objectsGetOutput != "" {
				fmt.Printf("  output: %s\n", objectsGetOutput)
			} else {
				fmt.Println("  output: stdout")
			}
			fmt.Println("No API call made.")
			return nil
		}

		if objectsGetOutput != "" {
			dir := filepath.Dir(objectsGetOutput)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("creating output directory: %w", err)
			}
			f, err := os.Create(objectsGetOutput)
			if err != nil {
				return fmt.Errorf("creating output file: %w", err)
			}
			defer f.Close()

			n, err := client.GetObject(objectsGetBucket, key, f)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Downloaded %s to %s (%s)\n", key, objectsGetOutput, output.FormatSize(n))
			return nil
		}

		_, err := client.GetObject(objectsGetBucket, key, os.Stdout)
		return err
	},
}

// ---- objects head ----

var objectsHeadBucket string

var objectsHeadCmd = &cobra.Command{
	Use:   "head <key>",
	Short: "Get metadata of an object without downloading it",
	Long: `Get metadata (size, content type, last modified, etag) of an object.

Examples:
  r2 objects head --bucket my-bucket path/to/file.txt
  r2 objects head --bucket my-bucket path/to/file.txt --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		if objectsHeadBucket == "" {
			return fmt.Errorf("--bucket is required")
		}
		if err := validate.BucketName(objectsHeadBucket); err != nil {
			return err
		}
		if err := validate.ObjectKey(key); err != nil {
			return err
		}

		detail, err := client.HeadObject(objectsHeadBucket, key)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(detail, output.IsPretty(cmd))
		}
		output.PrintKeyValue([][]string{
			{"Key", detail.Key},
			{"Size", output.FormatSize(detail.Size)},
			{"Content-Type", detail.ContentType},
			{"Last Modified", output.FormatTime(detail.LastModified)},
			{"ETag", detail.ETag},
		})
		return nil
	},
}

// ---- objects put ----

var (
	objectsPutBucket      string
	objectsPutFile        string
	objectsPutContentType string
)

var objectsPutCmd = &cobra.Command{
	Use:   "put <key>",
	Short: "Upload a file to a bucket",
	Long: `Upload a local file to an R2 bucket.

Examples:
  r2 objects put --bucket my-bucket path/to/file.txt --file ./local-file.txt
  r2 objects put --bucket my-bucket images/photo.jpg --file ./photo.jpg --content-type image/jpeg`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		if objectsPutBucket == "" {
			return fmt.Errorf("--bucket is required")
		}
		if objectsPutFile == "" {
			return fmt.Errorf("--file is required")
		}
		if err := validate.BucketName(objectsPutBucket); err != nil {
			return err
		}
		if err := validate.ObjectKey(key); err != nil {
			return err
		}

		if dryRunFlag {
			fmt.Printf("DRY RUN — would upload %q to bucket %q as %q\n", objectsPutFile, objectsPutBucket, key)
			if objectsPutContentType != "" {
				fmt.Printf("  content-type: %s\n", objectsPutContentType)
			}
			fmt.Println("No API call made.")
			return nil
		}

		if err := client.PutObject(objectsPutBucket, key, objectsPutFile, objectsPutContentType); err != nil {
			return err
		}

		fi, _ := os.Stat(objectsPutFile)
		size := ""
		if fi != nil {
			size = " (" + output.FormatSize(fi.Size()) + ")"
		}
		fmt.Fprintf(os.Stderr, "Uploaded %s to %s/%s%s\n", objectsPutFile, objectsPutBucket, key, size)
		return nil
	},
}

// ---- objects delete ----

var objectsDeleteBucket string

var objectsDeleteCmd = &cobra.Command{
	Use:   "delete <key>",
	Short: "Delete an object from a bucket",
	Long: `Delete an object from an R2 bucket. This action is irreversible.

Examples:
  r2 objects delete --bucket my-bucket path/to/file.txt`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		if objectsDeleteBucket == "" {
			return fmt.Errorf("--bucket is required")
		}
		if err := validate.BucketName(objectsDeleteBucket); err != nil {
			return err
		}
		if err := validate.ObjectKey(key); err != nil {
			return err
		}

		if dryRunFlag {
			fmt.Printf("DRY RUN — would delete object %q from bucket %q\n", key, objectsDeleteBucket)
			fmt.Println("No API call made.")
			return nil
		}

		if err := client.DeleteObject(objectsDeleteBucket, key); err != nil {
			return err
		}
		fmt.Printf("Object %s deleted from %s.\n", key, objectsDeleteBucket)
		return nil
	},
}

func init() {
	// list flags
	objectsListCmd.Flags().StringVar(&objectsListBucket, "bucket", "", "Bucket name (required)")
	objectsListCmd.Flags().StringVar(&objectsListPrefix, "prefix", "", "Filter by key prefix")
	objectsListCmd.Flags().Int32Var(&objectsListLimit, "limit", 100, "Maximum number of objects to return")
	objectsListCmd.Flags().StringVar(&objectsListContinuationToken, "continuation-token", "", "Continuation token for pagination")

	// get flags
	objectsGetCmd.Flags().StringVar(&objectsGetBucket, "bucket", "", "Bucket name (required)")
	objectsGetCmd.Flags().StringVar(&objectsGetOutput, "output", "", "Output file path (default: stdout)")

	// head flags
	objectsHeadCmd.Flags().StringVar(&objectsHeadBucket, "bucket", "", "Bucket name (required)")

	// put flags
	objectsPutCmd.Flags().StringVar(&objectsPutBucket, "bucket", "", "Bucket name (required)")
	objectsPutCmd.Flags().StringVar(&objectsPutFile, "file", "", "Local file to upload (required)")
	objectsPutCmd.Flags().StringVar(&objectsPutContentType, "content-type", "", "Content-Type header (auto-detected if omitted)")

	// delete flags
	objectsDeleteCmd.Flags().StringVar(&objectsDeleteBucket, "bucket", "", "Bucket name (required)")

	objectsCmd.AddCommand(
		objectsListCmd,
		objectsGetCmd,
		objectsHeadCmd,
		objectsPutCmd,
		objectsDeleteCmd,
	)
	rootCmd.AddCommand(objectsCmd)

	// Register schemas
	RegisterSchema("objects.list", SchemaEntry{
		Command:     "r2 objects list --bucket <name>",
		Description: "List objects in a bucket",
		Flags: []SchemaFlag{
			{Name: "--bucket", Type: "string", Required: true, Desc: "Bucket name"},
			{Name: "--prefix", Type: "string", Desc: "Filter by key prefix"},
			{Name: "--limit", Type: "int", Default: "100", Desc: "Max objects to return"},
			{Name: "--continuation-token", Type: "string", Desc: "Pagination token"},
		},
		Examples: []string{
			"r2 objects list --bucket my-bucket",
			"r2 objects list --bucket my-bucket --prefix images/ --json",
		},
		Mutating: false,
	})

	RegisterSchema("objects.get", SchemaEntry{
		Command:     "r2 objects get --bucket <name> <key>",
		Description: "Download an object",
		Args:        []SchemaArg{{Name: "key", Required: true, Desc: "Object key"}},
		Flags: []SchemaFlag{
			{Name: "--bucket", Type: "string", Required: true, Desc: "Bucket name"},
			{Name: "--output", Type: "string", Desc: "Output file path (default: stdout)"},
		},
		Examples: []string{
			"r2 objects get --bucket my-bucket file.txt --output ./file.txt",
		},
		Mutating: false,
	})

	RegisterSchema("objects.head", SchemaEntry{
		Command:     "r2 objects head --bucket <name> <key>",
		Description: "Get object metadata without downloading",
		Args:        []SchemaArg{{Name: "key", Required: true, Desc: "Object key"}},
		Flags: []SchemaFlag{
			{Name: "--bucket", Type: "string", Required: true, Desc: "Bucket name"},
		},
		Mutating: false,
	})

	RegisterSchema("objects.put", SchemaEntry{
		Command:     "r2 objects put --bucket <name> <key> --file <path>",
		Description: "Upload a file to a bucket",
		Args:        []SchemaArg{{Name: "key", Required: true, Desc: "Object key (destination path in bucket)"}},
		Flags: []SchemaFlag{
			{Name: "--bucket", Type: "string", Required: true, Desc: "Bucket name"},
			{Name: "--file", Type: "string", Required: true, Desc: "Local file to upload"},
			{Name: "--content-type", Type: "string", Desc: "Content-Type header"},
			{Name: "--dry-run", Type: "bool", Desc: "Validate without uploading"},
		},
		Mutating: true,
	})

	RegisterSchema("objects.delete", SchemaEntry{
		Command:     "r2 objects delete --bucket <name> <key>",
		Description: "Delete an object from a bucket (irreversible)",
		Args:        []SchemaArg{{Name: "key", Required: true, Desc: "Object key"}},
		Flags: []SchemaFlag{
			{Name: "--bucket", Type: "string", Required: true, Desc: "Bucket name"},
			{Name: "--dry-run", Type: "bool", Desc: "Validate without deleting"},
		},
		Mutating: true,
	})
}
