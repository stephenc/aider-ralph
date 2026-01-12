# Implementation of the Ralph Wiggum technique for Aider

## Project Overview

An iterative AI development methodology that repeatedly feeds aider a prompt until completion. Named after The Simpsons character, it embodies the philosophy of persistent iteration despite setbacks.

## Completion mechanism (how to mark progress)

This SPECS file uses a checkbox-based completion mechanism:

- For Markdown SPECS: each requirement MUST be a checkbox item using `- [ ]` (not done) and `- [x]` (done).
- If SPECS is ever converted to JSON: each requirement MUST be an object with a `completed` boolean field, and completion is tracked by setting `"completed": true`.

When a requirement is implemented, it should be marked complete by changing `- [ ]` to `- [x]` (or setting `"completed": true` in JSON).

## Requirements
- [x] Implement the equivalent of this bash loop: `while :; do aider --message "$(cat PROMPT.md)" --yes; done` but with safety rails
- [x] Richer and more nuanced termination conditions, i.e. use `<promise>COMPLETED</promise>` as that text is far less likely to be randomly output
- [x] Allow externalisation of the specs into a separate specs file so that a more standard prompt can be used
- [x] Update the instructions to allow and encourage forwarding notes to the next iteration
- [x] Use a sensible default prompt template if none provided. in fact all options should have sensible defaults and the specs file should be assumed. showing help if no specs file or specs file not specified on the cli. The prompt template will use PROMPT.md if that file is present and there is no override.
- [x] aider-ralph should always rescan files that it loads every iteration so that self-modification is permitted.
- [x] we should encourage the specs file to have a completion mechanism, if json formatted then each spec is an object with a completed boolean, if markdown then use a `- [ ] ` style list with the checkbox being checked after each one is implemented. Where necessary update the specs file to follow such a format.
- [x] the default prompt template should encourage self-prioritisation of the task to work on and identification of any requirements that are already met
- [x] the default prompt must make it absolutely clear that one and only one requirement is to be implemented at a time
- [x] the init mode should populate the current prompt template as PROMPT.md and create a seed SPECS.md as well as the `.ralph/notes.md` for notes. The default prompt template should indicate that `.ralph/notes.md` is append only and SPECS.md is only to be updated to add completion markers for implemented requirements
- [x] **PRIORITY** we should drink our own champaign. in the root of this repo should be the recommended PROMPT.md and we should embed it in the binary using go:embed
- [x] Fix bug where it kept on looping even after all the specs were marked as done. this should be a fix to the prompt so that we do not assume (other than the prompt) the format of the specs.
- [x] Add feature to allow a project specific RULES.md containing project specific invariants. this should be seeded by --init and include things like running the tests and linters and ensuring coverage is at least 75%
- [x] Add a github actions workflow to build the binary for common targets; osx/arm64, osx/x86_64, linux on common architectures, windows. ideally use go's cross-compilation rather than running on multiple build agents and have this run as a release job whenever a tag is created
- [x] prepare the repo for a v0.0.1 release
- [x] make a v0.0.1 release pushing the tag
- [x] fix the release workflow so that it uploads all the binaries not just the linux one
- [x] default max iterations to 30 rather than unbounded, let users specify 0 to select unbounded
