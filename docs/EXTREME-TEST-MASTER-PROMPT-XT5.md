# Master prompt: XT-5 — Preflight gate + static-only + main@merge + full matrix

**Version: XT-5** — **Extends:** [XT-1](EXTREME-TEST-MASTER-PROMPT.md) (doctrine), [XT-2](EXTREME-TEST-MASTER-PROMPT-XT2.md) (GATE 0, matrix), [XT-3](EXTREME-TEST-MASTER-PROMPT-XT3.md) (what changed), [XT-4](EXTREME-TEST-MASTER-PROMPT-XT4.md) (agent handoff, tables). **Supersedes:** nothing in v11 / [RELEASE-RUNBOOK.md](RELEASE-RUNBOOK.md). **Does not** replace GA/§A/§B.

## Why XT-5 exists

[XT-4](EXTREME-TEST-MASTER-PROMPT-XT4.md) showed: a long **Phase 1** can succeed while **Docker** is down, and **Phases 2–6** are **BLOCKED** — **correct** but **wasteful** if the goal was **end-to-end** that day. **XT-5** **reorders** checks: **verify Docker (or opt out) first**, then run the right **depth** of tests.

---

## Ladder

| Doc | Role |
|-----|------|
| v11 + [RELEASE-RUNBOOK.md](RELEASE-RUNBOOK.md) | GA, §A/§B, NOT RUN |
| [XT-1](EXTREME-TEST-MASTER-PROMPT.md) | Principles + adversarial plan |
| [XT-2](EXTREME-TEST-MASTER-PROMPT-XT2.md) | Live Docker, anti-summary, A–E |
| [XT-3](EXTREME-TEST-MASTER-PROMPT-XT3.md) | Change-aware, PR, optional fix loop |
| [XT-4](EXTREME-TEST-MASTER-PROMPT-XT4.md) | Evidence table, triage, honesty line |
| **XT-5** (this file) | **GATE 0.1** Docker first · **STATIC_ONLY** · **DOCKER_REQUIRED** · **ON_MAIN** / **MERGE_SHA** |

---

## Paste block (for Cursor) — everything in the fence

```
MASTER PROMPT: dockercomms — XT-5 “Preflight gate + static-only + main@merge + full matrix”
Extends: XT-1 (doctrine), XT-2 (GATE 0, matrix), XT-3 (what changed), XT-4 (agent handoff, tables).
Supersedes: nothing in v11 / RELEASE-RUNBOOK. Does not replace GA/§A/§B.

WHY XT-5 EXISTS
XT-4 showed: a long **Phase 1** can succeed while **Docker** is down, and **Phases 2–6** are **BLOCKED** — that is **correct** but **wastes** time if the user’s goal was **end-to-end** that day. XT-5 **reorders** checks: **verify Docker (or opt out) first**, then run the right **depth** of tests.

USER FLAGS (first message; defaults in parentheses)
- `CLONE=/path/to/dockercomms`  (required)
- `ON_MAIN=0|1`  (0 default) — if 1: `git fetch` and make **3-way** `HEAD` = `origin/main` = `gh` main **or** stop with **DRIFT**
- `MERGE_SHA=abc12`  (optional) — if set, `git checkout` that commit in a **detached** or **branches/verify** and state it in the header; else ignore
- `STATIC_ONLY=0|1`  (0) — if 1: run **only** Phase 1 + report; **no** docker-dependent scripts (treat as **intentional**)
- `DOCKER_REQUIRED=0|1`  (0 default) — if 1: if `docker info` **fails**, **abort entire session** with **one-page** **BLOCKED** report (no Phase 1) unless `STATIC_ONLY=1`
- Inherit from XT-4: `EXTREME`, `CONTINUE_ON_FAILURE`, `CODE_CHANGES_ALLOWED`, time budget, `PR_NUMBER`, `GH_PAT` present yes/no (never value)

REQUIRED READ (3–5 min, same as XT-4)
docs/TESTING-DOCTRINE.md, docs/EXTREME-TEST-MASTER-PROMPT.md (XT-1),
docs/EXTREME-TEST-MASTER-PROMPT-XT2.md, docs/EXTREME-TEST-MASTER-PROMPT-XT3.md, docs/EXTREME-TEST-MASTER-PROMPT-XT4.md,
docs/repro.md, README (exit codes), docs/RELEASE-RUNBOOK.md (context for GA, not a substitute for v11)

OPENING (ANTI-SUMMARY) — first lines MUST be:
`date -u` · `cd "$CLONE" && pwd` · `git status -sb` · `git rev-parse HEAD` · `go version` ·
**THEN** immediately **GATE 0.1 — DOCKER 10-SECOND CHECK**:
- `command -v docker; docker info` (capture exit code; if **non-zero**: print **first 20 lines** stderr/stdout; **one line** suggestion: e.g. start **Docker Desktop**, or **colima start**, or **remote** context)
- If **DOCKER_REQUIRED=1** and docker info **fails**:
  - Output only the **### XT-5 BLOCKED** section (template below) + honesty line, **and STOP** (unless `STATIC_ONLY=1` in same message — then continue with Phase 1 only)
- If **docker info** **succeeds**, print **GATE 0.1: OK** in one line and **continue** to preflight / phases.

GATE 0.2 — BRANCH / MAIN (if `ON_MAIN=1`)
- `git fetch origin`
- Compare `HEAD`, `origin/main`, `gh api` main — if **mismatch** and you are not intentionally on a topic branch, **state NO** 3-way match and **do not** claim “published main” results.

PHASES (INHERIT FROM XT-4, BUT ORDERED)
- **0** XT-3: what changed / targeted `go test -race` on touched packages
- **1** Static + unit (gofmt, vet, go test, race, lint, coverage) — **skip** if `DOCKER_REQUIRED=1` and docker **failed** and `STATIC_ONLY=0` (aborted at gate) — if `STATIC_ONLY=1` **always** run
- **2** `--check` scripts — only if `docker info` **ok**
- **3** `docker-e2e.sh` **gates** → `make clean && make build` + `file` (only if docker **ok**)
- **4** local registry **repro** + NEGs + I8 (bash) — only if docker **ok**
- **5** optional PAT + Section B + docker-e2e i|clf — only if PAT present
- **6** optional `linux-distro-matrix.sh` if `EXTREME=1` and docker **ok**

REPORT
Same sections as **XT-4**, with these **additions** at top of **### XT-5 RESULT** (or **XT-4/5** if you want one template):
- `GATE 0.1: Docker OK|FAIL`
- `ON_MAIN: yes|no` + 3-way match: yes|no|n/a
- `STATIC_ONLY: yes|no`
- `DOCKER_REQUIRED: yes|no`

### XT-5 BLOCKED (template when DOCKER_REQUIRED=1 and docker fails)
- Reason: `docker info` non-zero: <first line of error>
- Suggested user action: start Docker / fix socket / set DOCKER_HOST if remote
- Phase 1 **not** run: **yes|no** (if user did not set STATIC_ONLY, say **skipped**; if **STATIC_ONLY=1**, list **go test** outcome)

HONESTY LINE (unchanged in spirit)
Same as XT-4; never “no bugs globally.”

BANS
- Running **Phase 3+** with docker **down** and pretending you did E2E
- Claiming **3-way** **main** match on a **topic** branch when `ON_MAIN=1` was not set
- P1 on **script** “Docker not running” failures without repro on a **good** daemon

END XT-5
```

### XT-5 BLOCKED (copy when rendering reports)

- **Reason:** `docker info` non-zero: (first line of error)
- **Suggested user action:** start Docker Desktop / fix socket / `colima start` / set `DOCKER_HOST` if remote
- **Phase 1 not run:** yes | no (if `STATIC_ONLY=1`, list `go test` outcome)

---

## How to use in Cursor (one line under the block)

`CLONE: /path/to/dockercomms` · `DOCKER_REQUIRED: 0` · `STATIC_ONLY: 0` · `ON_MAIN: 0` · `CODE_CHANGES_ALLOWED: no`

---

## See also

- [EXTREME-TEST-MASTER-PROMPT-XT4.md](EXTREME-TEST-MASTER-PROMPT-XT4.md) (XT-4) · [TESTING-DOCTRINE.md](TESTING-DOCTRINE.md) · [RELEASE-RUNBOOK.md](RELEASE-RUNBOOK.md) (v11)
