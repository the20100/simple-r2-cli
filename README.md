# r2 — Cloudflare R2 CLI

Simple CLI to list, read, write, and delete objects in [Cloudflare R2](https://developers.cloudflare.com/r2/) object storage via the S3-compatible API.

## Install

```bash
git clone https://github.com/the20100/simple-r2-cli
cd simple-r2-cli
go build -o r2 .
mv r2 /usr/local/bin/
cd ..
rm -rf simple-r2-cli
```

Requires Go 1.22+.

## Authentication

Get your R2 API token from the Cloudflare dashboard: **R2 Object Storage > Overview > Manage R2 API Tokens**.

You need three values:
- **Account ID** — visible in the dashboard URL
- **Access Key ID** — from the API token
- **Secret Access Key** — shown once when creating the token

### Option A: Config file

```bash
r2 auth setup <account-id> <access-key-id> <secret-access-key>
```

### Option B: Environment variables

```bash
export R2_ACCOUNT_ID=your-account-id
export R2_ACCESS_KEY_ID=your-access-key-id
export R2_SECRET_ACCESS_KEY=your-secret-access-key
```

## Usage

```bash
# List buckets
r2 buckets list

# List objects in a bucket
r2 objects list --bucket my-bucket
r2 objects list --bucket my-bucket --prefix images/ --limit 50

# Get object metadata
r2 objects head --bucket my-bucket path/to/file.txt

# Download an object
r2 objects get --bucket my-bucket path/to/file.txt --output ./local-file.txt

# Upload a file
r2 objects put --bucket my-bucket path/to/file.txt --file ./local-file.txt

# Delete an object
r2 objects delete --bucket my-bucket path/to/file.txt
```

## Output

- **Terminal**: human-readable tables
- **Piped/redirected**: JSON (auto-detected)
- `--json`: force JSON output
- `--pretty`: force pretty-printed JSON

## Agent usage

This CLI is designed for AI agent consumption:

- **`r2 schema`** — dump all command schemas as JSON for introspection
- **`r2 schema objects.list`** — single command schema
- **`--dry-run`** — validate inputs without hitting the API (put, delete)
- **`--fields`** — limit response fields to reduce context window usage
- Auto-JSON output when stdout is not a TTY

## Self-update

```bash
r2 update
```

Pulls the latest source from GitHub, rebuilds, and atomically replaces the current binary.
