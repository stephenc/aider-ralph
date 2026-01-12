package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Version information
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[1;33m"
	colorBlue   = "\033[0;34m"
	colorPurple = "\033[0;35m"
	colorCyan   = "\033[0;36m"
	colorBold   = "\033[1m"
)

// Config holds the application configuration
type Config struct {
	MaxIterations     int
	CompletionPromise string
	Prompt            string
	PromptFile        string
	Delay             int
	Timeout           int // Timeout per iteration in seconds (0 = no timeout)
	LogFile           string
	Verbose           bool
	DryRun            bool
	AiderOpts         []string
	DoInit            bool
	ProjectName       string
	ShowVersion       bool
}

var config Config
var loopActive bool

func main() {
	parseArgs()

	if config.ShowVersion {
		fmt.Printf("aider-ralph %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	printBanner()

	if config.DoInit {
		initProject()
		os.Exit(0)
	}

	if err := validate(); err != nil {
		logError(err.Error())
		os.Exit(1)
	}

	// Setup signal handling
	setupSignalHandler()

	// Show configuration
	showConfig()

	// Run the main loop
	mainLoop()
}

func parseArgs() {
	// Manual argument parsing to allow flags in any order
	args := os.Args[1:]

	// First, find and extract aider options after --
	for i, arg := range args {
		if arg == "--" {
			config.AiderOpts = args[i+1:]
			args = args[:i]
			break
		}
	}

	// Parse remaining args manually
	i := 0
	var positionalArgs []string

	for i < len(args) {
		arg := args[i]

		switch arg {
		case "-m", "--max-iterations":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &config.MaxIterations)
				i += 2
			} else {
				i++
			}
		case "-c", "--completion-promise":
			if i+1 < len(args) {
				config.CompletionPromise = args[i+1]
				i += 2
			} else {
				i++
			}
		case "-f", "--file":
			if i+1 < len(args) {
				config.PromptFile = args[i+1]
				i += 2
			} else {
				i++
			}
		case "-d", "--delay":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &config.Delay)
				i += 2
			} else {
				i++
			}
		case "-l", "--log":
			if i+1 < len(args) {
				config.LogFile = args[i+1]
				i += 2
			} else {
				i++
			}
		case "-t", "--timeout":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &config.Timeout)
				i += 2
			} else {
				i++
			}
		case "-v", "--verbose":
			config.Verbose = true
			i++
		case "--dry-run":
			config.DryRun = true
			i++
		case "--init":
			config.DoInit = true
			i++
		case "--version":
			config.ShowVersion = true
			i++
		case "-h", "--help":
			usage()
			os.Exit(0)
		default:
			// Check for -m=N or --max-iterations=N style
			if strings.HasPrefix(arg, "-m=") {
				fmt.Sscanf(arg[3:], "%d", &config.MaxIterations)
			} else if strings.HasPrefix(arg, "--max-iterations=") {
				fmt.Sscanf(arg[17:], "%d", &config.MaxIterations)
			} else if strings.HasPrefix(arg, "-c=") {
				config.CompletionPromise = arg[3:]
			} else if strings.HasPrefix(arg, "--completion-promise=") {
				config.CompletionPromise = arg[21:]
			} else if strings.HasPrefix(arg, "-f=") {
				config.PromptFile = arg[3:]
			} else if strings.HasPrefix(arg, "--file=") {
				config.PromptFile = arg[7:]
			} else if strings.HasPrefix(arg, "-d=") {
				fmt.Sscanf(arg[3:], "%d", &config.Delay)
			} else if strings.HasPrefix(arg, "--delay=") {
				fmt.Sscanf(arg[8:], "%d", &config.Delay)
			} else if strings.HasPrefix(arg, "-l=") {
				config.LogFile = arg[3:]
			} else if strings.HasPrefix(arg, "--log=") {
				config.LogFile = arg[6:]
			} else if strings.HasPrefix(arg, "-t=") {
				fmt.Sscanf(arg[3:], "%d", &config.Timeout)
			} else if strings.HasPrefix(arg, "--timeout=") {
				fmt.Sscanf(arg[10:], "%d", &config.Timeout)
			} else if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "%sUnknown option: %s%s\n", colorRed, arg, colorReset)
				usage()
				os.Exit(1)
			} else {
				positionalArgs = append(positionalArgs, arg)
			}
			i++
		}
	}

	// Set default delay if not specified
	if config.Delay == 0 {
		config.Delay = 2
	}

	// Set default timeout (15 minutes = 900 seconds)
	if config.Timeout == 0 {
		config.Timeout = 900
	}

	// First positional argument is either project name (for init) or prompt
	if len(positionalArgs) > 0 {
		if config.DoInit {
			config.ProjectName = positionalArgs[0]
		} else {
			config.Prompt = positionalArgs[0]
		}
	}
}

func usage() {
	fmt.Print(`aider-ralph - Ralph Wiggum AI Loop Technique for Aider

USAGE:
    aider-ralph --init [PROJECT_NAME]
    aider-ralph [OPTIONS] "<prompt>" [-- AIDER_OPTIONS]
    aider-ralph [OPTIONS] -f PROMPT_FILE [-- AIDER_OPTIONS]

COMMANDS:
    --init [NAME]                Initialize project for aider-ralph
                                 Creates SPECS.md and .ralph/ directory

OPTIONS:
    -m, --max-iterations <N>     Stop after N iterations (default: unlimited)
                                 STRONGLY RECOMMENDED as a safety net

    -c, --completion-promise <TEXT>
                                 Phrase that signals completion (checked in output)
                                 e.g., "COMPLETE", "DONE", "FINISHED"

    -f, --file <PATH>            Read prompt from file instead of argument
                                 File is re-read each iteration (live updates)

    -d, --delay <SECONDS>        Delay between iterations (default: 2)

    -l, --log <PATH>             Log all output to file

    -t, --timeout <SECONDS>      Timeout per iteration (default: 900 / 15min)
                                 Kills aider if it hangs

    -v, --verbose                Show detailed progress information

    --dry-run                    Show what would be executed without running

    --version                    Show version information

    -h                           Show this help message

AIDER OPTIONS:
    Any options after -- are passed directly to aider.
    Common aider options:
        --model <MODEL>          LLM model to use (sonnet, gpt-4o, etc.)
                                 For Ollama: --model ollama/llama2 or --model ollama/codellama
        --api-key <PROVIDER>=<KEY>  API key for the provider
        --yes                    Auto-confirm all prompts
        --no-git                 Disable git integration

EXAMPLES:
    # Initialize a new project
    aider-ralph --init "My Todo App"

    # Simple loop with iteration limit
    aider-ralph "Build a REST API for todos" -m 10

    # With completion promise detection
    aider-ralph "Implement auth. Output DONE when complete." -c "DONE" -m 20

    # Using the specs file created by --init
    aider-ralph -f SPECS.md -m 30 -c COMPLETE -- --model sonnet --yes

    # Using Ollama with a local model
    aider-ralph -f SPECS.md -m 30 -c COMPLETE -- --model ollama/llama2 --yes

    # Using Ollama with CodeLlama
    aider-ralph -f SPECS.md -m 30 -c COMPLETE -- --model ollama/codellama --yes

More info: https://awesomeclaude.ai/ralph-wiggum
`)
}

func printBanner() {
	fmt.Print(colorCyan + colorBold)
	fmt.Println(`
    ____        __      __       ____        __      __
   / __ \____ _/ /___  / /_     / __ \____ _/ /___  / /_
  / /_/ / __ ` + "`" + `/ / __ \/ __ \   / /_/ / __ ` + "`" + `/ / __ \/ __ \
 / _, _/ /_/ / / /_/ / / / /  / _, _/ /_/ / / /_/ / / / /
/_/ |_|\__,_/_/ .___/_/ /_/  /_/ |_|\__,_/_/ .___/_/ /_/
             /_/                          /_/`)
	fmt.Println(colorReset)
	fmt.Println(colorYellow + "Ralph Wiggum AI Loop Technique for Aider" + colorReset)
	fmt.Println(colorCyan + `"I'm learnding!" - Ralph Wiggum` + colorReset)
	fmt.Println()
}

func timestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func logInfo(msg string) {
	fmt.Printf("%s[%s]%s %s\n", colorBlue, timestamp(), colorReset, msg)
}

func logOK(msg string) {
	fmt.Printf("%s[%s]%s ✅ %s\n", colorGreen, timestamp(), colorReset, msg)
}

func logWarn(msg string) {
	fmt.Printf("%s[%s]%s ⚠️  %s\n", colorYellow, timestamp(), colorReset, msg)
}

func logError(msg string) {
	fmt.Printf("%s[%s]%s ❌ %s\n", colorRed, timestamp(), colorReset, msg)
}

func logIter(msg string) {
	fmt.Printf("%s[%s]%s 🔄 %s\n", colorPurple, timestamp(), colorReset, msg)
}

func validate() error {
	// Check if aider is installed
	if _, err := exec.LookPath("aider"); err != nil {
		return fmt.Errorf("aider is not installed. Install with: pip install aider-chat")
	}

	// Must have either prompt or prompt file
	if config.Prompt == "" && config.PromptFile == "" {
		fmt.Println()
		usage()
		return fmt.Errorf("no prompt provided. Use a prompt argument or -f <file>")
	}

	// If file specified, check it exists
	if config.PromptFile != "" {
		if _, err := os.Stat(config.PromptFile); os.IsNotExist(err) {
			return fmt.Errorf("prompt file not found: %s", config.PromptFile)
		}
	}

	// Warn if no max iterations
	if config.MaxIterations == 0 {
		logWarn("No --max-iterations set. Loop will run indefinitely!")
		logWarn("Press Ctrl+C to stop, or set -m for safety")
		fmt.Println()
	}

	return nil
}

func showConfig() {
	logInfo("Configuration:")
	if config.PromptFile != "" {
		fmt.Printf("  %sPrompt file:%s %s\n", colorCyan, colorReset, config.PromptFile)
	} else {
		prompt := config.Prompt
		if len(prompt) > 50 {
			prompt = prompt[:50] + "..."
		}
		fmt.Printf("  %sPrompt:%s %s\n", colorCyan, colorReset, prompt)
	}

	if config.MaxIterations > 0 {
		fmt.Printf("  %sMax iterations:%s %d\n", colorCyan, colorReset, config.MaxIterations)
	} else {
		fmt.Printf("  %sMax iterations:%s unlimited\n", colorCyan, colorReset)
	}

	if config.CompletionPromise != "" {
		fmt.Printf("  %sCompletion promise:%s %s\n", colorCyan, colorReset, config.CompletionPromise)
	}

	fmt.Printf("  %sTimeout:%s %ds\n", colorCyan, colorReset, config.Timeout)

	if len(config.AiderOpts) > 0 {
		fmt.Printf("  %sAider options:%s %s\n", colorCyan, colorReset, strings.Join(config.AiderOpts, " "))
	}

	if config.LogFile != "" {
		fmt.Printf("  %sLog file:%s %s\n", colorCyan, colorReset, config.LogFile)
	}

	fmt.Println()
}

func getPrompt() (string, error) {
	if config.PromptFile != "" {
		data, err := os.ReadFile(config.PromptFile)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	return config.Prompt, nil
}

func checkCompletion(output string) bool {
	if config.CompletionPromise == "" {
		return false
	}
	return strings.Contains(output, config.CompletionPromise)
}

func runIteration(iteration int, logWriter io.Writer) bool {
	logIter(fmt.Sprintf("Iteration %d starting...", iteration))

	prompt, err := getPrompt()
	if err != nil {
		logError(fmt.Sprintf("Failed to read prompt: %v", err))
		return false
	}

	if config.Verbose {
		fmt.Printf("%s--- Prompt ---%s\n", colorCyan, colorReset)
		lines := strings.Split(prompt, "\n")
		for i, line := range lines {
			if i >= 20 {
				fmt.Println("... (truncated)")
				break
			}
			fmt.Println(line)
		}
		fmt.Printf("%s--------------%s\n", colorCyan, colorReset)
	}

	// Build aider command
	args := []string{"--message", prompt, "--yes"}
	args = append(args, config.AiderOpts...)

	if config.Verbose {
		logInfo(fmt.Sprintf("Running: aider %s", strings.Join(args, " ")))
	}

	if config.DryRun {
		logInfo(fmt.Sprintf("[DRY RUN] Would execute: aider %s", strings.Join(args, " ")))
		return false // Continue loop in dry run
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
	defer cancel()

	// Run aider with context
	cmd := exec.CommandContext(ctx, "aider", args...)
	cmd.Stdin = os.Stdin

	// Capture output while also displaying it
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logError(fmt.Sprintf("Failed to create stdout pipe: %v", err))
		return false
	}

	cmd.Stderr = cmd.Stdout // Combine stderr with stdout

	if err := cmd.Start(); err != nil {
		logError(fmt.Sprintf("Failed to start aider: %v", err))
		return false
	}

	// Read output line by line
	var outputBuilder strings.Builder
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
		outputBuilder.WriteString(line)
		outputBuilder.WriteString("\n")

		if logWriter != nil {
			fmt.Fprintln(logWriter, line)
		}
	}

	err = cmd.Wait()
	output := outputBuilder.String()

	// Check if killed due to timeout
	if ctx.Err() == context.DeadlineExceeded {
		logWarn(fmt.Sprintf("Iteration timed out after %ds - aider was killed", config.Timeout))
		return false
	}

	// Log iteration to file
	if logWriter != nil {
		fmt.Fprintf(logWriter, "\n=== End of Iteration %d ===\n\n", iteration)
	}

	// Check for completion
	if checkCompletion(output) {
		logOK(fmt.Sprintf("Completion promise '%s' detected!", config.CompletionPromise))
		return true
	}

	return false
}

func mainLoop() {
	loopActive = true
	currentIteration := 0

	// Open log file if specified
	var logWriter io.Writer
	if config.LogFile != "" {
		f, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			logError(fmt.Sprintf("Failed to open log file: %v", err))
		} else {
			defer f.Close()
			logWriter = f
			fmt.Fprintf(f, "=== aider-ralph session started at %s ===\n\n", timestamp())
		}
	}

	for loopActive {
		currentIteration++

		// Check max iterations
		if config.MaxIterations > 0 && currentIteration > config.MaxIterations {
			logWarn(fmt.Sprintf("Max iterations (%d) reached", config.MaxIterations))
			break
		}

		// Show progress
		fmt.Println()
		fmt.Printf("%s═══════════════════════════════════════════════════════════%s\n", colorBold, colorReset)
		if config.MaxIterations > 0 {
			fmt.Printf("%s  ITERATION %d / %d%s\n", colorPurple, currentIteration, config.MaxIterations, colorReset)
		} else {
			fmt.Printf("%s  ITERATION %d (unlimited)%s\n", colorPurple, currentIteration, colorReset)
		}
		fmt.Printf("%s═══════════════════════════════════════════════════════════%s\n", colorBold, colorReset)
		fmt.Println()

		if logWriter != nil {
			fmt.Fprintf(logWriter, "=== Iteration %d ===\n", currentIteration)
		}

		// Run iteration
		if runIteration(currentIteration, logWriter) {
			logOK(fmt.Sprintf("Loop completed successfully after %d iteration(s)!", currentIteration))
			loopActive = false
			break
		}

		// Delay between iterations
		if loopActive && config.Delay > 0 {
			logInfo(fmt.Sprintf("Waiting %ds before next iteration...", config.Delay))
			time.Sleep(time.Duration(config.Delay) * time.Second)
		}
	}

	fmt.Println()
	logInfo(fmt.Sprintf("Ralph loop finished. Total iterations: %d", currentIteration))

	if config.LogFile != "" {
		fmt.Printf("\n%s📋 Log saved to: %s%s\n", colorCyan, config.LogFile, colorReset)
	}
}

func setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println()
		logWarn("Interrupted by user (Ctrl+C)")
		loopActive = false
		os.Exit(130)
	}()
}

func initProject() {
	projectName := config.ProjectName
	if projectName == "" {
		// Use current directory name
		cwd, err := os.Getwd()
		if err != nil {
			projectName = "My Project"
		} else {
			projectName = filepath.Base(cwd)
		}
	}

	specsFile := "SPECS.md"
	ralphDir := ".ralph"
	configFile := filepath.Join(ralphDir, "config")
	logsDir := filepath.Join(ralphDir, "logs")

	fmt.Printf("%sInitializing aider-ralph project: %s%s%s\n\n", colorCyan, colorBold, projectName, colorReset)

	// Create SPECS.md
	if _, err := os.Stat(specsFile); err == nil {
		fmt.Printf("%s⚠️  %s already exists. Skipping...%s\n", colorYellow, specsFile, colorReset)
	} else {
		specsContent := fmt.Sprintf(`# %s

## Project Overview
<!-- Describe what this project does in 2-3 sentences -->


## Goals
<!-- What are you trying to build? Be specific. -->

1.
2.
3.

## Technical Requirements
<!-- List specific technical requirements -->

- [ ]
- [ ]
- [ ]

## Architecture & Design
<!-- Describe the high-level architecture -->


## Implementation Phases
<!-- Break the work into phases. Each phase should be completable in one Ralph loop session -->

### Phase 1: Foundation
<!-- Start with the basics -->

**Requirements:**
- [ ]
- [ ]

**Success Criteria:**
- All requirements implemented
- Tests passing

**When Phase 1 is complete, output:** `+"`PHASE1_COMPLETE`"+`

### Phase 2: Core Features
<!-- Build the main functionality -->

**Requirements:**
- [ ]
- [ ]

**Success Criteria:**
- All requirements implemented
- Tests passing

**When Phase 2 is complete, output:** `+"`PHASE2_COMPLETE`"+`

### Phase 3: Polish & Testing
<!-- Final touches -->

**Requirements:**
- [ ]
- [ ]

**Success Criteria:**
- All requirements implemented
- All tests passing
- Code reviewed and clean

**When Phase 3 is complete, output:** `+"`COMPLETE`"+`

## Development Process
<!-- Instructions for aider on how to work -->

1. Read and understand the current phase requirements
2. Implement one requirement at a time
3. Write tests for each requirement
4. Run tests after each change
5. If tests fail, debug and fix before moving on
6. Mark requirements as done [x] when complete
7. Move to next phase when all requirements are complete

## Commands
<!-- Useful commands for the project -->

`+"```bash"+`
# Run tests
npm test  # or pytest, cargo test, etc.

# Start development server
npm run dev

# Build
npm run build

# Lint
npm run lint
`+"```"+`

## Notes
<!-- Any additional context for the AI -->


---

## Completion Signal

When ALL phases are complete and the project is ready:

**COMPLETE**

If stuck for more than 10 iterations on the same issue:
- Document what's blocking progress
- List attempted solutions
- Suggest what information or help is needed
- Output: **NEEDS_HUMAN_REVIEW**
`, projectName)

		if err := os.WriteFile(specsFile, []byte(specsContent), 0644); err != nil {
			logError(fmt.Sprintf("Failed to create %s: %v", specsFile, err))
		} else {
			fmt.Printf("%s✅ Created %s%s\n", colorGreen, specsFile, colorReset)
		}
	}

	// Create .ralph directory
	if _, err := os.Stat(ralphDir); err == nil {
		fmt.Printf("%s⚠️  %s/ already exists. Skipping...%s\n", colorYellow, ralphDir, colorReset)
	} else {
		if err := os.MkdirAll(logsDir, 0755); err != nil {
			logError(fmt.Sprintf("Failed to create %s: %v", ralphDir, err))
		} else {
			fmt.Printf("%s✅ Created %s/ directory%s\n", colorGreen, ralphDir, colorReset)
		}
	}

	// Create config file
	if _, err := os.Stat(configFile); err == nil {
		fmt.Printf("%s⚠️  %s already exists. Skipping...%s\n", colorYellow, configFile, colorReset)
	} else {
		configContent := `# aider-ralph configuration
# These are default settings that can be overridden via command line

# Maximum iterations before stopping (safety net)
MAX_ITERATIONS=30

# Phrase that signals completion
COMPLETION_PROMISE=COMPLETE

# Delay between iterations in seconds
ITERATION_DELAY=2

# Default specs file
SPECS_FILE=SPECS.md

# Aider model (uncomment to set default)
# AIDER_MODEL=sonnet

# Aider options (space-separated)
# AIDER_EXTRA_OPTS=--yes
`
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			logError(fmt.Sprintf("Failed to create %s: %v", configFile, err))
		} else {
			fmt.Printf("%s✅ Created %s%s\n", colorGreen, configFile, colorReset)
		}
	}

	// Add to .gitignore if it exists
	gitignorePath := ".gitignore"
	if _, err := os.Stat(gitignorePath); err == nil {
		content, err := os.ReadFile(gitignorePath)
		if err == nil && !strings.Contains(string(content), ".ralph/logs") {
			f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
			if err == nil {
				defer f.Close()
				f.WriteString("\n# aider-ralph logs\n.ralph/logs/\n")
				fmt.Printf("%s✅ Added .ralph/logs/ to .gitignore%s\n", colorGreen, colorReset)
			}
		}
	}

	fmt.Println()
	fmt.Printf("%sProject initialized!%s\n", colorBold, colorReset)
	fmt.Println()
	fmt.Printf("%sNext steps:%s\n", colorCyan, colorReset)
	fmt.Printf("  1. Edit %sSPECS.md%s with your project requirements\n", colorBold, colorReset)
	fmt.Printf("  2. Optionally edit %s.ralph/config%s for default settings\n", colorBold, colorReset)
	fmt.Printf("  3. Run: %saider-ralph -f SPECS.md -m 30 -c COMPLETE -- --model sonnet%s\n", colorBold, colorReset)
	fmt.Println()
	fmt.Printf("%sTips:%s\n", colorCyan, colorReset)
	fmt.Println("  • Break work into small, verifiable phases")
	fmt.Println("  • Include test commands so aider can verify its work")
	fmt.Println("  • Use clear completion signals (PHASE1_COMPLETE, COMPLETE, etc.)")
	fmt.Println("  • Set realistic max-iterations as a safety net")
	fmt.Println()
}
