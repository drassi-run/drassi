<!--
SPDX-FileCopyrightText: (c) 2024 The Drassi Authors

SPDX-License-Identifier: Apache-2.0
-->

# Agent Instructions

> [!NOTE]
> This file provides context for AI agents working on the `drassi` project.

## 1. Project Overview & Architecture
Drassi (`/δράση/`) is a secure, sandboxed runner for GitHub Actions and other compatibles, supporting
multiple Git services (GitHub, Gitea) and offers strong isolation for job execution.

- **Go Version & Workspace**: Go 1.27 multi-module workspace (`go.work`):
  - `core/`: Core engine, executor, sandbox providers, and shared utilities (`drassi.run/core`)
  - `runner-gha/`: GitHub Actions runner daemon and CLI (`drassi.run/gha-runner`)
  - `runner-gitea/`: Gitea Actions runner daemon and CLI (`drassi.run/gitea-runner`)
- **Key Frameworks**:
  - Dependency Injection: `go.uber.org/dig` (configured in `core/wire/` and `runner-*/wire/`)
  - CLI commands: `github.com/spf13/cobra` (`runner-*/cmd/`)

## 2. Essential Commands

- **Environment & Build Tags**:
  Must include required build tags:
  ```bash
  export CGO_ENABLED=0
  export GOFLAGS="-tags=containers_image_openpgp,exclude_graphdriver_btrfs"
  ```
- **Testing**:
  ```bash
  go list -m | xargs -I {} go test "{}/..."
  ```
  Or run within a specific module:
  ```bash
  (cd core && go test ./...)
  ```
- **Linting**:
  ```bash
  golangci-lint run --timeout=10m ./core/...
  ```
  (configured via `.golangci.yaml`)
- **Pre-commit Checks**:
  ```bash
  pre-commit run --all-files
  ```
- **Building Binaries**:
  ```bash
  go build -trimpath -ldflags "-s -w" ./runner-gha
  go build -trimpath -ldflags "-s -w" ./runner-gitea
  ```

## 3. Agent Boundaries & Guardrails

- **Always**:
  - Use CodeGraph (`codegraph explore "<query>"`) before using `grep` or `find` to navigate code symbols and call paths.
  - Include the SPDX license header (via `hawkeye format` command) on all newly created files.
  - Run tests and pre-commit checks before completing any task.
- **Ask First**:
  - Introducing new dependencies or modifying `go.work`.
  - Changing public contracts in `core/` that affect both `runner-gha` and `runner-gitea`.
  - Modifying default sandbox isolation behaviors or security configurations.
- **Never**:
  - Bypass sandbox isolation logic without explicit user instruction.
  - Remove license headers, disable linters, or suppress security warnings without approval.
  - Commit private keys, secrets, or temporary credentials.

# Development Guidelines

<!-- Copy from https://github.com/multica-ai/andrej-karpathy-skills/blob/main/skills/karpathy-guidelines/SKILL.md -->

This **Karpathy principles** is behavioral guidelines to reduce common LLM coding mistakes, derived from [his observations](https://x.com/karpathy/status/2015883857489522876) on LLM coding pitfalls. Use when writing, reviewing, or refactoring code to avoid overcomplication, make surgical changes, surface assumptions, and define verifiable success criteria.

**Tradeoff:** These guidelines bias toward caution over speed. For trivial tasks, use judgment.

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:

- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:

- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:

- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:

- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:

```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.
