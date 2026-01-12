# Implementation of the Ralph Wiggum technique for Aider

## Project Overview

An iterative AI development methodology that repeatedly runs `aider` with a prompt until completion. Named after The Simpsons character, it embodies the philosophy of persistent iteration despite setbacks.

## Completion mechanism (how to mark progress)

This SPECS file uses a checkbox-based completion mechanism:

- For Markdown SPECS: each requirement MUST be a checkbox item using `- [ ]` (not done) and `- [x]` (done).
- If SPECS is ever converted to JSON: each requirement MUST be an object with a `completed` boolean field, and completion is tracked by setting `"completed": true`.

When a requirement is implemented, it should be marked complete by changing `- [ ]` to `- [x]` (or setting `"completed": true` in JSON).

## Requirements

### Core loop behavior
- [x] Implement the equivalent of this bash loop: `while :; do aider --message "$(cat PROMPT.md)" --yes; done` but with safety rails
- [x] Default max iterations to 30 rather than unbounded; let users specify 0 to select unbounded
- [x] Add timeout protection per iteration
- [x] Add optional delay between iterations
- [x] Add logging support (optional log file)

### Prompt/specs/notes inputs (reloaded every iteration)
- [x] Allow externalisation of the specs into a separate specs file so that a more standard prompt can be used
- [x] Use sensible defaults:
  - Specs file defaults to `SPECS.md`
  - Prompt template selection order: direct prompt arg > `-f/--file` > `PROMPT.md` if present > embedded default template
- [x] aider-ralph should always rescan files that it loads every iteration so that self-modification is permitted.
- [x] Update the instructions to allow and encourage forwarding notes to the next iteration
- [x] Support forwarding notes between iterations via `.ralph/notes.md` (append-only) using `<ralph_notes>...</ralph_notes>` blocks

### Completion / termination
- [x] Richer and more nuanced termination conditions using an XML-like completion tag/value to reduce accidental matches
- [x] Default completion detection uses:
  - Tag: `ralph_status`
  - Value: `COMPLETED`
  - Example:
    ```
    <ralph_status>
    COMPLETED
    </ralph_status>
    ```
- [x] Legacy completion detection via substring match is supported
- [x] Fix bug where it kept on looping even after all the specs were marked as done (handled via prompt guidance; do not assume a specific SPECS format beyond what the prompt instructs)

### Init mode / templates / conventions
- [x] The init mode should populate the current prompt template as `PROMPT.md` and create a seed `SPECS.md` as well as `.ralph/notes.md` for notes.
- [x] The default prompt template should indicate that `.ralph/notes.md` is append only and `SPECS.md` is only to be updated to add completion markers for implemented requirements
- [x] The default prompt template should encourage self-prioritisation of the task to work on and identification of any requirements that are already met
- [x] The default prompt must make it absolutely clear that one and only one requirement is to be implemented at a time
- [x] Add feature to allow a project specific `CONVENTIONS.md` containing project specific invariants (seeded by `--init`)
- [x] Change `RULES.md` to `CONVENTIONS.md` (seed conventions if not already present; docs refer to conventions rather than rules)
- [x] The initial seed content for `CONVENTIONS.md` should come from a `templates/CONVENTIONS.md` file in our project that is injected via go's embed mechanism
- [x] **PRIORITY** Drink our own champagne: in the root of this repo should be the recommended `PROMPT.md` and we should embed it in the binary using go:embed

### Release / docs
- [x] Add a github actions workflow to build the binary for common targets; osx/arm64, osx/x86_64, linux on common architectures, windows. ideally use go's cross-compilation rather than running on multiple build agents and have this run as a release job whenever a tag is created
- [x] Fix the release workflow so that it uploads all the binaries not just the linux one
- [x] Prepare the repo for a v0.0.1 release
- [x] Make a v0.0.1 release pushing the tag
- [x] Update the readme to document all the command line options

### Maintenance
- [x] Tidy up this spec so that the outdated specifications reflect current design.
