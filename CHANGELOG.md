# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
## [0.3.2] - 2026-03-17
## [0.3.2] - 2026-03-17
### Added
- Ternary operator (`condition ? trueBranch : falseBranch`)
- Support for nested and chained ternary expressions
### Fixed
- Added `?` token to golexer, was previously emitted as `ILLEGAL`
## [0.3.1] - 2026-03-17
### Added
- Dot access on table literals now works correctly (keys stored as strings)
- Index variable support in `for in` statements (`for i, v in col {}`)

### Fixed
- Table literal keys defined as bare identifiers are now treated as string keys instead of environment lookups
- Dot assignment (`table.field = value`) now correctly mutates table fields
- Built-in args now start at index 2
## [0.3.0] - 2026-03-17
### Added
- `toJson` and `prettyJson` now return `{ok, value, error}` result objects for proper error handling

### CI
- Release workflow now requires CI to pass before running

## [0.2.4] - 2026-03-16
### Added
- `else if` support in if expressions

### Fixed
- Parser `synchronize()` now recovers on all statement starters (`for`, `switch`, `spawn`, `use`, `break`, `continue`)
- Removed misleading duplicate assignment handling in `parseExpressionStatement` (added clarifying comment)

## [0.2.3] - 2026-03-13
### Fixed
- HTTP builtins now send proper headers

## [0.2.2] - 2026-03-13
### Fixed
- Shell errors are no longer swallowed; proper error handling added

## [0.2.1] - 2026-03-13
### Added
- `confirm` CLI utility for user confirmation prompts
- `select` CLI utility for interactive selection menus

## [0.2.0] - 2026-03-13
### Added
- `prettyJson` for nicer JSON formatting

## [0.1.1] - 2026-03-13
### Fixed
- Build command now uses embedded std files

## [0.0.5] - 2026-03-13
### Fixed
- Corrected goreleaser config for v2 syntax

## [0.0.4] - 2026-03-13
### CI
- Use goreleaser prebuilt binary for linux runner

## [0.0.3] - 2026-03-13
### Fixed
- Properly handle compound assignment on undeclared variables

## [0.0.2] - 2026-03-13
### Fixed
- Removed invalid files entry from archives config

## [0.0.1] - 2026-03-13
### Added
- Initial release of the language interpreter

### Features
- Basic parser with statements
- If/else blocks
- For loops
- Functions, arrow functions, closures
- Switch statements
- Hashmaps, arrays, booleans, strings, null
- && and || operators
- Compound assignment operators (+=, -=)
- Built-in functions
- CLI utilities
- Concurrency with spawn blocks
- Standard library
- Installer with progress indicator
- CI/CD with GoReleaser for releases
