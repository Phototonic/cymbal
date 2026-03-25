# cymbal

Fast, language-agnostic code indexer and symbol navigator built on [tree-sitter](https://tree-sitter.github.io/).

cymbal parses your codebase into a local SQLite index, then gives you instant symbol search, cross-references, impact analysis, and scoped diffs — all from the command line. Designed to be called by AI agents, editor plugins, or directly from your terminal.

## Install

Homebrew:

```sh
brew install 1broseidon/tap/cymbal
```

Or with Go (requires CGO for tree-sitter + SQLite):

```sh
CGO_CFLAGS="-DSQLITE_ENABLE_FTS5" go install github.com/1broseidon/cymbal@latest
```

Or grab a binary from [releases](https://github.com/1broseidon/cymbal/releases).

## Quick start

```sh
# Index the current project
cymbal index .

# Investigate any symbol — one call, right answer
cymbal investigate handleAuth    # function → source + callers + impact
cymbal investigate UserModel     # type → definition + members + references

# Or use specific commands when you need control
cymbal search handleAuth         # find a symbol
cymbal search "TODO" --text      # full-text grep
cymbal show handleAuth           # read source
cymbal outline internal/auth/handler.go  # file structure
cymbal refs handleAuth           # who calls this?
cymbal importers internal/auth   # who imports this package?
cymbal impact handleAuth         # what breaks if I change this?
cymbal diff handleAuth main      # git diff scoped to a function
cymbal context handleAuth        # bundled: source + types + callers + imports
cymbal ls                        # file tree
```

## Commands

| Command | What it does |
|---------|-------------|
| `investigate` | **Start here.** Kind-adaptive exploration — one call, right shape |
| `index` | Parse and index a directory |
| `ls` | File tree, repo list, or `--stats` overview |
| `search` | Symbol search (or `--text` for grep) |
| `show` | Display a symbol's source code |
| `outline` | List all symbols in a file |
| `refs` | Find references / call sites |
| `importers` | Reverse import lookup — who imports this? |
| `impact` | Transitive callers — what's affected by a change? |
| `diff` | Git diff scoped to a symbol's line range |
| `context` | Bundled view: source + types + callers + imports |

All commands support `--json` for structured output.

## Agent integration

cymbal is designed as the code navigation layer for AI agents. One command handles most investigations — specific commands exist as escape hatches when you need more control.

Add this to your agent's system prompt (e.g., `CLAUDE.md`, `agent.md`, or MCP tool descriptions):

```markdown
## Code Exploration Policy
Use `cymbal` CLI for code navigation — prefer it over Read, Grep, Glob, or Bash for code exploration.
- **Default**: `cymbal investigate <symbol>` — returns source, callers, impact, or members based on what the symbol is. Use this first.
- Before reading a file: `cymbal outline <file>` or `cymbal show <file:L1-L2>`
- Before searching: `cymbal search <query>` (symbols) or `cymbal search <query> --text` (grep)
- Before exploring structure: `cymbal ls` (tree) or `cymbal ls --stats` (overview)
- To find usage: `cymbal refs <symbol>` or `cymbal refs <symbol> --importers`
- If a project is not indexed, run `cymbal index .` first (takes <100ms).
- Use `cymbal show <symbol>` to read a specific function/type instead of reading the whole file.
- All commands support `--json` for structured output.
```

### Why this works

An agent investigating a function typically makes 2-3 sequential tool calls: search → show → refs. Each call costs a reasoning step (~500 tokens of "let me think about what to call next"). `cymbal investigate` collapses that into one call — cymbal looks at the symbol's kind and returns the right shape:

| Symbol kind | What you get |
|---|---|
| function/method | Source + callers + shallow impact chain |
| class/struct/type | Source + members + references |
| ambiguous name | Ranked candidates with file and kind |

The agent says "I'm looking at X" and cymbal says "here's what you need to know about X, given what X is."

## Supported languages

cymbal uses tree-sitter grammars. Currently supported:

Go, Python, JavaScript, TypeScript, TSX, Rust, C, C++, C#, Java, Ruby, Swift, Kotlin, Scala, PHP, Lua, Bash, HCL, Dockerfile, YAML, TOML, HTML, CSS

Adding a language requires a tree-sitter grammar and a symbol extraction query — see `internal/parser/` for examples.

## How it works

1. **Index** — tree-sitter parses each file into an AST. cymbal extracts symbols (functions, types, variables, imports) and references (calls, type usage) and stores them in SQLite with FTS5 full-text search. Each repo gets its own database at `~/.cymbal/repos/<hash>/index.db`.

2. **Query** — all commands read from the current repo's SQLite index. Symbol lookups, cross-references, and import graphs are SQL queries. No re-parsing needed. No cross-repo bleed.

3. **Incremental** — re-indexing skips unchanged files using mtime (nanosecond) + file size checks. Only changed files are re-parsed and re-hashed. Reindex completes in 2-15ms for most repos.

## Docs

- [Changelog](./CHANGELOG.md)

## License

[MIT](./LICENSE)
