# Master prompt: XT-3 — Change-aware live matrix + optional fix loop

**Version: XT-3** — next step after [XT-2](EXTREME-TEST-MASTER-PROMPT-XT2.md). **Supersedes:** nothing in v11 / [RELEASE-RUNBOOK.md](RELEASE-RUNBOOK.md). **Extends:** [XT-1](EXTREME-TEST-MASTER-PROMPT.md) (doctrine, adversarial plan), [XT-2](EXTREME-TEST-MASTER-PROMPT-XT2.md) (live Docker, anti-summary, GATE 0, matrix A–E).

**Use when:** you have a **PR**, a **feature branch**, or you just **merged** to `main` and need a **proof** run + **targeted** tests on **touched** code.

**GA / §A / §B** — still **[RELEASE-RUNBOOK.md](RELEASE-RUNBOOK.md) + v11** only; **XT-3** is **QA / merge confidence**, not §A/§B proof.

---

## Ladder (prompts)

| Prompt / doc | Role |
|--------------|------|
| **v11** + [RELEASE-RUNBOOK.md](RELEASE-RUNBOOK.md) | Operator / GA / §A/§B / NOT RUN |
| [XT-1](EXTREME-TEST-MASTER-PROMPT.md) | Principles + full adversarial plan |
| [XT-2](EXTREME-TEST-MASTER-PROMPT-XT2.md) | Live Docker + no-summary + A–E matrix |
| **XT-3** (this file) | What changed first + PR/merge confidence + optional fix loop |

---

## Paste block (for Cursor) — everything in the fence

```
MASTER PROMPT: dockercomms — XT-3 “Change-aware live matrix + optional fix loop”
Supersedes: nothing in v11 / RELEASE-RUNBOOK. Extends: XT-1 (doctrine), XT-2 (live Docker + anti-summary).
Use when: you have a **PR**, a **feature branch**, or you just **merged** to `main` and need a **proof** run + **targeted** tests on **touched** code.
MISSION
Run the **strongest** test pass that matches **what changed** (not only full `./...` every time), then **optionally** apply **surgical** fixes with **tests** if `CODE_CHANGES_ALLOWED=yes` in the user message. Output is **evidence-first**; **no** “no bugs in the universe” — only **no unexpected failures in the run matrix** for this session.
REQUIRED READ (same session, first 5 min)
- `docs/TESTING-DOCTRINE.md`
- `docs/EXTREME-TEST-MASTER-PROMPT.md` (XT-1)
- `docs/EXTREME-TEST-MASTER-PROMPT-XT2.md` (XT-2) — **GATE 0 Docker**, **anti-summary** opening, **host rebuild** after container gates
- `docs/RELEASE-RUNBOOK.md` — for **GA** you still use **v11** separately; XT-3 is **QA / merge confidence**, not §A/§B proof
ANTI-SUMMARY (same as XT-2, HARD)
First output lines = UTC, path, **3-way SHA** (if on `main` with `git fetch`) or **branch + merge-base** + first **command + output**. No plan wall before **docker info** or **go version**.
PRE-FLIGHT: WHAT CHANGED? (XT-3 NEW)
1. `git status -sb` and `git rev-parse HEAD` and `git branch --show-current`
2. If on **PR** or user gave `PR_NUMBER=n`: `gh pr view n` (or `gh pr diff n --name-only` if large) — list **touched** paths; **if** not available, `git log -1 --name-only` and `git diff origin/main...HEAD --name-only` (or `main`…HEAD)
3. **Targeted** test first (fast signal):
   - For each **changed** package under `pkg/`, `cmd/`, `internal/`: `go test -race ./path/...` 
   - If only **docs** / **.md** changed, say **“no Go diff; skip targeted go test; run full static + XT-2 matrix for smoke only”**
4. Then run **global** `go test ./...` and `go test -race ./...` and **lint** / **coverage** per project (same as XT-2 **phase A**)
THEN: XT-2 CORE (INHERIT)
- **GATE 0** `docker info`  
- If Docker OK: `./scripts/docker-e2e.sh` **gates** → **`make clean && make build` + `file` one line**  
- **B** and **C** and **D** and **E** as in **XT-2** (`--check` scripts, local `repro`+NEG+flake with **bash**, integration skip or full chain with PAT)
PR-SPECIFIC CHECKS (XT-3)
- If `PR_NUMBER` in message: `gh pr checks <n> --watch` (or non-watch) until **conclusion** or timeout — **attach** the **failing** check name + last lines if any  
- If PR is **from fork**, note **if** **secrets** workflows are **not** run — do **not** expect full Actions parity locally
OPTIONAL: FIX LOOP (only if `CODE_CHANGES_ALLOWED=yes` in first message)
For each **unexpected** PRODUCT row (5 fields per TESTING-DOCTRINE):
1. **Minimal** code change; **add** or **tighten** a **test** in the same package.  
2. `go test -race ./<pkg>/...` then **`go test -race ./...` if affordable**  
3. **One** commit message style: `fix(<area>): <short>`
If `CODE_CHANGES_ALLOWED=no`: **file** findings in a **bug register** table only — **no** file edits.
REPORT (END)
- **Change summary**: 1 line + **touched** paths (from pre-flight)
- **SHA** / **branch** / **PR** link (if any)
- **Table**: same structure as **XT-2** + column **Targeted?** (yes if package had Go changes)
- **Merge recommendation** (if PR): `READY` | `NEEDS FIX` | `NEEDS HUMAN` — one line each with **2** evidence bullets, **not** a lecture
- If full matrix clean: **suggested** **CHANGELOG** line: `Tested: <short-sha> (XT-3: targeted + full matrix, Docker: yes/no, …)` — **optional** user paste
USER MESSAGE (FIRST) SHOULD INCLUDE
- `Clone: /path/to/dockercomms`
- `PR_NUMBER: <n> | 0` (0 = not a PR run)
- `CODE_CHANGES_ALLOWED: yes|no` + if yes **scope** in one line
- `Time budget: e.g. 90m`
- `GH_PAT on host: yes|no` (value never in chat)
- `E2E_LINUX_MATRIX: 0|1` · `CONTINUE_ON_FAILURE: yes|no` (if yes, log all failures, still stop on P0)
- Inherit: XT-2 flags as needed
BANNED
- Skipping **targeted** `go test` on **changed** packages when Go files changed
- Merging “GA done” with XT-3 — **v11** only for that
- Summary-first opening; secrets in log; emojis in report (project default: off)
END XT-3
```

---

## How to use in Cursor (one line to paste under the block)

`Clone: /path/to/dockercomms` · `PR_NUMBER: 0` · `CODE_CHANGES_ALLOWED: no` · `Time budget: 2h` · `GH_PAT on host: no`

---

## See also

- [TESTING-DOCTRINE.md](TESTING-DOCTRINE.md)
- [EXTREME-TEST-MASTER-PROMPT.md](EXTREME-TEST-MASTER-PROMPT.md) (XT-1)
- [EXTREME-TEST-MASTER-PROMPT-XT2.md](EXTREME-TEST-MASTER-PROMPT-XT2.md) (XT-2)
- [RELEASE-RUNBOOK.md](RELEASE-RUNBOOK.md) (v11 / GA)
