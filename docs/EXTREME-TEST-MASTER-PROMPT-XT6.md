# Master prompt: XT-6 — CI workflow parity + RC bundle (optional)

**Version: XT-6** — **Extends:** [XT-1](EXTREME-TEST-MASTER-PROMPT.md) through [XT-5](EXTREME-TEST-MASTER-PROMPT-XT5.md). **Supersedes:** nothing in v11 / [RELEASE-RUNBOOK.md](RELEASE-RUNBOOK.md). **Does not** replace §A/§B/GA.

## Why XT-6 exists

Local `go test ./...` can be **green** while **CI** fails (different **Go** version, **make** targets, missing **step**, **OS**). **XT-6** forces the agent to **read** `.github/workflows/*.yml` and **execute the same shell steps** the workflow uses (as closely as the host allows), then run **XT-5** on top. Optional **RC** = one **numbered** report block for a **PR** or **release** candidate—**not** GA.

---

## Ladder

| Doc | Role |
|-----|------|
| v11 + [RELEASE-RUNBOOK.md](RELEASE-RUNBOOK.md) | GA, §A/§B |
| [XT-1](EXTREME-TEST-MASTER-PROMPT.md) … [XT-5](EXTREME-TEST-MASTER-PROMPT-XT5.md) | See each file |
| **XT-6** (this file) | **CI mirror** + [XT-5](EXTREME-TEST-MASTER-PROMPT-XT5.md) inherit · optional **RC_BUNDLE** · optional **`act`** |

---

## Paste block (for Cursor) — everything in the fence

```
MASTER PROMPT: dockercomms — XT-6 “CI workflow parity + RC bundle (optional)”
Extends: XT-1 (doctrine) … XT-5 (Docker preflight, ON_MAIN, STATIC_ONLY).
Supersedes: nothing in v11 / RELEASE-RUNBOOK. Does not replace §A/§B/GA.

WHY XT-6 EXISTS
Local `go test ./...` can be **green** while **CI** fails (different **Go** version, different **make** targets, missing **step**, **OS**). XT-6 forces the agent to **read** `.github/workflows/*.yml` and **execute the same commands** the workflow uses (as closely as the host allows), then run **XT-5** on top. Optional **RC** = one **numbered** report block you can attach to a **PR** or **release** candidate without calling it GA.

USER FLAGS (first message; merge with XT-4/5)
- `CLONE=/path/to/dockercomms`  (**required**)
- `CI_PARITY=1`  (default **1** for XT-6) — if **0**, skip §A workflow mirror
- Inherit: `ON_MAIN`, `MERGE_SHA`, `STATIC_ONLY`, `DOCKER_REQUIRED`, `EXTREME`, `CONTINUE_ON_FAILURE`, `CODE_CHANGES_ALLOWED`, `PR_NUMBER`, time budget, `GH_PAT` present yes/no
- `RC_BUNDLE=0|1`  (0 default) — if **1**, add **§ RC CANDIDATE** block (below) for human paste into PR/release notes
- `ACT=0|1`  (0) — if **1** and `act` is on `PATH`, try `act -l` and **optionally** `act push` **or** one job with `--dryrun` if supported; if `act` missing, **NOT RUN (dep)** one line (do not fake `act` results)

REQUIRED READ
docs/TESTING-DOCTRINE.md
docs/EXTREME-TEST-MASTER-PROMPT.md (XT-1) … docs/EXTREME-TEST-MASTER-PROMPT-XT5.md (XT-5) (all as in XT-5)
`.github/workflows/ci.yml`  (**mandatory** for this prompt)
`.github/workflows/codeql.yml`  (read; **do not** “run” CodeQL locally — note **“CodeQL: CI only; local NOT RUN (tooling)”** unless user has `codeql` and explicitly allows)
`Makefile`  (targets referenced from CI)
`README`  (exit codes)

OPENING (ANTI-SUMMARY)
First lines: `date -u` · `pwd` · `git status -sb` · `go version` · `git rev-parse HEAD` · **then** **GATE 0.1** from **XT-5** (`docker info`).

§A — CI WORKFLOW MIRROR (if `CI_PARITY=1`)
1) Open **`.github/workflows/ci.yml`**. In your report, paste a **short** bullet list: **triggers** (on: push, pull_request…), **jobs** names, **runner** (ubuntu-*), **Go** version (setup-go or container image), and **ordered** run steps (install deps, `make` targets, `go test` flags, **coverage** gate, **lint**).  
2) For **each** step that is a **shell** command you can run on **this** host, **run it** in the **same order** (adapt paths only; **do not** change flags unless Mac vs Linux **forces** it — if forced, mark **CI_PARITY: PARTIAL** and explain in **one** line).  
3) If the workflow uses **a matrix** (os, go), run the **go** and **test** step for **`go` on PATH** and note: **“CI runs matrix; local = single column.”**  
4) If a step is **only** on **ubuntu** and cannot run (e.g. `apt-get` only), mark **NOT RUN (host os)**.  
5) If **coverage** or **lint** in CI use **different** version than local, run local anyway and add **one** line: **delta vs CI** if **visible** in workflow file.  
6) Conclude **§A** with: **`CI_PARITY: STRONG | PARTIAL | NOT RUN (reason)`**.

§B — CODEQL
- Default: **NOT RUN (local): CodeQL executed on GitHub for push/PR; local scan requires CodeQL bundle**. One line: link to **CodeQL** workflow in repo or `gh run` last **Analyze** if `gh` works.

§C — INHERIT XT-5
Run **GATE 0.2** (`ON_MAIN` if set), then **phases 0–6** from **XT-5/4** as applicable (**Docker** required paths, **PAT** paths, **EXTREME** for matrix script).

REPORT (append to XT-4/5 structure)
- Top of **### XT-6 RESULT**:
  - `CI_PARITY: STRONG|PARTIAL|NOT_RUN`
  - `Workflows read: ci.yml (yes), codeql.yml (yes)`
  - `Go version: workflow says …, local: …`

§ RC CANDIDATE (only if `RC_BUNDLE=1` — for humans, not GA)
Paste **exact** block in the final output:

  ### Release candidate / PR test bundle (not GA; not v11)
  - **Date (UTC)**: …
  - **Local HEAD** / **Branch**: …
  - **3-way `main` match (if run with ON_MAIN=1)**: yes|no|n/a
  - **CI mirror**: STRONG|PARTIAL — 1 line
  - **Go tests (local)**: pass|fail (scope)
  - **Docker (GATE 0.1)**: ok|fail
  - **Container gates + host rebuild (if run)**: ok|skip
  - **Local registry E2E (if run)**: ok|skip
  - **GHCR chain (if run)**: ok|not run (auth)
  - **Honesty**: "RC bundle is evidence of testing effort; it does not replace v11 operator GA/§A/§B." 

BANS
- Claiming **“same as CI”** without **opening** `ci.yml` and **listing** the steps you ran.  
- Skipping **XT-5** **GATE 0.1** when the workflow uses **Docker** and you plan **docker-e2e** in the same session.  
- Merging this block with **GREEN GA** in **v11**; **v11** remains separate.

HONESTY LINE
Same as XT-4/5, plus: **"CI mirror is best-effort; runner OS and Go image may differ from darwin/ local go."**

END XT-6
```

---

## How to use in Cursor (one line)

`CLONE: /path/to/dockercomms` · `CI_PARITY: 1` · `RC_BUNDLE: 0` · `ACT: 0` · `DOCKER_REQUIRED: 0` · `STATIC_ONLY: 0`

---

## This repo’s `ci.yml` (reference — do not treat as the paste block; always re-read the file)

Current workflow (check out-of-date): job **`test`**, `runs-on: ubuntu-latest`, `go-version-file: go.mod`, steps in order: **gofmt** (fail if any output) → **go vet** → **go test** → **go test -race** → **golangci-lint** (install v2.10.1) → **make coverage-gate**.

---

## See also

- [EXTREME-TEST-MASTER-PROMPT-XT5.md](EXTREME-TEST-MASTER-PROMPT-XT5.md) · [TESTING-DOCTRINE.md](TESTING-DOCTRINE.md) · [RELEASE-RUNBOOK.md](RELEASE-RUNBOOK.md)
