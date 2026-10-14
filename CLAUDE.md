# CLAUDE.md

@AGENTS.md

## Claude-specific notes

- AGENTS.md is the source of truth for this repo: Part A is how I work on Go projects, Part B
  the rules of this one. When you deviate from a rule, say which one and why.
- New requirements go through the superpowers skills in the order given in AGENTS.md section B5:
  brainstorming (with its approval gates), writing-plans, executing-plans with
  test-driven-development inside each task, verification-before-completion, then `/simplify`
  plus the `code-simplifier` agent. Do not skip the brainstorming step because a request looks
  small; classify it out loud and let the human override.
- "Tests pass" means `make test`, `make lint` and `go vet ./...` exited 0 in this session;
  after repository changes also `make seed && make test-integration` with the compose database
  up; after Docker or compose changes also `make docker-up-app` and a smoke request through the
  container. Report the real output when something fails.
- The human owns git: never `git commit`, `push`, `init`, `checkout`, `switch`, `branch`,
  `merge`, `rebase`, `reset` or `stash`. At a commit point, stop, list the files to stage and
  propose `feat: [TASK-NNN] - brief description`, then wait. Put the same rule in every
  subagent prompt and check `git status` and `git branch --show-current` afterwards.
- Subagents used here: read-only research agents for library comparisons (they return a table
  and a recommendation), `/simplify` review agents and the `code-simplifier` agent for cleanup
  passes. None of them edits files when another process is editing the same files; they
  report diffs instead.
- Prefer one layer per diff; when a constructor signature changes, update
  `cmd/server/main.go` in the same change.
- Scratch files go in the session scratchpad, not in the repo.
