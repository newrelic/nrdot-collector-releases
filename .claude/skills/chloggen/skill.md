---
name: chloggen
description: |
  Use when adding a changelog entry to this repo — creating or editing files in
  `.chloggen/`, running `make chlog-new/-validate/-preview/-update`, or preparing
  a user-facing PR whose release notes should show up in CHANGELOG.md. Covers
  the draft-PR-first dance that lets `make chlog-new` auto-fill the PR number
  and change type.
---

# chloggen Workflow

This repo uses chloggen to generate CHANGELOG.md from per-PR YAML entries in `.chloggen/`. Every user-facing PR needs one entry.

## Commands

- `make chlog-new` — create a template entry named after the current branch, auto-filling `issues:` and `change_type` if a PR already exists.
- `make chlog-validate` — validate all pending entries. Fails if any entry has empty `issues:`.
- `make chlog-preview` — dry-run the CHANGELOG update.
- `make chlog-update` — consume entries and write to CHANGELOG.md (release only).

Change types allowed here: `feature`, `bug_fix`, `docs`. The wrapper in `scripts/chloggen-wrapper.sh` translates these to chloggen's built-ins — don't hand-edit it.

## Preferred workflow (draft PR first)

`make chlog-new` reads the PR for the current branch to prefill `issues:` and infer `change_type` from the PR title's conventional-commit prefix (`feat` → feature, `fix` → bug_fix, `docs` → docs). To use that, create the PR as a draft first so nothing publishes before the entry lands.

1. Make code changes on your branch, commit, push.
2. `gh pr create --draft --title "feat: ..." --body "..."`
3. `make chlog-new` — writes `.chloggen/<branch>.yaml` with `issues: [N]` and `change_type` prefilled.
4. Edit the file: fill in `note:` (one-line description). Add `subtext:` only if extra detail is warranted.
5. `make chlog-validate` — should pass now.
6. Commit the entry, push.
7. `gh pr ready <PR>` — promote out of draft.

## If you already opened a non-draft PR

`make chlog-new` still works; run it and it will fill in the PR number. Or hand-edit the template and set `issues: [N]` yourself. Commit as a follow-up.

## Gotchas

- `chlog-validate` errors with `specify one or more issues #'s` when `issues:` is empty — that means no PR existed when `make chlog-new` ran. Fill it in and re-run.
- Entry filename is `<branch-name>.yaml` (slashes replaced with underscores). Don't rename it.
- `.chloggen/TEMPLATE.yaml` and `.chloggen/config.yaml` are not entries; the wrapper skips them.
