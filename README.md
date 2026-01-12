# aider-ralph

The **Ralph Wiggum AI Loop Technique** for [Aider](https://aider.chat/).

An iterative AI development methodology that repeatedly feeds aider a prompt until completion. Named after The Simpsons character, it embodies the philosophy of persistent iteration despite setbacks.

> *"I'm learnding!"* — Ralph Wiggum

## Philosophy

- **Iteration > Perfection**: Don't aim for perfect on first try. Let the loop refine the work.
- **Failures Are Data**: Deterministically bad means failures are predictable and informative.
- **Operator Skill Matters**: Success depends on writing good prompts, not just having a good model.
- **Persistence Wins**: Keep trying until success. The loop handles retry logic automatically.

## Installation

### Download Binary (Recommended)

Download the latest release for your platform from the [Releases](https://github.com/YOUR_USERNAME/aider-ralph/releases) page.

```bash
# macOS (Apple Silicon)
curl -L https://github.com/YOUR_USERNAME/aider-ralph/releases/latest/download/aider-ralph-darwin-arm64 -o aider-ralph
chmod +x aider-ralph
sudo mv aider-ralph /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/YOUR_USERNAME/aider-ralph/releases/latest/download/aider-ralph-darwin-amd64 -o aider-ralph
chmod +x aider-ralph
sudo mv aider-ralph /usr/local/bin/

# Linux (x86_64)
curl -L https://github.com/YOUR_USERNAME/aider-ralph/releases/latest/download/aider-ralph-linux-amd64 -o aider-ralph
chmod +x aider-ralph
sudo mv aider-ralph /usr/local/bin/
```

### Build from Source

```bash
# Requires Go 1.21+
git clone https://github.com/YOUR_USERNAME/aider-ralph.git
cd aider-ralph
go build -o aider-ralph .
sudo mv aider-ralph /usr/local/bin/
```

### Requirements

- [Aider](https://aider.chat/) installed (`pip install aider-chat`)
- An API key for your preferred LLM provider

## Quick Start

```bash
# Initialize a new project
aider-ralph --init "My Todo App"

# Edit SPECS.md with your requirements
vim SPECS.md

# (Optional) Create a prompt template file
# If PROMPT.md exists, aider-ralph will use it automatically.
vim PROMPT.md

# Run the loop (defaults to SPECS.md)
aider-ralph -m 30 -- --model sonnet
```

## Key Concepts

### Specs (what to build)

By default, aider-ralph assumes your specs live in `SPECS.md` and will re-read it every iteration (so the AI can modify it and the next iteration will see the changes).

You can also pass specs inline as a positional argument:

```bash
aider-ralph "Build a REST API for todos" -m 10 -- --model sonnet
```

Or specify a different specs file:

```bash
aider-ralph -s path/to/SPECS.md -m 30 -- --model sonnet
```

> `-f/--file` is a legacy alias for `-s/--specs`.

### Prompt template (how to work)

Each iteration, aider-ralph builds the message to aider by combining:

1. A **prompt template** (instructions / methodology)
2. The **specs**
3. Any **notes from previous iterations**

If you do not specify a template:
- If `PROMPT.md` exists, it is used automatically.
- Otherwise, an internal default template is used.

You can explicitly set a template file:

```bash
aider-ralph -p PROMPT.md -s SPECS.md -m 30 -- --model sonnet
```

### Notes forwarded between iterations

aider-ralph supports forwarding context to the next iteration via a notes file (default: `.ralph/notes.md`).

If the model outputs a section titled **"Notes for next iteration"**, aider-ralph will append it to the notes file, and include those notes in the next iteration’s prompt.

You can override the notes file path:

```bash
aider-ralph --notes-file .ralph/my-notes.md -m 30 -- --model sonnet
```

### Completion promise (termination condition)

By default, aider-ralph stops when it detects this string in aider’s output:

```
<promise>COMPLETED</promise>
```

This is intentionally more specific than plain `COMPLETE` to reduce accidental matches.

You can override it:

```bash
aider-ralph -c "<promise>DONE</promise>" -m 30 -- --model sonnet
```

## Usage

```
aider-ralph --init [PROJECT_NAME]
aider-ralph [OPTIONS] "<specs>" [-- AIDER_OPTIONS]
aider-ralph [OPTIONS] -s SPECS_FILE [-- AIDER_OPTIONS]
aider-ralph [OPTIONS] -f SPECS_FILE [-- AIDER_OPTIONS]   (legacy alias for -s)
```

### Commands

| Command | Description |
|---------|-------------|
| `--init [NAME]` | Initialize project with SPECS.md and .ralph/ directory |

### Options

| Option | Description |
|--------|-------------|
| `-m, --max-iterations <N>` | Stop after N iterations (strongly recommended) |
| `-c, --completion-promise <TEXT>` | Phrase that signals completion (default: `<promise>COMPLETED</promise>`) |
| `-s, --specs <PATH>` | Specs file to load each iteration (default: `SPECS.md`) |
| `-f, --file <PATH>` | Legacy alias for `--specs` |
| `-p, --prompt-template <PATH>` | Prompt template file (default: `PROMPT.md` if present, else internal default) |
| `--notes-file <PATH>` | Notes file forwarded between iterations (default: `.ralph/notes.md`) |
| `-d, --delay <SECONDS>` | Delay between iterations (default: 2) |
| `-t, --timeout <SECONDS>` | Timeout per iteration (default: 900 / 15min) |
| `-l, --log <PATH>` | Log all output to file |
| `-v, --verbose` | Show detailed progress information |
| `--dry-run` | Show what would be executed without running |
| `--version` | Show version information |
| `-h` | Show help message |

### Aider Options

Any options after `--` are passed directly to aider:

```bash
aider-ralph -s SPECS.md -m 20 -- --model sonnet --api-key anthropic=sk-xxx
```

Common aider options:
- `--model <MODEL>` — LLM model (sonnet, gpt-4o, deepseek, etc.)
- `--api-key <PROVIDER>=<KEY>` — API key
- `--yes` — Auto-confirm all prompts
- `--no-git` — Disable git integration

## Examples

### Default (SPECS.md + PROMPT.md if present)

```bash
aider-ralph -m 30 -- --model sonnet --yes
```

### With Completion Detection Override

```bash
aider-ralph -s SPECS.md -m 30 -c "<promise>DONE</promise>" -- --model sonnet --yes
```

### Logging

```bash
aider-ralph -s SPECS.md -m 100 \
  --log .ralph/logs/overnight-$(date +%Y%m%d).log \
  -- --model sonnet --yes
```

## Project Initialization

Running `aider-ralph --init` creates:

```
your-project/
├── SPECS.md           # Main specs file (re-read each iteration)
└── .ralph/
    ├── config         # Default configuration (informational)
    ├── logs/          # Output logs directory
    └── notes.md       # Notes forwarded between iterations
```

## The Basic Loop

At its core, aider-ralph implements this loop concept:

```bash
while :; do aider --message "$(cat PROMPT.md)" --yes; done
```

The Go binary adds:
- **Iteration limits** — Safety net with `-m` flag
- **Timeout protection** — Kills hung processes (default 15min)
- **Completion detection** — Auto-stop when promise is detected
- **Logging** — Full output capture for review
- **Notes forwarding** — Carry context between iterations
- **Rescanning** — Re-reads template/specs/notes every iteration
- **Cross-platform** — Single binary for macOS/Linux

## Resources

- [Ralph Wiggum Technique](https://awesomeclaude.ai/ralph-wiggum) — Full documentation
- [Aider](https://aider.chat/) — AI pair programming tool
- [Geoffrey Huntley's Blog](https://ghuntley.com/ralph/) — Original technique creator

## License

MIT
