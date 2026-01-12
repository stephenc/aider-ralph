package main

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
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

//go:embed PROMPT.md
var embeddedPromptTemplate string

//go:embed templates/CONVENTIONS.md
var embeddedConventionsTemplate string

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
	CompletionPromise string // legacy/simple substring completion
	CompletionTag     string // XML-like tag name for completion promise, e.g. "promise"
	CompletionValue   string // value inside the tag, e.g. "COMPLETED"

	Prompt     string
	PromptFile string

	SpecsFile string

	NotesFile string

	Delay   int
	Timeout int // Timeout per iteration in seconds (0 = no timeout)

	LogFile     string
	Verbose     bool
	DryRun      bool
	AiderOpts   []string
	DoInit      bool
	ProjectName string
	ShowVersion bool
}

var config Config
var loopActive bool

const defaultSpecsFile = "SPECS.md"
const defaultPromptFile = "PROMPT.md"
const defaultCompletionTag = "ralph_status"
const defaultCompletionValue = "COMPLETED"
const defaultMaxIterations = 30

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

	// Defaults
	config.SpecsFile = defaultSpecsFile
	config.CompletionTag = defaultCompletionTag
	config.CompletionValue = defaultCompletionValue
	config.MaxIterations = defaultMaxIterations

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
		case "--completion-tag":
			if i+1 < len(args) {
				config.CompletionTag = args[i+1]
				i += 2
			} else {
				i++
			}
		case "--completion-value":
			if i+1 < len(args) {
				config.CompletionValue = args[i+1]
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
		case "-s", "--specs":
			if i+1 < len(args) {
				config.SpecsFile = args[i+1]
				i += 2
			} else {
				i++
			}
		case "--notes-file":
			if i+1 < len(args) {
				config.NotesFile = args[i+1]
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
			} else if strings.HasPrefix(arg, "--completion-tag=") {
				config.CompletionTag = arg[len("--completion-tag="):]
			} else if strings.HasPrefix(arg, "--completion-value=") {
				config.CompletionValue = arg[len("--completion-value="):]
			} else if strings.HasPrefix(arg, "-f=") {
				config.PromptFile = arg[3:]
			} else if strings.HasPrefix(arg, "--file=") {
				config.PromptFile = arg[7:]
			} else if strings.HasPrefix(arg, "-s=") {
				config.SpecsFile = arg[3:]
			} else if strings.HasPrefix(arg, "--specs=") {
				config.SpecsFile = arg[8:]
			} else if strings.HasPrefix(arg, "--notes-file=") {
				config.NotesFile = arg[len("--notes-file="):]
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

COMMON (RECOMMENDED):
    aider-ralph -s SPECS.md -m 30 -- --model sonnet --yes

OPTIONS:
    -m, --max-iterations <N>     Stop after N iterations (default: 30)
                                 Set to 0 for unlimited (not recommended)

    -s, --specs <PATH>           Specs file to load each iteration (default: SPECS.md)
                                 If missing and no prompt is provided, help is shown.

    -f, --file <PATH>            Read prompt template from file instead of argument
                                 File is re-read each iteration (live updates)
                                 If not provided, PROMPT.md is used if present.

    -c, --completion-promise <TEXT>
                                 Legacy completion detection: substring match in output

    --completion-tag <TAG>       Safer completion detection using an XML-like tag
                                 default: ralph_status

    --completion-value <VALUE>   Value inside the completion tag
                                 default: COMPLETED
                                 Example: <ralph_status>COMPLETED</ralph_status>

    --notes-file <PATH>          File to store iteration notes and feed into next iteration
                                 If not set, .ralph/notes.md is used if it exists.

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

EXAMPLES:
    # Initialize a new project
    aider-ralph --init "My Todo App"

    # Use default SPECS.md and PROMPT.md (if present)
    aider-ralph -m 30 -- --model sonnet --yes

    # Unlimited iterations (not recommended)
    aider-ralph -m 0 -- --model sonnet --yes

    # Explicit specs and completion tag
    aider-ralph -s SPECS.md -m 30 --completion-tag promise --completion-value COMPLETED -- --model sonnet --yes
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
	fmt.Println(colorYellow + "Hey — welcome to aider-ralph (the Ralph Wiggum loop for Aider)." + colorReset)
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

	// If no prompt and no prompt file, we will try PROMPT.md if present, else default template.
	// But we still require a specs file to exist (or a direct prompt argument) to avoid running with nothing.
	if config.Prompt == "" && config.PromptFile == "" {
		if fileExists(defaultPromptFile) {
			config.PromptFile = defaultPromptFile
		}
	}

	// Default notes file behavior: if not specified, use .ralph/notes.md if it exists.
	if config.NotesFile == "" {
		defaultNotes := filepath.Join(".ralph", "notes.md")
		if fileExists(defaultNotes) {
			config.NotesFile = defaultNotes
		}
	}

	// If no prompt argument and no prompt file and no specs file, show help.
	// We treat specs as the primary input; prompt template can be defaulted.
	if config.Prompt == "" && config.PromptFile == "" && config.SpecsFile == "" {
		fmt.Println()
		usage()
		return fmt.Errorf("no prompt or specs provided")
	}

	// If specs file is set (default SPECS.md), require it to exist unless user provided a direct prompt argument.
	// This matches: "specs file should be assumed. showing help if no specs file or specs file not specified on the cli."
	// Here, specs is always assumed; if it doesn't exist and user didn't provide a direct prompt, we error with help.
	if config.SpecsFile != "" {
		if !fileExists(config.SpecsFile) && config.Prompt == "" {
			fmt.Println()
			usage()
			return fmt.Errorf("specs file not found: %s (provide -s <file> or pass a direct prompt)", config.SpecsFile)
		}
	}

	// If prompt file specified, check it exists
	if config.PromptFile != "" {
		if !fileExists(config.PromptFile) {
			return fmt.Errorf("prompt file not found: %s", config.PromptFile)
		}
	}

	// Warn if user explicitly selected unlimited iterations
	if config.MaxIterations == 0 {
		logWarn("Max iterations set to 0 (unlimited). Loop will run indefinitely!")
		logWarn("Press Ctrl+C to stop, or set -m for safety")
		fmt.Println()
	}

	return nil
}

func showConfig() {
	logInfo("Configuration:")

	if config.SpecsFile != "" {
		fmt.Printf("  %sSpecs file:%s %s\n", colorCyan, colorReset, config.SpecsFile)
	}

	if config.PromptFile != "" {
		fmt.Printf("  %sPrompt template file:%s %s\n", colorCyan, colorReset, config.PromptFile)
	} else if config.Prompt != "" {
		prompt := config.Prompt
		if len(prompt) > 50 {
			prompt = prompt[:50] + "..."
		}
		fmt.Printf("  %sPrompt (direct):%s %s\n", colorCyan, colorReset, prompt)
	} else {
		fmt.Printf("  %sPrompt template:%s (built-in default)\n", colorCyan, colorReset)
	}

	if config.MaxIterations > 0 {
		fmt.Printf("  %sMax iterations:%s %d\n", colorCyan, colorReset, config.MaxIterations)
	} else {
		fmt.Printf("  %sMax iterations:%s unlimited\n", colorCyan, colorReset)
	}

	if config.CompletionTag != "" && config.CompletionValue != "" {
		fmt.Printf("  %sCompletion tag:%s <%s>%s</%s>\n", colorCyan, colorReset, config.CompletionTag, config.CompletionValue, config.CompletionTag)
	}
	if config.CompletionPromise != "" {
		fmt.Printf("  %sCompletion promise (legacy):%s %s\n", colorCyan, colorReset, config.CompletionPromise)
	}

	fmt.Printf("  %sTimeout:%s %ds\n", colorCyan, colorReset, config.Timeout)

	if config.NotesFile != "" {
		fmt.Printf("  %sNotes file:%s %s\n", colorCyan, colorReset, config.NotesFile)
	}

	if len(config.AiderOpts) > 0 {
		fmt.Printf("  %sAider options:%s %s\n", colorCyan, colorReset, strings.Join(config.AiderOpts, " "))
	}

	if config.LogFile != "" {
		fmt.Printf("  %sLog file:%s %s\n", colorCyan, colorReset, config.LogFile)
	}

	fmt.Println()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func getSpecs() (string, error) {
	if config.SpecsFile == "" {
		return "", nil
	}
	data, err := os.ReadFile(config.SpecsFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func getNotes() (string, error) {
	if config.NotesFile == "" {
		return "", nil
	}
	data, err := os.ReadFile(config.NotesFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func getPromptTemplate() (string, error) {
	// Direct prompt overrides everything
	if config.Prompt != "" {
		return config.Prompt, nil
	}

	// Prompt file if specified (or defaulted to PROMPT.md)
	if config.PromptFile != "" {
		data, err := os.ReadFile(config.PromptFile)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	// Built-in default template
	return defaultPromptTemplate(), nil
}

func defaultPromptTemplate() string {
	// Uses the embedded PROMPT.md from the repository root
	return strings.TrimSpace(embeddedPromptTemplate)
}

func buildIterationPrompt() (string, error) {
	template, err := getPromptTemplate()
	if err != nil {
		return "", err
	}

	specs, err := getSpecs()
	if err != nil {
		return "", err
	}

	notes, err := getNotes()
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString(template)
	b.WriteString("\n\n")
	if config.SpecsFile != "" {
		b.WriteString("=== SPECS (reloaded each iteration) ===\n")
		if specs == "" {
			b.WriteString("(empty)\n")
		} else {
			b.WriteString(specs)
			if !strings.HasSuffix(specs, "\n") {
				b.WriteString("\n")
			}
		}
		b.WriteString("=== END SPECS ===\n\n")
	}

	if notes != "" {
		b.WriteString("=== PRIOR_NOTES (carry forward) ===\n")
		b.WriteString(notes)
		if !strings.HasSuffix(notes, "\n") {
			b.WriteString("\n")
		}
		b.WriteString("=== END PRIOR_NOTES ===\n\n")
	}

	return b.String(), nil
}

func checkCompletion(output string) bool {
	// Prefer tag-based completion (low collision)
	if config.CompletionTag != "" && config.CompletionValue != "" {
		tag := regexp.QuoteMeta(config.CompletionTag)
		val := regexp.QuoteMeta(config.CompletionValue)
		// Match multi-line format:
		// <ralph_status>
		// COMPLETED
		// </ralph_status>
		// Also matches single-line: <ralph_status>COMPLETED</ralph_status>
		re := regexp.MustCompile(fmt.Sprintf(`(?s)<%s>\s*%s\s*</%s>`, tag, val, tag))
		if re.FindStringIndex(output) != nil {
			return true
		}
	}

	// Legacy substring completion
	if config.CompletionPromise != "" {
		return strings.Contains(output, config.CompletionPromise)
	}

	return false
}

func extractRalphNotes(output string) string {
	// Extract the last <ralph_notes>...</ralph_notes> block if present.
	re := regexp.MustCompile(`(?is)<ralph_notes>\s*(.*?)\s*</ralph_notes>`)
	matches := re.FindAllStringSubmatch(output, -1)
	if len(matches) == 0 {
		return ""
	}
	last := matches[len(matches)-1]
	if len(last) < 2 {
		return ""
	}
	return strings.TrimSpace(last[1])
}

func appendNotes(iteration int, notes string) error {
	if config.NotesFile == "" || strings.TrimSpace(notes) == "" {
		return nil
	}

	dir := filepath.Dir(config.NotesFile)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(config.NotesFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	header := fmt.Sprintf("\n## Iteration %d (%s)\n\n", iteration, timestamp())
	if _, err := f.WriteString(header); err != nil {
		return err
	}
	if _, err := f.WriteString(notes); err != nil {
		return err
	}
	if !strings.HasSuffix(notes, "\n") {
		if _, err := f.WriteString("\n"); err != nil {
			return err
		}
	}
	return nil
}

func runIteration(iteration int, logWriter io.Writer) bool {
	logIter(fmt.Sprintf("Iteration %d starting...", iteration))

	prompt, err := buildIterationPrompt()
	if err != nil {
		logError(fmt.Sprintf("Failed to build prompt: %v", err))
		return false
	}

	if config.Verbose {
		fmt.Printf("%s--- Prompt (assembled) ---%s\n", colorCyan, colorReset)
		lines := strings.Split(prompt, "\n")
		for i, line := range lines {
			if i >= 40 {
				fmt.Println("... (truncated)")
				break
			}
			fmt.Println(line)
		}
		fmt.Printf("%s-------------------------%s\n", colorCyan, colorReset)
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

	_ = cmd.Wait()
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

	// Extract and persist notes for next iteration
	notes := extractRalphNotes(output)
	if notes != "" {
		if err := appendNotes(iteration, notes); err != nil {
			logWarn(fmt.Sprintf("Failed to append notes: %v", err))
		} else if config.Verbose {
			logInfo("Appended <ralph_notes> to notes file for next iteration")
		}
	}

	// Check for completion
	if checkCompletion(output) {
		if config.CompletionTag != "" && config.CompletionValue != "" {
			logOK(fmt.Sprintf("Completion tag '<%s>%s</%s>' detected!", config.CompletionTag, config.CompletionValue, config.CompletionTag))
		} else {
			logOK(fmt.Sprintf("Completion promise '%s' detected!", config.CompletionPromise))
		}
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

	specsFile := defaultSpecsFile
	ralphDir := ".ralph"
	configFile := filepath.Join(ralphDir, "config")
	logsDir := filepath.Join(ralphDir, "logs")
	notesFile := filepath.Join(ralphDir, "notes.md")
	conventionsFile := "CONVENTIONS.md"

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

- [ ] Goal 1
- [ ] Goal 2
- [ ] Goal 3

## Technical Requirements
<!-- List specific technical requirements -->

- [ ] Requirement 1
- [ ] Requirement 2
- [ ] Requirement 3

## Notes
<!-- Any additional context -->

---

## Completion Signal

When ALL requirements are complete and the project is ready, output exactly:

<ralph_status>
COMPLETED
</ralph_status>
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

	// Create notes file (empty starter)
	if _, err := os.Stat(notesFile); err == nil {
		fmt.Printf("%s⚠️  %s already exists. Skipping...%s\n", colorYellow, notesFile, colorReset)
	} else {
		_ = os.MkdirAll(filepath.Dir(notesFile), 0755)
		if err := os.WriteFile(notesFile, []byte("# Ralph Notes\n\n"), 0644); err != nil {
			logError(fmt.Sprintf("Failed to create %s: %v", notesFile, err))
		} else {
			fmt.Printf("%s✅ Created %s%s\n", colorGreen, notesFile, colorReset)
		}
	}

	// Create CONVENTIONS.md (project-specific invariants/conventions)
	if _, err := os.Stat(conventionsFile); err == nil {
		fmt.Printf("%s⚠️  %s already exists. Skipping...%s\n", colorYellow, conventionsFile, colorReset)
	} else {
		conventionsContent := strings.TrimSpace(embeddedConventionsTemplate) + "\n"
		conventionsContent = strings.ReplaceAll(conventionsContent, "{{PROJECT_NAME}}", projectName)

		if err := os.WriteFile(conventionsFile, []byte(conventionsContent), 0644); err != nil {
			logError(fmt.Sprintf("Failed to create %s: %v", conventionsFile, err))
		} else {
			fmt.Printf("%s✅ Created %s%s\n", colorGreen, conventionsFile, colorReset)
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

# Safer completion signal (recommended)
COMPLETION_TAG=promise
COMPLETION_VALUE=COMPLETED

# Legacy completion substring (optional)
# COMPLETION_PROMISE=COMPLETE

# Delay between iterations in seconds
ITERATION_DELAY=2

# Default specs file
SPECS_FILE=SPECS.md

# Default notes file
NOTES_FILE=.ralph/notes.md

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
	fmt.Printf("  1. Edit %s%s%s with your project requirements\n", colorBold, specsFile, colorReset)
	fmt.Printf("  2. Optionally edit %s%s%s for project-specific conventions\n", colorBold, conventionsFile, colorReset)
	fmt.Printf("  3. Optionally edit %s.ralph/config%s for default settings\n", colorBold, colorReset)
	fmt.Printf("  4. Run: %saider-ralph -s %s -m 30 -- --model sonnet --yes%s\n", colorBold, specsFile, colorReset)
	fmt.Println()
	fmt.Printf("%sTips:%s\n", colorCyan, colorReset)
	fmt.Println("  • Break work into small, verifiable phases")
	fmt.Println("  • Use checkbox specs (- [ ] / - [x]) or JSON completed booleans")
	fmt.Println("  • Use low-collision completion signals like <ralph_status>COMPLETED</ralph_status>")
	fmt.Println("  • Add <ralph_notes>...</ralph_notes> to carry context forward")
	fmt.Println("  • Set realistic max-iterations as a safety net")
	fmt.Println()
}
