# Changelog

All notable changes to docpush will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-04-14

### Added

- `scaffold` command — generates markdown stubs from existing Confluence child pages under a given parent
- `--parent-id` flag to specify the Confluence parent page to scaffold from
- `--output-dir` flag to control where generated stubs are written (default: `docs/features`)
- `--dry-run` flag to preview scaffold output without writing files
- Paginated `GetChildPages` Confluence API client method
- `WriteConfig` helper to persist updated `.docpush.yaml` after scaffolding

## [0.2.0] - 2026-04-14

### Added

- Source banner prepended to Confluence pages with an info panel linking back to the GitHub source file
- Read-only page restrictions via Confluence REST API — locks pages so only the service account can edit
- `sourceBaseUrl` config field in `.docpush.yaml` defaults to generate source links
- `readOnly` config field in `.docpush.yaml` defaults to enable/disable page restrictions

## [0.1.1] - 2026-04-14

### Fixed

- Page titles now fall back to the top-level `title` frontmatter field when no target-specific title is set
- Pages that only had `title` in frontmatter (not under `docpush.confluence.title`) were using the file path as the page title

## [0.1.0] - 2026-04-14

### Fixed

- `.gitignore` bare `docpush` pattern was excluding the `cmd/docpush/` directory, causing `go install` to fail with "module found but does not contain package"
- Changed `.gitignore` entry from `docpush` to `/docpush` to only match the root binary

### Added

- Initial tagged release
- Markdown-to-Confluence sync via YAML frontmatter configuration
- Confluence Data Center REST API client (create, update, search, labels)
- `.docpush.yaml` config file with targets, defaults, and per-page overrides
- Goldmark-based Markdown-to-XHTML conversion
- YAML frontmatter parsing for page metadata (title, pageId, parentId, labels)
- Diff-based sync — only updates pages when content has changed
- Page ID write-back to `.docpush.yaml` after creating new pages
- Label management on Confluence pages
- CLI with `sync` command and `--dry-run` flag

[0.3.0]: https://github.com/micahcourey/docpush/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/micahcourey/docpush/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/micahcourey/docpush/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/micahcourey/docpush/releases/tag/v0.1.0
