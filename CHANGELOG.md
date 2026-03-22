# Changelog
All notable changes to this project will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
## [v0.4.6] - 2026-03-22

## [v0.4.6] - 2026-03-22

### Fixed
- **builtins**: HTTP builtins now accept a table as the request body — automatically serialized to JSON (`b396e04`)

## [v0.4.5] - 2026-03-20

### Added
- **builtins**: add `printn` for printing without a trailing newline; `print` remains the default with newline (`299a7f7`)

## [v0.4.4] - 2026-03-20

### Added
- **builtins**: add `println` as an alias for `print`; `print` outputs with a trailing newline by default (`fe86934`)

## [v0.4.3] - 2026-03-19

### Added
- **builtins**: add regex builtins — `reMatch`, `reFind`, `reFindAll`, `reReplace`, `reSplit`, `reGroups`

## [v0.4.2] - 2026-03-19

### Added
- **interpreter**: `const` keyword for immutable bindings — reassignment throws a runtime error
- **builtins**: `range(start, end, step?)` builtin for numeric iteration with optional step and countdown support

### Fixed
- **builtins**: `sort()` now handles both string and numeric arrays
- **builtins**: `str()`, `int()`, `float()` added as short aliases for type conversion functions

## [v0.4.1] - 2026-03-19

### Fixed
- **print**: table internals no longer leak into output — nested tables render with proper indentation
- **sort**: now handles string arrays in addition to numeric arrays
- **args**: added bounds checking to prevent panic when no arguments are passed
- **builtins**: added `str()`, `int()`, `float()` as short aliases for type conversion functions

## [v0.4.0] - 2026-03-18

### Added
- **String interpolation** — embed expressions directly in strings with `${}` syntax (`"hello ${name}"`)
- **Pipe operator** `|>` — chain function calls left to right (`arr |> filter(fn) |> map(fn)`)
- **`try` expression** — unwraps result tables and propagates errors, eliminating manual `if !res.ok` boilerplate
- **Postfix `++`/`--`** — increment and decrement operators (`i++`, `i--`)

## [v0.3.2] - 2026-03-17

### Added
- Ternary operator (`condition ? trueBranch : falseBranch`) with support for nested and chained expressions

### Fixed
- `?` token was previously emitted as `ILLEGAL` by golexer — now handled correctly

## [v0.3.1] - 2026-03-17

### Added
- Index variable support in `for in` loops (`for i, v in col {}`)
- Dot access on table literals now works correctly

### Fixed
- Table literal keys defined as bare identifiers are now treated as string keys instead of variable lookups
- Dot assignment (`table.field = value`) now correctly mutates in place
- `args` builtin now strips the binary and script path — user args start at index 0

## [v0.3.0] - 2026-03-17

### Added
- `toJson` and `prettyJson` now return `{ok, value, error}` result objects for proper error handling

### CI
- Release workflow now requires CI to pass before running

## [v0.2.4] - 2026-03-16

### Added
- `else if` chaining in if expressions

### Fixed
- Parser `synchronize()` now recovers on all statement starters (`for`, `switch`, `spawn`, `use`, `break`, `continue`)
- Removed misleading duplicate assignment handling in `parseExpressionStatement`

## [v0.2.3] - 2026-03-13

### Fixed
- HTTP builtins now send proper request headers

## [v0.2.2] - 2026-03-13

### Fixed
- Shell errors are no longer silently swallowed

## [v0.2.1] - 2026-03-13

### Added
- `confirm` — interactive CLI confirmation prompt
- `select` — interactive CLI selection menu

## [v0.2.0] - 2026-03-13

### Added
- `prettyJson` builtin for formatted JSON output

## [v0.1.1] - 2026-03-13

### Fixed
- `lgs build` now correctly uses embedded stdlib files

## [v0.0.5] - 2026-03-13

### Fixed
- goreleaser config updated for v2 syntax

## [v0.0.4] - 2026-03-13

### CI
- Use goreleaser prebuilt binary on Linux runner

## [v0.0.3] - 2026-03-13

### Fixed
- Compound assignment on undeclared variables now handled correctly

## [v0.0.2] - 2026-03-13

### Fixed
- Removed invalid `files` entry from goreleaser archives config

## [v0.0.1] - 2026-03-13

### Added
- Initial release — full language interpreter with:
  - Parser, AST, and tree-walking evaluator
  - If/else, for, for-in, switch, break, continue
  - Functions, arrow functions, closures
  - Tables, arrays, strings, booleans, null
  - `&&`/`||` with short-circuit evaluation
  - Compound assignment operators (`+=`, `-=`, `*=`, `/=`, `%=`)
  - Built-in functions for I/O, strings, arrays, math, JSON, HTTP, file I/O, time, color output
  - CLI utilities (`confirm`, `select`, `prompt`)
  - Concurrency via `spawn` blocks and `spawn for-in`
  - Standard library (`std/array`, `std/string`, `std/math`, `std/path`, `std/time`, `std/type`, `std/log`, `std/testing`)
  - Sandbox mode with configurable capability restrictions
  - Go embedding API (`Register`, `SetVar`, `GetVar`, `Call`, `Run`)
  - Binary compilation via `lgs build`
  - CI/CD with GoReleaser
