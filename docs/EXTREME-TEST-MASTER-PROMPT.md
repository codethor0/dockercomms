# Master prompt: extreme testing and surgical bug-hunt (XT-1)

**Version: XT-1** — independent of the operator / GA runbook (e.g. v11). Use **v11** for closure evidence; use **this** file for adversarial QA.

Repository: local clone of `github.com/codethor0/dockercomms` (or fork). Paste the block below into a **testing-only** Cursor chat (or subagent).

---

## Paste block (for Cursor)

```
MASTER PROMPT: Extreme end-to-end testing and surgical bug-hunt — dockercomms
Repository: local clone of `codethor0/dockercomms` (or fork). Version: XT-1 (independent of operator runbook v11).
MISSION
Systematically stress the application to **discover** defects (correctness, security, UX, docs mismatch) and **recommend or apply surgical fixes** without broad refactors. Prefer **evidence** over intuition. This session is **not** a GA/closure run unless the user also pastes the v11 prompt.
SCOPE — “EXTENSIVE / EXTREME” MEANS
1. **All layers** where bugs hide: **unit** (`go test`), **integration** (tags, GHCR if creds), **container E2E** (`./scripts/docker-e2e.sh`), **local registry** (`docs/repro.md`), **CLI** contracts (`--help`, exit codes, stderr shape).
2. **Adversarial** and **edge**: invalid flags, empty inputs, huge path components, odd filenames, **wrong** digests, **unsigned** vs **signed** combinations, **two clients** to same repo (ordering), **resource** limits (large file if safe), **killed** registry mid-op (chaos, **if** you can do it safely).
3. **Contract / oracle**: `README` **## Exit Codes**, `SPEC.md` / `ARCH.md`, `docs/FINAL-RELEASE-GO-NO-GO.md` product checks — any mismatch = **bug** (code) or **doc debt** (docs), labeled explicitly.
4. **Concurrency & race**: `go test -race ./...` mandatory each session; if you add a repro for a race, **stress** the minimal code path.
5. **Fuzz (Go)**: for **pure** parsers/validators (paths, digests, filenames) use `go test -fuzz=...` **where** the repo already has fuzz tests or you add a **small** `fuzz_test.go` in the **relevant** package with **user authorization**; **do not** add fuzz in unrelated packages.
6. **Matrix** (if time): host **darwin**; optional `linux-distro-matrix.sh` **or** `docker-e2e.sh full` when `GH_PAT` exists — do **not** claim GHCR E2E without creds.
7. **Property / metamorphic** (lightweight): e.g. **round-trip** send→recv with **same** bytes; **idempotent** ack; **verify** of known-good artifact **succeeds**; document **1–2** invariants in the **Bug register**.
NON-NEGOTIABLE (LLM + HUMAN)
- **Never** claim a test “passed” without **verbatim** (or tailed) **command output** in the report, or a **citable** CI log. **No** fabricating exit codes.
- **Never** “fix” without: **(a)** a **repro** command, **(b)** **expected vs actual**, **(c)** **tests** updated or a **regression** test **if** the user authorized code changes in this session.
- **Surgical** diffs: **touch only** files required for the bug; **no** style-only churn; **no** feature creep.
- **Secrets:** never log `GH_PAT` or file contents. **I7 isolation**: unique OCI **repo** / `isolation_id` for any payload proof; no **P1** from **shared** inbox without isolation retry (per `docs/RELEASE-RUNBOOK.md` if present).
- **Bash** for **NEG** / `set +e` / `E1=$?` **drivers**; note if **zsh** default shell.
RECON (first 30 min of work)
- `git fetch` → **SHA**; `git status` — if dirty, list paths and decide: commit, stash, or **document** B-PROC-1.
- `ls scripts/`, `ls docs/`; read `README`, `docs/repro.md`, **exit code** section, `make help` or `Makefile`.
- Inventory **oracles**: table of **exit code → meaning** from **docs**; note gaps (“DOC-DEBT: exit X undocumented”).
- List **subcommands** and **critical flags** (send/recv/verify/ack/version) from **--help**.
PHASED EXECUTION (do in order; **stop** on first **unexpected** product failure in a **phase** until triaged, unless the user said “run all phases and log failures only”)
**P1 — Static and unit**
- `gofmt -l` (if non-empty, either format with permission or file as **style** bug); `go vet ./...`; `go test ./...`; `go test -race ./...`; `golangci-lint run ./...` (or project standard); `make coverage-gate`.
**P2 — Script preflight (no full GHCR)**
- `./scripts/run-integration.sh --check` ; `./scripts/login-and-run-integration.sh --check` ; `./scripts/docker-e2e.sh` **gates**; after gates, **`make clean && make build`** on **host** (rebuild if Linux binary overwrote).
**P3 — Local registry E2E (isolated)**
- Follow `docs/repro.md`: **one** **happy** round-trip per **isolation** key; **two** **negatives** (e.g. wrong digest, not-found) with **Expected per doc**; **two** **flake** iterations on **isolated** repos. **Bash** driver. Record **triage** PRODUCT | B-PROC-1 | DOC-DEBT | B-UX-HOST.
**P4 — CLI surface**
- All **--help**; **unknown** flags; **missing** required args; **version** output. Compare to **doc** and **Cobra** patterns.
**P5 — Combinations / matrix (if time)**
- Table: (sign: on/off) × (verify default / explicit) × (local vs no registry) for **at least** the paths the **doc** says are **supported**; any **combination** that **panics** or **wrong** exit = **P1** or **P0**.
**P6 — Stress / edge (opt-in)**
- **Very large** file: only with **spare** disk; **time cap**; **abandon** and note **N/A (resource)** if unwise.
- **Chaos** (e.g. stop `registry:2` mid-send): only with **isolation** and **clear** expected (retry? error code?).
**P7 — Auth / GHCR (opt-in)**
- Only with **`GH_PAT`** or `~/.dockercomms_gh_pat` on **host**; else **NOT RUN (auth)** with exact skip messages — do **not** treat as “pass”.
IF YOU FIX BUGS (only if user said “you may change code in this session”)
- **Minimal** patch; add or adjust **test** in the same package; run **`go test` + race** on **touched** packages and **full** `./...` if small; **no** unrelated formatting.
- One commit (or one PR) per **logical** fix unless user asked to batch.
DELIVERABLES (end of session, mandatory)
1. **Executive**: **#** tests run, **#** **unexpected** product findings (**P0**/**P1**), **#** **doc** issues, **#** **blocked** (auth, time).
2. **Bug register** (table): **ID | Severity (P0–P3) | Title | Repro (commands) | Expected | Actual | Oracle (doc ref) | Suspected file area | Suggested fix size (S/M/L)**. **Empty** unexpected **PRODUCT** = “no new unexpected failures in *this* matrix” — **not** “no bugs in universe.”
3. **NOT RUN** table for anything skipped.
4. **Gaps** for a **next** pass: **fuzz** targets, **long** chaos, **multi-Go-version** matrix, `linux-distro-matrix.sh`, **Scorecard** — one line each.
BANNED
- Weakening **tests** or **linters** to get green. Global “no bugs.” Large refactors. Committing **secrets** or **prompts** **unless** the user explicitly allows **this** file in the repo. Emojis in the report (optional: user style).
END XT-1
```

---

## How to use in Cursor

- Open a **new** chat (or a sub agent) so context is testing-only.
- Paste the **Part B** block; add one line, for example:
  - `Clone: /path/to/dockercomms`
  - `You may modify code: yes` or `no`
  - `Time budget: e.g. 2h`
- Keep using **v11** in **separate** chats for GA / NOT RUN evidence if you do not want bug-hunt and closure mixed.

## GitHub

This file is the canonical copy of the prompt on disk. Nothing “syncs” to GitHub until you **commit** and **push**.

## Related

- **`docs/TESTING-DOCTRINE.md`** — Part A (short project principles).
- **`docs/EXTREME-TEST-MASTER-PROMPT-XT2.md`** (XT-2) — live Docker + execution evidence; anti-summary opening; read after XT-1.
- **`docs/EXTREME-TEST-MASTER-PROMPT-XT3.md`** (XT-3) — change-aware matrix, PR checks, optional fix loop; after XT-2.
- **`docs/EXTREME-TEST-MASTER-PROMPT-XT4.md`** (XT-4) — evidence tables, triage, honesty line.
- **`docs/EXTREME-TEST-MASTER-PROMPT-XT5.md`** (XT-5) — GATE 0.1 Docker first, `STATIC_ONLY`, `DOCKER_REQUIRED`, `ON_MAIN` / `MERGE_SHA`.
- **`docs/EXTREME-TEST-MASTER-PROMPT-XT6.md`** (XT-6) — mirror `ci.yml` steps locally; CodeQL note; optional `RC_BUNDLE`, `act`.
- **`docs/RELEASE-RUNBOOK.md`** — operator closure, GA, NOT RUN.
