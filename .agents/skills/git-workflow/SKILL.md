---
name: git-workflow
description: >-
  Use this skill when the user explicitly asks to run their git workflow on the current changes — for example "run my git workflow", "git-workflow this", "stage, commit, push and open a PR", or "ship these changes". It takes the working-tree changes through four steps in order: stage only the files related to the change (leaving unrelated changes untouched), commit with a ≤60-character subject prefixed by feat/refactor/docs/bugfix, push the branch to remote, and open a pull request using the repo's .github PR template (falling back to a sensible default body when none exists).
---

# Git Workflow

A skill for taking the current working-tree changes from unstaged to an open pull request, in four ordered steps:

1. **Stage** only the files that belong to the change.
2. **Commit** with a short, prefixed subject line.
3. **Push** the branch to remote.
4. **Open a PR** using the repository's `.github` PR template.

The discipline that makes this skill worth having is **scope**: it stages and commits exactly the change at hand — nothing unrelated rides along — and it produces a tidy, conventionally-prefixed commit and a PR that follows the repo's own template.

## When to use this skill

Use it when the user explicitly asks to run their git workflow on the current changes. Typical openings:

- "Run my git workflow."
- "git-workflow this."
- "Stage, commit, push, and open a PR."
- "Ship these changes."

Do **not** trigger it as a reflex after every edit. It is a deliberate "take this change all the way to a PR" action. If the user only asks to commit (no PR), do just the steps they asked for.

## Core principles

1. **Stage only what's related.** Never `git add -A` or `git add .`. Stage files explicitly. Unrelated changes stay unstaged and get called out, not swept in.
2. **One change per commit.** If the working tree contains several unrelated changes, surface that and ask which one to ship rather than bundling them.
3. **Tidy, prefixed commit.** Subject ≤ 60 characters **including the prefix**, prefixed with exactly one of `feat`, `refactor`, `docs`, `bugfix`, chosen from what the diff actually does.
4. **Never commit on the default branch.** If on `main`/`master`, create a feature branch first.
5. **Follow the repo's PR template.** Use `.github/PULL_REQUEST_TEMPLATE.md` when present; fall back to a clean default only when there's no template.

## The flow

### Step 1 — Stage only the related files

- Inspect the working tree: `git status` and `git diff` (and `git diff --staged` to see anything already staged).
- Identify which changed/untracked files belong to the change being shipped. Group by intent.
- Stage them **explicitly** by path: `git add <path> [<path> ...]`. Do not use `git add -A`, `git add .`, or `git add -u`.
- If there are unrelated changes in the tree, leave them unstaged and tell the user what you skipped and why.
- If the tree spans **multiple unrelated changes** and it's not obvious which one to ship, stop and ask the user which change this commit/PR is for.

### Step 2 — Commit

- Write a subject line that is **≤ 60 characters total, prefix included**.
- Prefix with exactly one of, based on what the diff does:
  - `feat` — a new feature / new capability.
  - `refactor` — behavior-preserving restructuring.
  - `docs` — documentation-only changes.
  - `bugfix` — a fix for incorrect behavior.
- Format: `feat: short imperative summary` (e.g. `bugfix: handle empty token in auth header`).
- If the summary won't fit in 60 chars, **tighten the wording** — don't truncate mid-word and don't drop the prefix.
- Optionally add a body paragraph explaining the *why* when it isn't obvious from the subject. Keep the subject and body separated by a blank line.
- **Do not add a `Co-Authored-By` trailer** or any other co-author/attribution line to the commit.
- Show the user the proposed message, then commit the staged files.

### Step 3 — Push

- Determine the current branch. If it's the default branch (`main` or `master`), **stop and create a feature branch first** (e.g. `git checkout -b <descriptive-name>`), then commit/push there.
- Push to remote. On the branch's first push, set upstream: `git push -u origin <branch>`. Otherwise `git push`.
- If the push is rejected (remote has new commits), report it and let the user decide how to integrate — don't force-push.

### Step 4 — Open the pull request

- Look for the repo's PR template, in order:
  - `.github/PULL_REQUEST_TEMPLATE.md`
  - `.github/pull_request_template.md`
  - `.github/PULL_REQUEST_TEMPLATE/` (a directory of templates — ask the user which to use if several)
- **If a template exists**, fill in its sections from the diff (summary, motivation, test plan, checklists, etc.). Don't leave required sections blank; check off only what's actually true.
- **If no template exists**, fall back to a clean default body:

  ```markdown
  ## Summary
  <1–3 bullets describing what changed and why>

  ## Test plan
  <how this was/should be verified>
  ```

- Create the PR with `gh pr create`, giving it a title that matches the commit's intent and the filled-in body. Use a heredoc for the body to preserve formatting.
- Report the resulting **PR URL** back to the user.

## Handling edge cases

- **Nothing staged-worthy / clean tree.** If there are no changes to ship, say so and stop.
- **Mixed unrelated changes.** Stage only the related subset; explicitly list what you left out. If the intended change is ambiguous, ask before committing.
- **On the default branch.** Create a feature branch before committing/pushing; never commit straight to `main`/`master`.
- **Subject can't fit 60 chars.** Reword to fit; never truncate or drop the prefix.
- **Unsure which prefix applies.** Pick the one matching the dominant effect of the diff; if the change is genuinely two things (e.g. a feature + a fix), that's a sign it should be two commits.
- **No `gh` / not authenticated.** Report it and give the user the branch + a ready-to-paste PR title and body so they can open the PR manually.
- **Push rejected.** Surface the rejection; don't force-push. Let the user decide how to reconcile with remote.
- **PR already exists for the branch.** Report the existing PR URL instead of creating a duplicate.

## Tone

Act as a careful release hand: explicit about exactly which files were staged and which were deliberately left out, transparent about the commit message and PR body before they land, and quick to stop and ask when the change to ship is ambiguous. Show the user each artifact (staged set, commit message, PR URL) as it's produced.
