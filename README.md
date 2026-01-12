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

```bash
# Clone the repo
git clone https://github.com/YOUR_USERNAME/aider-ralph.git
cd aider-ralph

# Make executable
chmod +x aider-ralph

# Optionally, add to PATH
ln -s "$(pwd)/aider-ralph" /usr/local/bin/aider-ralph
```

### Requirements

- Bash 4.0+
- [Aider](https://aider.chat/) installed (`pip install aider-chat`)
- An API key for your preferred LLM provider

## Quick Start

```bash
# Initialize a new project
aider-ralph --init "My Todo App"

# Edit SPECS.md with your requirements
vim SPECS.md

# Run the loop
aider-ralph -f SPECS.md -m 30 -c "COMPLETE" -- --model sonnet
```

## Usage

```
aider-ralph --init [PROJECT_NAME]
aider-ralph [OPTIONS] "<prompt>" [-- AIDER_OPTIONS]
aider-ralph [OPTIONS] -f PROMPT_FILE [-- AIDER_OPTIONS]
```

### Commands

| Command | Description |
|---------|-------------|
| `--init [NAME]` | Initialize project with SPECS.md and .ralph/ directory |

### Options

| Option | Description |
|--------|-------------|
| `-m, --max-iterations <N>` | Stop after N iterations (strongly recommended) |
| `-c, --completion-promise <TEXT>` | Phrase that signals completion (e.g., "DONE") |
| `-f, --file <PATH>` | Read prompt from file (re-read each iteration) |
| `-d, --delay <SECONDS>` | Delay between iterations (default: 2) |
| `-l, --log <PATH>` | Log all output to file |
| `-v, --verbose` | Show detailed progress information |
| `--dry-run` | Show what would be executed without running |
| `-h, --help` | Show help message |

### Aider Options

Any options after `--` are passed directly to aider:

```bash
aider-ralph "Build an API" -m 20 -- --model sonnet --api-key anthropic=sk-xxx
```

Common aider options:
- `--model <MODEL>` — LLM model (sonnet, gpt-4o, deepseek, etc.)
- `--api-key <PROVIDER>=<KEY>` — API key
- `--yes` — Auto-confirm all prompts
- `--no-git` — Disable git integration

## Examples

### Simple Loop

```bash
aider-ralph "Build a REST API for todos" --max-iterations 10
```

### With Completion Detection

```bash
aider-ralph "Implement user auth. Output DONE when complete." \
    --completion-promise "DONE" \
    --max-iterations 20
```

### Using a Specs File

```bash
aider-ralph -f SPECS.md -m 30 -c "COMPLETE" -- --model sonnet --yes
```

### TDD Development Loop

```bash
aider-ralph -f tdd-prompt.md \
    --max-iterations 50 \
    --completion-promise "ALL_TESTS_PASS" \
    -- --model gpt-4o
```

### Overnight Batch Processing

```bash
# Run before bed
aider-ralph -f SPECS.md -m 100 -c "COMPLETE" \
    --log .ralph/logs/overnight-$(date +%Y%m%d).log \
    -- --model sonnet --yes
```

## Project Initialization

Running `aider-ralph --init` creates:

```
your-project/
├── SPECS.md           # Main specs/prompt file
└── .ralph/
    ├── config         # Default configuration
    └── logs/          # Output logs directory
```

### SPECS.md Template

The generated SPECS.md includes:

- **Project Overview** — What the project does
- **Goals** — Numbered list of objectives
- **Technical Requirements** — Checkbox list
- **Implementation Phases** — Broken into completable chunks with signals
- **Development Process** — Instructions for the AI
- **Commands** — Test/build/lint commands
- **Completion Signals** — Clear markers per phase

### Configuration

Edit `.ralph/config` to set defaults:

```bash
# Maximum iterations before stopping
MAX_ITERATIONS=30

# Phrase that signals completion
COMPLETION_PROMISE=COMPLETE

# Delay between iterations in seconds
ITERATION_DELAY=2

# Default specs file
SPECS_FILE=SPECS.md

# Aider model (uncomment to set)
# AIDER_MODEL=sonnet
```

## Writing Good Prompts

### 1. Clear Completion Criteria

```markdown
## Bad
Build a todo API and make it good.

## Good
Build a REST API for todos. When complete:
- All CRUD endpoints working
- Input validation in place
- Tests passing (coverage > 80%)
- README with API docs

Output: **COMPLETE**
```

### 2. Incremental Goals

```markdown
## Bad
Create a complete e-commerce platform.

## Good
Phase 1: User authentication (JWT, tests)
Phase 2: Product catalog (list/search, tests)
Phase 3: Shopping cart (add/remove, tests)

Output **COMPLETE** when all phases done.
```

### 3. Self-Correction Pattern

```markdown
## Bad
Write code for feature X.

## Good
Implement feature X following TDD:
1. Write failing tests
2. Implement feature
3. Run tests
4. If any fail, debug and fix
5. Refactor if needed
6. Repeat until all green
7. Output: **COMPLETE**
```

### 4. Escape Hatches

Always include what to do if stuck:

```markdown
If stuck for more than 10 iterations:
- Document what's blocking progress
- List attempted solutions
- Output: **NEEDS_HUMAN_REVIEW**
```

## The Basic Loop

At its core, aider-ralph is just a wrapper around this simple bash loop:

```bash
while :; do aider --message "$(cat PROMPT.md)" --yes; done
```

The tool adds:
- Iteration limits for safety
- Completion detection
- Logging and progress display
- Configuration management

## Real-World Results

The Ralph Wiggum technique has been used to:

- Generate **6 repositories overnight** at Y Combinator hackathon
- Complete a **$50k contract for $297** in API costs
- Create an entire **programming language (CURSED)** over 3 months

## Resources

- [Ralph Wiggum Technique](https://awesomeclaude.ai/ralph-wiggum) — Full documentation
- [Aider](https://aider.chat/) — AI pair programming tool
- [Geoffrey Huntley's Blog](https://ghuntley.com/ralph/) — Original technique creator

## License

MIT
