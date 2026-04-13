# docpush

Markdown → Confluence automation. Converts markdown documentation to Confluence storage format and publishes it via a GitHub Action (primary) or CLI (secondary).

## Install

```bash
go install https://github.com/micahcourey/docpush/cmd/docpush@latest
```

Or download a pre-built binary from the [releases page].

## Usage

### Sync (publish to Confluence)

```bash
# Sync specific files
docpush sync --target confluence --files docs/features/mass-communications.md

# Dry-run (preview converted XHTML without publishing)
docpush sync --dry-run --files docs/features/mass-communications.md

# Sync all mapped files
docpush sync --target confluence --all --create-if-missing
```

### Status

```bash
docpush status --target confluence
```

### Link (associate file → Confluence page)

```bash
docpush link docs/features/foo.md 123456789 --target confluence
```

### Init (scaffold config)

```bash
docpush init
```

## Configuration

Create a `.docpush.yaml` in your repo root:

```yaml
targets:
  confluence:
    type: confluence
    url: https://<your-domain>.atlassian.net
    space: <space-name>
    defaults:
      parentId: 987654321
      labels:
        - feature-doc
        - auto-synced

pages:
  docs/features/mass-communications.md:
    confluence:
      pageId: 123456789
  docs/features/file-exchange-storage.md:
    confluence:
      pageId: 123456790
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `CONFLUENCE_URL` | Confluence base URL (overrides config file) |
| `CONFLUENCE_PAT` | Personal Access Token for authentication |

## GitHub Action

The primary publish path. See `.github/workflows/docpush.yml` in the docs repo.
