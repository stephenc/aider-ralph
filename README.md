# aider-ralph

The **Ralph Wiggum AI Loop Technique** for [Aider](https://aider.chat/).

An iterative AI development methodology that repeatedly runs `aider` with a prompt until completion. Named after The Simpsons character, it embodies the philosophy of persistent iteration despite setbacks.

> *"I'm learnding!"* — Ralph Wiggum

## Philosophy

- **Iteration > Perfection**: Don't aim for perfect on first try. Let the loop refine the work.
- **Failures Are Data**: Deterministically bad means failures are predictable and informative.
- **Operator Skill Matters**: Success depends on writing good specs and prompts, not just having a good model.
- **Persistence Wins**: Keep trying until success. The loop handles retry logic automatically.

## Installation

### Download Binary (Recommended)

Download the latest release for your platform from the [Releases](https://github.com/stephenc/aider-ralph/releases) page.

> Note: Release assets are packaged as `.tar.gz` (macOS/Linux) or `.zip` (Windows). Download the archive, extract it, then move the binary into your PATH.

```bash
# macOS (Apple Silicon)
curl -L https://github.com/stephenc/aider-ralph/releases/latest/download/aider-ralph-darwin-arm64.tar.gz -o aider-ralph-darwin-arm64.tar.gz
tar -xzf aider-ralph-darwin-arm64.tar.gz
chmod +x aider-ralph-darwin-arm64
sudo mv aider-ralph-darwin-arm64 /usr/local/bin/aider-ralph

# macOS (Intel)
curl -L https://github.com/stephenc/aider-ralph/releases/latest/download/aider-ralph-darwin-amd64.tar.gz -o aider-ralph-darwin-amd64.tar.gz
tar -xzf aider-ralph-darwin-amd64.tar.gz
chmod +x aider-ralph-darwin-amd64
sudo mv aider-ralph-darwin-amd64 /usr/local/bin/aider-ralph

# Linux (x86_64)
curl -L https://github.com/stephenc/aider-ralph/releases/latest/download/aider-ralph-linux-amd64.tar.gz -o aider-ralph-linux-amd64.tar.gz
tar -xzf aider-ralph-linux-amd64.tar.gz
chmod +x aider-ralph-linux-amd64
sudo mv aider-ralph-linux-amd64 /usr/local/bin/aider-ralph

# Linux (arm64)
curl -L https://github.com/stephenc/aider-ralph/releases/latest/download/aider-ralph-linux-arm64.tar.gz -o aider-ralph-linux-arm64.tar.gz
tar -xzf aider-ralph-linux-arm64.tar.gz
chmod +x aider-ralph-linux-arm64
sudo mv aider-ralph-linux-arm64 /usr/local/bin/aider-ralph
```

### Build from Source

```bash
# Requires Go 1.22+
git clone https://github.com/stephenc/aider-ralph.git
cd aider-ralph
go build -o aider-ralph .
sudo mv aider-ralph /usr/local/bin/
```

### Requirements

- [Aider](https://aider.chat/) installed (`pip install aider-chat`)
- An API key for your preferred LLM provider

## Quick Start

### 1) Initialize a project (recommended)

```bash
aider-ralph --init "My Todo App"
```

This creates:
- `SPECS.md` (your requirements; re-read every iteration)
- `.ralph/notes.md` (notes forwarded between iterations)
- `.ralph/logs/` (optional logs directory)
- `.ralph/config` (informational defaults)
- `CONVENTIONS.md` (project-specific conventions/invariants, e.g. tests/linters/coverage expectations)

### 2) Edit your specs

```bash
vim SPECS.md
```

Recommended formats:
- **Markdown**: use checkboxes `- [ ]` and mark done as `- [x]`
- **JSON**: use objects with a boolean `completed` field

### 3) Run the loop

```bash
aider-ralph -m 30 -- --model sonnet --yes
```

By default, aider-ralph will:
- Load specs from `SPECS.md` (each iteration)
- Use `PROMPT.md` as a prompt template if it exists; otherwise use a built-in default template embedded in the binary
- Look for the completion signal `<ralph_status>COMPLETED</ralph_status>` in aider output

## Key Concepts

### Specs (what to build)

By default, aider-ralph assumes your specs live in `SPECS.md` and will re-read it every iteration (so the AI can modify it and the next iteration will see the changes).

You can specify a different specs file:

```bash
aider-ralph -s path/to/SPECS.md -m 30 -- --model sonnet --yes
```

### Prompt template (how to work)

Each iteration, aider-ralph assembles the message to aider from:
1. A **prompt template** (instructions / methodology)
2. The **specs** (from `SPECS.md` by default)
3. Any **notes from previous iterations** (from `.ralph/notes.md` if present)

Prompt template selection:
- If you pass a direct prompt argument, it is used as-is.
- Else if `-f/--file` is provided, that file is used as the template.
- Else if `PROMPT.md` exists, it is used automatically.
- Otherwise, a built-in default template (embedded in the binary) is used.

### Notes forwarded between iterations

aider-ralph supports forwarding context to the next iteration via a notes file.

- Default behavior: if `.ralph/notes.md` exists, it will be used automatically.
- You can override the notes file path:

```bash
aider-ralph --notes-file .ralph/my-notes.md -m 30 -- --model sonnet --yes
```

To write notes for the next iteration, have the model include a block like:

```text
<ralph_notes>
- What changed
- What to do next
- Any blockers
</ralph_notes>
```

aider-ralph will extract the last `<ralph_notes>...</ralph_notes>` block from aider output and append it to the notes file, then include the notes in the next iteration’s prompt.

### Conventions (project invariants)

Projects can optionally include a `CONVENTIONS.md` file containing project-specific conventions/invariants (for example: “run tests”, “run linters”, “keep coverage above 75%”, etc.).

When present, the model should follow `CONVENTIONS.md` strictly.

### Completion promise (termination condition)

By default, aider-ralph stops when it detects this XML-like completion tag in aider’s output:

```text
<ralph_status>
COMPLETED
</ralph_status>
```

This is intentionally more specific than plain `COMPLETED` to reduce accidental matches.

You can override the tag/value:

```bash
aider-ralph --completion-tag promise --completion-value DONE -m 30 -- --model sonnet --yes
```

Legacy option (substring match) is also supported:

```bash
aider-ralph -c "DONE" -m 30 -- --model sonnet --yes
```

## Usage

```text
aider-ralph --init [PROJECT_NAME]
aider-ralph [OPTIONS] "<prompt>" [-- AIDER_OPTIONS]
aider-ralph [OPTIONS] -f PROMPT_FILE [-- AIDER_OPTIONS]
```

### Commands

| Command | Description |
|--------|-------------|
| `--init [NAME]` | Initialize project with `SPECS.md` and `.ralph/` directory |

### Options

| Option | Description |
|--------|-------------|
| `-m, --max-iterations <N>` | Stop after N iterations (default: 30). Set to `0` for unlimited (not recommended). |
| `-s, --specs <PATH>` | Specs file to load each iteration (default: `SPECS.md`) |
| `-f, --file <PATH>` | Prompt template file (default: `PROMPT.md` if present, else embedded template) |
| `--notes-file <PATH>` | Notes file forwarded between iterations (default: `.ralph/notes.md` if present) |
| `--completion-tag <TAG>` | Completion tag name (default: `ralph_status`) |
| `--completion-value <VALUE>` | Completion tag value (default: `COMPLETED`) |
| `-c, --completion-promise <TEXT>` | Legacy completion detection (substring match) |
| `-d, --delay <SECONDS>` | Delay between iterations (default: 2) |
| `-t, --timeout <SECONDS>` | Timeout per iteration (default: 900 / 15min) |
| `-l, --log <PATH>` | Log all output to file |
| `-v, --verbose` | Show detailed progress information |
| `--dry-run` | Show what would be executed without running |
| `--version` | Show version information |
| `-h, --help` | Show help message |

### Aider Options

Any options after `--` are passed directly to aider:

```bash
aider-ralph -s SPECS.md -m 20 -- --model sonnet --api-key anthropic=sk-xxx --yes
```

Common aider options:
- `--model <MODEL>` — LLM model (sonnet, gpt-4o, deepseek, etc.)
- `--api-key <PROVIDER>=<KEY>` — API key
- `--yes` — Auto-confirm all prompts
- `--no-git` — Disable git integration

## The Basic Loop

At its core, aider-ralph implements this loop concept:

```bash
while :; do aider --message "$(cat PROMPT.md)" --yes; done
```

The Go binary adds:
- **Iteration limits** — Safety net with `-m` flag
- **Timeout protection** — Kills hung processes (default 15min)
- **Completion detection** — Auto-stop when completion is detected
- **Logging** — Full output capture for review
- **Notes forwarding** — Carry context between iterations
- **Rescanning** — Re-reads template/specs/notes every iteration
- **Cross-platform** — Single binary for macOS/Linux/Windows

## Resources

- [Ralph Wiggum Technique](https://awesomeclaude.ai/ralph-wiggum) — Full documentation
- [Aider](https://aider.chat/) — AI pair programming tool
- [Geoffrey Huntley's Blog](https://ghuntley.com/ralph/) — Original technique creator

## License

MIT
