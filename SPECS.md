# Implementation of the Ralph Wiggum technique for Aider

## Project Overview

An iterative AI development methodology that repeatedly feeds aider a prompt until completion. Named after The Simpsons character, it embodies the philosophy of persistent iteration despite setbacks.

## Requirements
- [x] Implement the equivalent of this bash loop: `while :; do aider --message "$(cat PROMPT.md)" --yes; done` but with safety rails
- [x] Richer and more nuanced termination conditions, i.e. use `<promise>COMPLETED</promise>` as that text is far less likely to be randomly output
- [x] Allow externalisation of the specs into a separate specs file so that a more standard prompt can be used
- [x] Update the instructions to allow and encourage forwarding notes to the next iteration
- [ ] Use a sensible default prompt template if none provided. in fact all options should have sensible defaults and the specs file should be assumed. showing help if no specs file or specs file not specified on the cli. The prompt template will use PROMPT.md if that file is present and there is no override.
- [x] aider-ralph should always rescan files that it loads every iteration so that self-modification is permitted.
- [ ] we should encourage the specs file to have a completion mechanism, if json formatted then each spec is an object with a completed boolean, if markdown then use a `- [ ] ` style list with the checkbox being checked after each one is implemented. Where necessary update the specs file to follow such a format.
- [ ] the default prompt template should encourage self-prioritisation of the task to work on and identification of any requirements that are already met
- [ ] the default prompt must make it absolutely clear that one and only one requirement is to be implemented at a time
- [ ] the init mode should populate the current prompt template as PROMPT.md and create a seed SPECS.md as well as the `.ralph/notes.md` for notes. The default prompt template should indicate that `.ralph/notes.md` is append only and SPECS.md is only to be updated to add completion markers for implemented requirements
- [ ] **PRIORITY** we should drink our own champaign. in the root of this repo should be the recommended PROMPT.md and we should embed it in the binary using go:embed

